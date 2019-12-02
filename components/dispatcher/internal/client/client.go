package client

import (
	"bytes"
	"context"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/pkg/errors"
	"github.com/slok/goresilience"
	"github.com/slok/goresilience/retry"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"
)

const responseStatus = "Response.Status"

// Create a shared go http client with a timeout (timeout should be slightly larger than eventproxy timeout)
var client = &http.Client{
	Timeout: 30 * time.Second,
}

// Client represents anything that can dispatch an event
// to a downstream service
type Client interface {
	Dispatch(message []byte) error
}

// httpClient is an implementation of Client which dispatches via HTTP
// and provides exponential backoff capabilities
type httpClient struct {
	uri                  string
	exponentialBackoff   bool
	initialRetryInterval int64
	maxRetryTime         int64
}

func NewHttpClient(uri string, exponentialBackoff bool, initialRetryInterval int64, maxRetryTime int64) httpClient {
	return httpClient{uri: uri, exponentialBackoff: exponentialBackoff, initialRetryInterval: initialRetryInterval, maxRetryTime: maxRetryTime}
}

func (hc httpClient) Dispatch(message []byte) error {
	// Configure The Logger
	var logger *zap.Logger
	if log.Logger().Core().Enabled(zap.DebugLevel) {
		logger = log.Logger().With(zap.String("Message", string(message)), zap.String("uri", hc.uri))
	} else {
		logger = log.Logger().With(zap.String("uri", hc.uri))
	}

	runner := retry.New(retry.Config{DisableBackoff: hc.exponentialBackoff, Times: hc.calculateNumberOfRetries(), WaitBase: time.Millisecond * time.Duration(hc.initialRetryInterval)})

	err := sendKafkaConsumerMessageToSubscriber(runner, hc, message, logger)

	// Retries failed
	if err != nil {
		logger.Error("Failed to send after configured number of retries", zap.Error(err))
		return err
	}
	return nil
}

func sendKafkaConsumerMessageToSubscriber(runner goresilience.Runner, hc httpClient, message []byte, logger *zap.Logger) error {
	err := runner.Run(context.TODO(), func(_ context.Context) error {

		// Send The Kafka ConsumerMessage To Subscriber URI Via HTTP POST Request
		response, err := client.Post(hc.uri, "application/json", bytes.NewBuffer(message))
		if err != nil {
			logger.Warn("Failed to send message to subscriber service, retrying", zap.Error(err))
			return err
		}

		// Log The Response & Defer Closing Response Body
		if response != nil {

			defer func() {

				closeResponseBody(response, logger)

			}()

			// required to read the body (even though we don't need it) to avoid memory leaks when reusing http client:
			_, _ = io.Copy(ioutil.Discard, response.Body)

			err = logResponse(logger, response)
			if err != nil {
				return err
			}

		} else {
			logger.Warn("Received nil response when sending message to subscriber service, retrying")
			return errors.New("Received nil response when sending message to subscriber service")
		}

		return nil
	})
	return err
}

func closeResponseBody(response *http.Response, logger *zap.Logger) {
	if response.Body != nil {
		err := response.Body.Close()
		if err != nil {
			logger.Warn("Failed to close HTTP response body", zap.Error(err))
		}
	}
}

func logResponse(logger *zap.Logger, response *http.Response) error {
	if response.StatusCode >= 500 || response.StatusCode == 404 || response.StatusCode == 429 {
		logger.Warn("Failed to send message to subscriber service, retrying", zap.String(responseStatus, response.Status))
		return errors.New("Server returned a bad response code: " + strconv.Itoa(response.StatusCode))
	} else if response.StatusCode > 299 {
		logger.Warn("Failed to send message to subscriber service, not retrying", zap.String(responseStatus, response.Status))
	} else {
		logger.Info("Successfully sent message to subscriber service", zap.String(responseStatus, response.Status))
	}
	return nil
}

// Convert defined max retry time to the approximate number
// of retries, taking into account the exponential backoff algorithm
func (hc httpClient) calculateNumberOfRetries() int {
	return int(math.Round(math.Log2(float64(hc.maxRetryTime)/float64(hc.initialRetryInterval))) + 1)
}

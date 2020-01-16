package client

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/pkg/errors"
	"github.com/slok/goresilience/retry"
	"go.uber.org/zap"
	"math"
	"net/http"
	"strconv"
	"time"
)

// Create a shared go http client with a timeout
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// Client represents anything that can dispatch an event
// to a downstream service
type RetriableClient interface {
	Dispatch(message cloudevents.Event) error
}

// retriableCloudEventClient is a client implementation that interprets
// kafka messages as cloud events and utilizes the cloud event library
// and supports retries with exponential backoff
type retriableCloudEventClient struct {
	uri                  string
	exponentialBackoff   bool
	initialRetryInterval int64
	maxRetryTime         int64
	cloudEventClient     cloudevents.Client
}

var _ RetriableClient = &retriableCloudEventClient{}

func NewRetriableCloudEventClient(uri string, exponentialBackoff bool, initialRetryInterval int64, maxRetryTime int64) retriableCloudEventClient {
	transport, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(uri),
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV03),
	)
	transport.Client = httpClient

	if err != nil {
		panic("failed to create transport, " + err.Error())
	}

	ceClient, err := cloudevents.NewClient(transport)
	if err != nil {
		panic("unable to create cloudevent client: " + err.Error())
	}

	return retriableCloudEventClient{uri: uri, exponentialBackoff: exponentialBackoff, initialRetryInterval: initialRetryInterval, maxRetryTime: maxRetryTime, cloudEventClient: ceClient}
}

func (rcec retriableCloudEventClient) Dispatch(event cloudevents.Event) error {
	// Configure The Logger
	var logger *zap.Logger
	if log.Logger().Core().Enabled(zap.DebugLevel) {
		logger = log.Logger().With(zap.String("Event", event.String()), zap.String("uri", rcec.uri))
	} else {
		logger = log.Logger().With(zap.String("uri", rcec.uri))
	}

	runner := retry.New(retry.Config{DisableBackoff: rcec.exponentialBackoff, Times: rcec.calculateNumberOfRetries(), WaitBase: time.Millisecond * time.Duration(rcec.initialRetryInterval)})

	err := runner.Run(context.TODO(), func(_ context.Context) error {
		rctx, _, err := rcec.cloudEventClient.Send(context.Background(), event)
		if err != nil {
			transportContext := cloudevents.HTTPTransportContextFrom(rctx)
			return logResponse(logger, transportContext.StatusCode, err)
		}
		return nil
	})

	// Retries failed
	if err != nil {
		logger.Error("Failed to send after configured number of retries", zap.Error(err))
		return err
	}
	return nil
}

func logResponse(logger *zap.Logger, statusCode int, err error) error {
	if statusCode >= 500 || statusCode == 404 || statusCode == 429 {
		logger.Warn("Failed to send message to subscriber service, retrying", zap.Int("statusCode", statusCode))
		return errors.New("Server returned a bad response code: " + strconv.Itoa(statusCode))
	} else if statusCode > 299 {
		logger.Warn("Failed to send message to subscriber service, not retrying", zap.Int("statusCode", statusCode))
	} else if statusCode == 0 {
		return errors.Wrap(err, "Validation Error")
	} else {
		logger.Debug("Successfully sent message to subscriber service", zap.Int("statusCode", statusCode))
	}
	return nil
}

// Convert defined max retry time to the approximate number
// of retries, taking into account the exponential backoff algorithm
func (rcec retriableCloudEventClient) calculateNumberOfRetries() int {
	return int(math.Round(math.Log2(float64(rcec.maxRetryTime)/float64(rcec.initialRetryInterval))) + 1)
}

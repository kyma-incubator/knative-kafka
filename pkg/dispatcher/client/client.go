package client

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/kyma-incubator/knative-kafka/pkg/dispatcher/subscription"
	"github.com/pkg/errors"
	"github.com/slok/goresilience/retry"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/channel"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// 3 Digit Word Boundary HTTP Status Code Regular Expression
var HttpStatusCodeRegExp = regexp.MustCompile("(^|\\s)([12345]\\d{2})(\\s|$)")

// Client represents anything that can dispatch an event
// to a downstream service
type RetriableClient interface {
	Dispatch(message *cloudevents.Event, subscription subscription.Subscription) error
}

// retriableCloudEventClient is a client implementation that interprets
// kafka messages as cloud events and utilizes the cloud event library
// and supports retries with exponential backoff
type RetriableCloudEventClient struct {
	logger               *zap.Logger
	exponentialBackoff   bool
	initialRetryInterval int64
	maxRetryTime         int64
	messageDispatcher    channel.MessageDispatcher
}

var _ RetriableClient = &RetriableCloudEventClient{}

func NewRetriableCloudEventClient(logger *zap.Logger, exponentialBackoff bool, initialRetryInterval int64, maxRetryTime int64) RetriableCloudEventClient {

	return RetriableCloudEventClient{
		logger:               logger,
		exponentialBackoff:   exponentialBackoff,
		initialRetryInterval: initialRetryInterval,
		maxRetryTime:         maxRetryTime,
		messageDispatcher:    channel.NewMessageDispatcher(logger),
	}
}

// Dispatch A Cloud Event To The Specified Destination URL
func (c RetriableCloudEventClient) Dispatch(event *cloudevents.Event, subscription subscription.Subscription) error {

	// Configure The Logger
	logger := c.logger.With(zap.String("Subscriber URI", subscription.SubscriberURI.String()))
	if c.logger.Core().Enabled(zap.DebugLevel) {
		logger = c.logger.With(zap.String("Event", event.String()))
	}

	// Create A New Retry Runner With Configured Backoff Behavior
	retryRunner := retry.New(
		retry.Config{
			DisableBackoff: c.exponentialBackoff,
			Times:          c.calculateNumberOfRetries(),
			WaitBase:       time.Millisecond * time.Duration(c.initialRetryInterval),
		},
	)

	//
	// Convert The CloudEvent To A binding/Message
	//
	// TODO - It is potentially inefficient to take the KafkaMessage which we've already turned into a Cloud Event, only
	//        to re-convert it into a binding/Message.  The current implementation is based on CloudEvent Events, however,
	//        and until a "protocol" implementation for Confluent Kafka exists this is the simplest path forward.  Once
	//        such a protocol implementation exists, it would be more efficient to convert directly from the KafkaMessage
	//        to a binding/Message.
	//
	message := binding.ToMessage(event)

	// Extract The Relevant Knative Subscription Event URLs
	var destinationURL *url.URL
	if !subscription.SubscriberURI.IsEmpty() {
		destinationURL = subscription.SubscriberURI.URL()
	}
	var replyURL *url.URL
	if !subscription.ReplyURI.IsEmpty() {
		replyURL = subscription.ReplyURI.URL()
	}
	var deadLetterURL *url.URL
	if subscription.Delivery != nil && subscription.Delivery.DeadLetterSink != nil && !subscription.Delivery.DeadLetterSink.URI.IsEmpty() {
		// TODO - Currently ignoring dead-letter configuration due to wrapping retry implementation - would send one deadletter for every retry :(
		// deadLetterURL = subscription.Delivery.DeadLetterSink.URI.URL()
		logger.Warn("Subscription Delivery DeadLetterSink Not Currently Supported!")
	}

	// Attempt To Dispatch The CloudEvent Message Via Knative Message Dispatcher With Retry Wrapper
	err := retryRunner.Run(context.Background(), func(ctx context.Context) error {
		err := c.messageDispatcher.DispatchMessage(ctx, message, nil, destinationURL, replyURL, deadLetterURL)
		return c.logResponse(err)
	})

	// Retries failed
	if err != nil {
		logger.Error("Failed to send after configured number of retries", zap.Error(err))
		return err
	}

	// Return Success
	return nil
}

// Utility Function For Logging The Response From A Dispatch Request
func (c RetriableCloudEventClient) logResponse(err error) error {

	if err == nil {
		c.logger.Debug("Successfully Sent Cloud Event To Subscription")
		return nil
	} else {
		statusCode := c.parseHttpStatusCodeFromError(err)
		logger := c.logger.With(zap.Error(err), zap.Int("StatusCode", statusCode))
		if statusCode >= 500 || statusCode == 404 || statusCode == 429 {
			logger.Warn("Failed to send message to subscriber service, retrying")
			return errors.New("Server returned a bad response code")
		} else if statusCode >= 300 {
			logger.Warn("Failed to send message to subscriber service, not retrying")
		} else if statusCode == -1 {
			logger.Warn("No response code detected in error, retrying")
			return errors.New("No response code detected in error, retrying")
		}
		return nil
	}
}

//
// Parse The HTTP Status Code From The Specified Error - HACK ALERT!
//
// This is necessary due to the Knative Eventing Channel's MessageDispatcher implementation
// NOT returning the HTTP Status Code other than as part of the error string, combined with
// the DispatchMessage() functionality not supporting retry.  We do, though, want to use their
// implementation in order to align with standards and take advantage of tracing logic.
// Therefore, since we are providing our own wrapping retry logic, which depends on the
// resulting HTTP Status Code, we're forced to parse it from the error text here.  It is
// hoped that we can provide a better solution directly in the knative channel implementation
// in the near future.
//
func (c RetriableCloudEventClient) parseHttpStatusCodeFromError(err error) int {

	// Default Value Indicates The HTTP Status Code Was Not Found In Error
	statusCode := -1

	// Validate The Error
	if err != nil && len(err.Error()) >= 3 {

		// Get The Error String Value
		errString := err.Error()

		// Match/Group Any 3 Digit Status Code SubString From The Error String
		statusCodeStringSubMatch := HttpStatusCodeRegExp.FindAllStringSubmatch(errString, -1)

		// If A Match Was Found
		if len(statusCodeStringSubMatch) >= 1 {

			// Log Warning If Multiple Potential HTTP StatusCodes Detected
			if len(statusCodeStringSubMatch) > 1 {
				c.logger.Warn("Encountered Multiple Possible HTTP Status Codes In Error String - Using First One")
			}

			// Take The First Potential HTTP StatusCode (For lack anything more sophisticated ;)
			statusCodeString := statusCodeStringSubMatch[0][2]

			// And Convert To An Int
			code, conversionErr := strconv.Atoi(statusCodeString)
			if conversionErr != nil {
				// Conversion Error - Should Be Impossible Due To RegExp Match - But Log A Warning Just To Be Safe ; )
				c.logger.Warn("Failed To Convert Parsed HTTP Status Code String To Int", zap.String("HTTP Status Code String", statusCodeString), zap.Error(conversionErr))
			} else {
				statusCode = code
			}
		}
	}

	// Return The HTTP Status Code
	return statusCode
}

// Convert defined max retry time to the approximate number
// of retries, taking into account the exponential backoff algorithm
func (c RetriableCloudEventClient) calculateNumberOfRetries() int {
	return int(math.Round(math.Log2(float64(c.maxRetryTime)/float64(c.initialRetryInterval))) + 1)
}

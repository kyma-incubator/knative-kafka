package dispatcher

import (
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaconsumer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/consumer"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/dispatcher/internal/client"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

// Define A Dispatcher Config Struct To Hold Configuration
type DispatcherConfig struct {
	Brokers                     string
	Topic                       string
	Offset                      string
	PollTimeoutMillis           int
	OffsetCommitCount           int64
	OffsetCommitDuration        time.Duration
	OffsetCommitDurationMinimum time.Duration
	Username                    string
	Password                    string
	Client                      client.RetriableClient
	ChannelKey                  string
}

type Subscription struct {
	URI     string
	GroupId string
}

type ConsumerOffset struct {
	consumer         kafkaconsumer.ConsumerInterface
	lastOffsetCommit time.Time
	offsets          map[int32]kafka.Offset
}

// Define a Dispatcher Struct to hold Dispatcher Config and dispatcher implementation details
type Dispatcher struct {
	DispatcherConfig
	consumers          map[Subscription]*ConsumerOffset
	consumerUpdateLock sync.Mutex
}

// Create A New Dispatcher Of Specified Configuration
func NewDispatcher(dispatcherConfig DispatcherConfig) *Dispatcher {

	// Create The Dispatcher With Specified Configuration
	dispatcher := &Dispatcher{
		DispatcherConfig: dispatcherConfig,
		consumers:        make(map[Subscription]*ConsumerOffset),
	}

	// Return The Dispatcher
	return dispatcher
}

// Close The Running Consumers Connections
func (d *Dispatcher) StopConsumers() {
	for subscription, _ := range d.consumers {
		d.stopConsumer(subscription)
	}
}

func (d *Dispatcher) stopConsumer(subscription Subscription) {
	log.Logger().Info("Stopping Consumer", zap.String("GroupId", subscription.GroupId), zap.String("topic", d.Topic), zap.String("URI", subscription.URI))
	err := d.consumers[subscription].consumer.Close()
	if err != nil {
		log.Logger().Error("Unable To Stop Consumer", zap.Error(err))
	}
	delete(d.consumers, subscription)
}

// Create And Start A Consumer
func (d *Dispatcher) initConsumer(subscription Subscription) (*ConsumerOffset, error) {

	// Create Consumer
	log.Logger().Info("Creating Consumer", zap.String("GroupId", subscription.GroupId), zap.String("topic", d.Topic), zap.String("URI", subscription.URI))
	consumer, err := kafkaconsumer.CreateConsumer(d.Brokers, subscription.GroupId, d.Offset, d.Username, d.Password)
	if err != nil {
		log.Logger().Error("Failed To Create New Consumer", zap.Error(err))
		return nil, err
	}

	// Subscribe To The Topic
	err = consumer.Subscribe(d.Topic, nil)
	if err != nil {
		log.Logger().Error("Failed To Subscribe To Topic", zap.String("Topic", d.Topic), zap.Error(err))
		return nil, err
	}

	consumerOffset := ConsumerOffset{consumer: consumer, lastOffsetCommit: time.Now(), offsets: make(map[int32]kafka.Offset)}

	// Start Consuming Messages From Topic (Async)
	go d.handleKafkaMessages(consumerOffset, subscription)

	// Return The Consumer
	return &consumerOffset, nil
}

// Consumer Message Handler - Wait For Consumer Messages, Log Them & Commit The Offset
func (d *Dispatcher) handleKafkaMessages(consumerOffset ConsumerOffset, subscription Subscription) {

	// Configure The Logger
	logger := log.Logger().With(zap.String("GroupID", subscription.GroupId))

	// Infinite Message Processing Loop
	for {

		// Poll For A New Event Message Until We Get One (Timeout is how long to wait before requesting again)
		event := consumerOffset.consumer.Poll(d.PollTimeoutMillis)

		if event == nil {
			// Commit Offsets If The Amount Of Time Since Last Commit Is "OffsetCommitDuration" Or More
			currentTimeDuration := time.Now().Sub(consumerOffset.lastOffsetCommit)

			if currentTimeDuration > d.OffsetCommitDurationMinimum && currentTimeDuration >= d.OffsetCommitDuration {
				d.commitOffsets(logger, consumerOffset)
			}

			// Continue Polling For Events
			continue
		}

		// Handle Event Messages / Errors Based On Type
		switch e := event.(type) {

		case *kafka.Message:
			// Dispatch The Message - Send To Target URL
			logger.Debug("Received Kafka Message - Dispatching",
				zap.String("Message", string(e.Value)),
				zap.String("Topic", *e.TopicPartition.Topic),
				zap.Int32("Partition", e.TopicPartition.Partition),
				zap.Any("Offset", e.TopicPartition.Offset))

			_ = d.Client.Dispatch(convertToCloudEvent(e), subscription.URI) // Ignore Errors - Dispatcher Will Retry And We're Moving On!

			// Update Stored Offsets Based On The Processed Message
			d.updateOffsets(consumerOffset.consumer, e)
			currentTimeDuration := time.Now().Sub(consumerOffset.lastOffsetCommit)

			// If "OffsetCommitCount" Number Of Messages Have Been Processed Since Last Offset Commit, Then Do One Now
			if currentTimeDuration > d.OffsetCommitDurationMinimum &&
				int64(e.TopicPartition.Offset-consumerOffset.offsets[e.TopicPartition.Partition]) >= d.OffsetCommitCount {
				d.commitOffsets(logger, consumerOffset)
			}

		case kafka.Error:
			logger.Warn("Received Kafka Error", zap.Error(e))

		default:
			logger.Info("Received Unknown Event Type - Ignoring", zap.String("Message", e.String()))
		}
	}
}

// Store Updated Offsets For The Partition If Consumer Still Has It Assigned
func (d *Dispatcher) updateOffsets(consumer kafkaconsumer.ConsumerInterface, message *kafka.Message) {
	// Store The Updated Offsets
	offsets := []kafka.TopicPartition{message.TopicPartition}
	offsets[0].Offset++
	consumer.StoreOffsets(offsets)
}

// Commit The Stored Offsets For Partitions Still Assigned To This Consumer
func (d *Dispatcher) commitOffsets(logger *zap.Logger, consumerOffset ConsumerOffset) {
	partitions, err := consumerOffset.consumer.Commit()
	if err != nil {
		// Don't Log Error If There Weren't Any Offsets To Commit
		if err.(kafka.Error).Code() != kafka.ErrNoOffset {
			logger.Error("Error Committing Offsets", zap.Error(err))
		}
	} else {
		//logger.Debug("Committed Partitions", zap.Any("partitions", partitions))
		consumerOffset.offsets = make(map[int32]kafka.Offset)
		for _, partition := range partitions {
			consumerOffset.offsets[partition.Partition] = partition.Offset
		}
	}
	consumerOffset.lastOffsetCommit = time.Now()
}

func (d *Dispatcher) UpdateSubscriptions(subscriptions []Subscription) (map[Subscription]error, error) {

	failedSubscriptions := make(map[Subscription]error)
	activeSubscriptions := make(map[Subscription]bool)

	d.consumerUpdateLock.Lock()
	defer d.consumerUpdateLock.Unlock()

	for _, subscription := range subscriptions {

		// Create Consumer If Doesn't Already Exist
		if _, ok := d.consumers[subscription]; !ok {
			consumer, err := d.initConsumer(subscription)
			if err != nil {
				failedSubscriptions[subscription] = err
			} else {
				d.consumers[subscription] = consumer
				activeSubscriptions[subscription] = true
			}
		}
	}

	// Stop Consumers For Removed Subscriptions
	for runningSubscription, _ := range d.consumers {
		if !activeSubscriptions[runningSubscription] {
			d.stopConsumer(runningSubscription)
		}
	}

	return failedSubscriptions, nil
}

// Convert Kafka Message To A Cloud Event, Eventually We Should Consider Writing A Cloud Event SDK Codec For This
func convertToCloudEvent(message *kafka.Message) cloudevents.Event {
	event := cloudevents.NewEvent(cloudevents.VersionV03)
	event.SetData(message.Value)
	for _, header := range message.Headers {
		h := header.Key
		var v = string(header.Value)
		switch h {
		case "ce_datacontenttype":
			event.SetDataContentType(v)
		case "ce_specversion":
			event.SetSpecVersion(v)
		case "ce_type":
			event.SetType(v)
		case "ce_source":
			event.SetSource(v)
		case "ce_id":
			event.SetID(v)
		case "ce_time":
			t, _ := time.Parse(time.RFC3339, v)
			event.SetTime(t)
		case "ce_subject":
			event.SetSubject(v)
		case "ce_dataschema":
			event.SetDataSchema(v)
		default:
			// Must Be An Extension, Remove The "ce_" Prefix
			event.SetExtension(strings.TrimPrefix(h, "ce_"), v)
		}
	}

	return event
}

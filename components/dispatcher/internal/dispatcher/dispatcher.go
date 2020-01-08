package dispatcher

import (
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaconsumer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/consumer"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/dispatcher/internal/client"
	"go.uber.org/zap"
	eventingchannels "knative.dev/eventing/pkg/channel"
	"strings"
	"time"
)

// Define A Dispatcher Config Struct To Hold Configuration
type DispatcherConfig struct {
	SubscriberUri               string
	Brokers                     string
	Topic                       string
	Offset                      string
	Concurrency                 int64
	GroupId                     string
	PollTimeoutMillis           int
	OffsetCommitCount           int64
	OffsetCommitDuration        time.Duration
	OffsetCommitDurationMinimum time.Duration
	Username                    string
	Password                    string
	Client                      client.RetriableClient
}

// Define a Dispatcher Struct to hold Dispatcher Config and dispatcher implementation details
type Dispatcher struct {
	DispatcherConfig
	consumers        []kafkaconsumer.ConsumerInterface
	offsetMessages   []map[int32]kafka.Offset
	lastOffsetCommit []time.Time
}

// Create A New Dispatcher Of Specified Configuration
func NewDispatcher(dispatcherConfig DispatcherConfig) *Dispatcher {

	// Create The Dispatcher With Specified Configuration
	dispatcher := &Dispatcher{
		DispatcherConfig: dispatcherConfig,

		consumers:        make([]kafkaconsumer.ConsumerInterface, 0, dispatcherConfig.Concurrency),
		offsetMessages:   make([]map[int32]kafka.Offset, dispatcherConfig.Concurrency, dispatcherConfig.Concurrency),
		lastOffsetCommit: make([]time.Time, dispatcherConfig.Concurrency),
	}

	// Return The Dispatcher
	return dispatcher
}

// Start The Kafka Consumers
func (d *Dispatcher) StartConsumers() {

	// Create The Configured Number Of Concurrent Consumers
	for index := int64(0); index < d.Concurrency; index++ {

		// Create & Start A Single Kafka Consumer For Configured GroupID
		consumer, err := d.initConsumer(d.GroupId, index)
		if err != nil {
			log.Logger().Fatal("Failed To Start Consumer", zap.Error(err), zap.Int64("Index", index))
		}

		// Track The Consumer For Clean Shutdown
		d.consumers = append(d.consumers, consumer)
	}
}

// Close The Running Consumers Connections
func (d *Dispatcher) StopConsumers() {
	for _, consumer := range d.consumers {
		err := consumer.Close()
		if err != nil {
			log.Logger().Error("Failed To Close Consumer", zap.Error(err))
		}
	}
}

// Create And Start A Consumer
func (d *Dispatcher) initConsumer(groupId string, index int64) (kafkaconsumer.ConsumerInterface, error) {

	// Create Consumer
	log.Logger().Info("Creating Consumer", zap.String("GroupID", groupId), zap.Int64("Index", index))
	consumer, err := kafkaconsumer.CreateConsumer(d.Brokers, d.GroupId, d.Offset, d.Username, d.Password)
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

	// Start Consuming Messages From Topic (Async)
	go d.handleKafkaMessages(consumer, index, groupId)

	// Return The Consumer
	return consumer, nil
}

// Consumer Message Handler - Wait For Consumer Messages, Log Them & Commit The Offset
func (d *Dispatcher) handleKafkaMessages(consumer kafkaconsumer.ConsumerInterface, index int64, groupId string) {

	// Configure The Logger
	logger := log.Logger().With(zap.String("GroupID", groupId), zap.Int64("Index", index))

	// Initialize The Consumer's OffsetMessages Array With A Map For Each PartitionKey
	d.offsetMessages[index] = make(map[int32]kafka.Offset)
	d.lastOffsetCommit[index] = time.Now()

	// Infinite Message Processing Loop
	for {

		// Poll For A New Event Message Until We Get One (Timeout is how long to wait before requesting again)
		event := consumer.Poll(d.PollTimeoutMillis)

		if event == nil {
			// Commit Offsets If The Amount Of Time Since Last Commit Is "OffsetCommitDuration" Or More
			currentTimeDuration := time.Now().Sub(d.lastOffsetCommit[index])

			if currentTimeDuration > d.OffsetCommitDurationMinimum && currentTimeDuration >= d.OffsetCommitDuration {

				d.commitOffsets(logger, consumer, index)
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

			_ = d.Client.Dispatch(convertToCloudEvent(e)) // Ignore Errors - Dispatcher Will Retry And We're Moving On!
			eventchannels.NewDispatcher()

			// Update Stored Offsets Based On The Processed Message
			d.updateOffsets(logger, consumer, e)
			currentTimeDuration := time.Now().Sub(d.lastOffsetCommit[index])

			// If "OffsetCommitCount" Number Of Messages Have Been Processed Since Last Offset Commit, Then Do One Now
			if currentTimeDuration > d.OffsetCommitDurationMinimum &&
				int64(e.TopicPartition.Offset-d.offsetMessages[index][e.TopicPartition.Partition]) >= d.OffsetCommitCount {
				d.commitOffsets(logger, consumer, index)
			}

		case kafka.Error:
			logger.Warn("Received Kafka Error", zap.Error(e))

		default:
			logger.Info("Received Unknown Event Type - Ignoring", zap.String("Message", e.String()))
		}
	}
}

// Store Updated Offsets For The Partition If Consumer Still Has It Assigned
func (d *Dispatcher) updateOffsets(logger *zap.Logger, consumer kafkaconsumer.ConsumerInterface, message *kafka.Message) {
	// Store The Updated Offsets
	offsets := []kafka.TopicPartition{message.TopicPartition}
	offsets[0].Offset++
	consumer.StoreOffsets(offsets)
}

// Commit The Stored Offsets For Partitions Still Assigned To This Consumer
func (d *Dispatcher) commitOffsets(logger *zap.Logger, consumer kafkaconsumer.ConsumerInterface, index int64) {
	partitions, err := consumer.Commit()
	if err != nil {
		// Don't Log Error If There Weren't Any Offsets To Commit
		if err.(kafka.Error).Code() != kafka.ErrNoOffset {
			logger.Error("Error Committing Offsets", zap.Error(err))
		}
	} else {
		logger.Info("Committed Partitions", zap.Any("partitions", partitions))
		d.offsetMessages[index] = make(map[int32]kafka.Offset)
		for _, partition := range partitions {
			d.offsetMessages[index][partition.Partition] = partition.Offset
		}
	}
	d.lastOffsetCommit[index] = time.Now()
}

// Convert Kafka Message To A Cloud Event, Eventually We Should Consider Writing A Cloud Event SDK Codec For This
func convertToCloudEvent(message *kafka.Message) cloudevents.Event {
	event := cloudevents.NewEvent(cloudevents.VersionV03)
	event.SetData(message.Value)
	for _, header := range message.Headers {
		h := header.Key
		v := string(header.Value)
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

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
	SubscriberUri               string
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
	ChannelName                 string
}

type Subscription struct {
	URI     string
	GroupId string
}

type DispatcherClient struct {
	Client   client.RetriableClient
	Consumer kafkaconsumer.ConsumerInterface
}

// Define a Dispatcher Struct to hold Dispatcher Config and dispatcher implementation details
type Dispatcher struct {
	DispatcherConfig
	consumers          map[Subscription]DispatcherClient
	offsetMessages     []map[int32]kafka.Offset
	lastOffsetCommit   []time.Time
	consumerUpdateLock sync.Mutex
}

// Create A New Dispatcher Of Specified Configuration
func NewDispatcher(dispatcherConfig DispatcherConfig) *Dispatcher {

	// Create The Dispatcher With Specified Configuration
	dispatcher := &Dispatcher{
		DispatcherConfig: dispatcherConfig,

		consumers:        make(map[Subscription]DispatcherClient),
		offsetMessages:   make([]map[int32]kafka.Offset, 0),
		lastOffsetCommit: make([]time.Time, 0),
	}

	// Return The Dispatcher
	return dispatcher
}

// Start The Kafka Consumers
func (d *Dispatcher) StartConsumers() {
	//
	//// Create The Configured Number Of Concurrent Consumers
	//for index := int64(0); index < d.Concurrency; index++ {
	//
	//	// Create & Start A Single Kafka Consumer For Configured GroupID
	//	consumer, err := d.initConsumer(d.GroupId)
	//	if err != nil {
	//		log.Logger().Fatal("Failed To Start Consumer", zap.Error(err), zap.Int64("Index", index))
	//	}
	//
	//	// Track The Consumer For Clean Shutdown
	//	d.consumers = append(d.consumers, consumer)
	//}
}

// Close The Running Consumers Connections
func (d *Dispatcher) StopConsumers() {
	for _, consumer := range d.consumers {
		err := consumer.Consumer.Close()
		if err != nil {
			log.Logger().Error("Failed To Close Consumer", zap.Error(err))
		}
	}
}

// Create And Start A Consumer
func (d *Dispatcher) initConsumer(groupId string, uri string) (*DispatcherClient, error) {

	// Create Consumer
	log.Logger().Info("Creating Consumer", zap.String("GroupId", groupId), zap.String("topic", d.Topic), zap.String("URI", uri))
	consumer, err := kafkaconsumer.CreateConsumer(d.Brokers, groupId, d.Offset, d.Username, d.Password)
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
	go d.handleKafkaMessages(consumer, groupId)

	dispatcherClient := DispatcherClient{Consumer: consumer, Client: client.NewRetriableCloudEventClient(uri, true, 500, 5000)}

	// Return The Consumer
	return &dispatcherClient, nil
}

// Consumer Message Handler - Wait For Consumer Messages, Log Them & Commit The Offset
func (d *Dispatcher) handleKafkaMessages(consumer kafkaconsumer.ConsumerInterface, groupId string) {

	// Configure The Logger
	logger := log.Logger().With(zap.String("GroupID", groupId))

	// Initialize The Consumer's OffsetMessages Array With A Map For Each PartitionKey
	//d.offsetMessages[index] = make(map[int32]kafka.Offset)
	//d.lastOffsetCommit[index] = time.Now()

	// Infinite Message Processing Loop
	for {

		// Poll For A New Event Message Until We Get One (Timeout is how long to wait before requesting again)
		event := consumer.Poll(d.PollTimeoutMillis)

		if event == nil {
			//// Commit Offsets If The Amount Of Time Since Last Commit Is "OffsetCommitDuration" Or More
			//currentTimeDuration := time.Now().Sub(d.lastOffsetCommit[index])
			//
			//if currentTimeDuration > d.OffsetCommitDurationMinimum && currentTimeDuration >= d.OffsetCommitDuration {
			//
			//	d.commitOffsets(logger, consumer, index)
			//}

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

			// Update Stored Offsets Based On The Processed Message
			//d.updateOffsets(logger, consumer, e)
			//currentTimeDuration := time.Now().Sub(d.lastOffsetCommit[index])
			//
			//// If "OffsetCommitCount" Number Of Messages Have Been Processed Since Last Offset Commit, Then Do One Now
			//if currentTimeDuration > d.OffsetCommitDurationMinimum &&
			//	int64(e.TopicPartition.Offset-d.offsetMessages[index][e.TopicPartition.Partition]) >= d.OffsetCommitCount {
			//	d.commitOffsets(logger, consumer, index)
			//}

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

func (d *Dispatcher) UpdateSubscriptions(subscriptions []Subscription) {

	d.consumerUpdateLock.Lock()
	defer d.consumerUpdateLock.Unlock()

	for _, subscription := range subscriptions {

		// Create Consumer If Doesn't Already Exist
		if _, ok := d.consumers[subscription]; !ok {
			consumer, _ := d.initConsumer(subscription.GroupId, subscription.URI)
			d.consumers[subscription] = *consumer
		}
	}

	// Need To Handle Subscription Deletions

	//keys := make([]Subscription, len(d.consumers))
	//for key, _ := range d.consumers {
	//	keys = append(keys, key)
	//}

	//deleted := difference(keys, subscriptions)
	//for _, deletedSub := range deleted {
	//	// stop consumer
	//	d.consumers[deletedSub].Close()
	//	delete(d.consumers, deletedSub)
	//}

}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []Subscription) []Subscription {
	mb := make(map[Subscription]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []Subscription
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
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

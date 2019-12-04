package channel

import (
	"errors"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaproducer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/producer"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"go.uber.org/zap"
)

// Constants
const (
	KafkaProducerConfigPropertyBootstrapServers = "bootstrap.servers"
	KafkaProducerConfigPropertyPartitioner      = "partitioner"
	KafkaProducerConfigPropertyPartitionerValue = "murmur2_random"
	KafkaProducerConfigPropertyIdempotence      = "enable.idempotence"
	KafkaProducerConfigPropertyUsername         = "sasl.username"
	KafkaProducerConfigPropertyPassword         = "sasl.password"
)

// Define A Channel Config Struct To Hold Configuration
type ChannelConfig struct {
	Brokers       string
	Topic         string
	ClientId      string
	KafkaUsername string
	KafkaPassword string
}

// Define a Channel Struct to hold channel config and channel implementation details
type Channel struct {
	ChannelConfig
	producer    kafkaproducer.ProducerInterface
	stopChan    chan struct{}
	stoppedChan chan struct{}
}

// Create A New Channel Of Specified Configuration
func NewChannel(channelConfig ChannelConfig) *Channel {

	// Create The Channel With Specified Configuration
	channel := &Channel{
		ChannelConfig: channelConfig,
		stopChan:      make(chan struct{}),
		stoppedChan:   make(chan struct{}),
	}

	// Create The Kafka Producer
	producer, err := kafkaproducer.CreateProducer(channelConfig.Brokers, channelConfig.KafkaUsername, channelConfig.KafkaPassword)
	if err != nil {
		log.Logger().Fatal("Failed To Create Kafka Producer - Exiting", zap.Error(err))
	}

	// Assign The Producer To Channel And Start Processing!
	channel.producer = producer
	channel.processSuccessesAndErrors()

	// Return The Channel
	return channel
}

//
// Channel Functions
//

// Send The Specified Message To The Kafka Topic With Partition Key And Wait For The Delivery Report
func (c *Channel) SendMessage(partitionKey string, message []byte) error {

	// Create The Producer Message
	producerMessage := c.createProducerMessage(partitionKey, message)

	// Create a channel that corresponds to only this message being sent. The Kafka Producer will deliver the
	// report on this channel thus informing us that the message is persisted in kafka.
	deliveryReportChannel := make(chan kafka.Event)

	// Send The Producer Message To Kafka Via Producer
	err := c.producer.Produce(&producerMessage, deliveryReportChannel)
	if err != nil {
		log.Logger().Error("Failed to Produce Message", zap.Error(err))
		return err
	}

	// Block on the deliveryReportChannel for this message and return the m.TopicPartition.Error if there is one.
	select {
	case msg := <-deliveryReportChannel:
		// Close the channel for safety
		close(deliveryReportChannel)
		switch ev := msg.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				log.Logger().Error("Delivery failed", zap.Error(m.TopicPartition.Error))
			} else {
				log.Logger().Debug("Delivered message to kafka",
					zap.String("topic", *m.TopicPartition.Topic),
					zap.Int32("partition", m.TopicPartition.Partition),
					zap.String("offset", m.TopicPartition.Offset.String()))
			}
			return m.TopicPartition.Error
		case kafka.Error:
			log.Logger().Warn("Kafka error", zap.Error(ev))
			return errors.New("kafka error occurred")
		default:
			log.Logger().Info("Ignored event", zap.String("event", ev.String()))
			return errors.New("kafka ignored the event")
		}
	}
}

// Close The Channel (Stop Processing)
func (c *Channel) Close() {

	// Setup The Logger
	logger := log.Logger().With(zap.String("ClientId", c.ClientId), zap.String("Topic", c.Topic))

	// Stop Processing Success/Error Messages From Producer
	logger.Info("Stopping Kafka Producer Success/Error Processing")
	close(c.stopChan)
	<-c.stoppedChan // Block On Stop Completion

	// Close The Producer
	logger.Info("Closing Kafka Producer")
	c.producer.Close()
}

// Fork A Go Routine With Infinite Loop To Process Any
// Errors That The Kafka Producer Might Send Back To The
// Default Delivery Report Channel
func (c *Channel) processSuccessesAndErrors() {
	go func() {
		for {
			select {
			case msg := <-c.producer.Events():
				switch ev := msg.(type) {
				case *kafka.Message:
					log.Logger().Warn("Message arrived on the wrong channel", zap.Any("Message", msg))
				case kafka.Error:
					log.Logger().Warn("Kafka error", zap.Error(ev))
				default:
					log.Logger().Info("Ignored event", zap.String("event", ev.String()))
				}

			case <-c.stopChan:
				close(c.stoppedChan) // Inform On Stop Completion & Return
				return
			}
		}
	}()
}

// Create A Kafka Message From Message
func (c *Channel) createProducerMessage(partitionKey string, message []byte) kafka.Message {

	// Create The Producer Message
	producerMessage := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &c.Topic,
			Partition: kafka.PartitionAny, // Required For Producer Level Partitioner! (see KafkaProducerConfigPropertyPartitioner)
		},
		Value: message,
	}

	// Populate The PartitionKey If Valid
	if len(partitionKey) > 0 {
		producerMessage.Key = []byte(partitionKey)
	}

	// Return The Constructed Kafka Message
	return producerMessage
}

package producer

import (
	"errors"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/message"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/util"
	kafkaproducer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/producer"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/prometheus"
	"go.uber.org/zap"
	eventingChannel "knative.dev/eventing/pkg/channel"
)

// Package Variables
var (
	kafkaProducer  kafkaproducer.ProducerInterface
	stopChannel    chan struct{}
	stoppedChannel chan struct{}
	metrics        *prometheus.MetricsServer
)

// Wrapper Around Common Kafka Producer Creation To Facilitate Unit Testing
var createProducerFunctionWrapper = func(brokers string, username string, password string) (kafkaproducer.ProducerInterface, error) {
	return kafkaproducer.CreateProducer(brokers, username, password)
}

// Initialize The Producer
func InitializeProducer(brokers string, username string, password string, metricsServer *prometheus.MetricsServer) error {

	// Create The Producer Using The Specified Kafka Authentication
	producer, err := createProducerFunctionWrapper(brokers, username, password)
	if err != nil {
		log.Logger().Error("Failed To Create Kafka Producer - Exiting", zap.Error(err))
		return err
	}

	// Assign The Producer Singleton Instance
	log.Logger().Info("Successfully Created Kafka Producer")
	kafkaProducer = producer

	// Reset The Stop Channels
	stopChannel = make(chan struct{})
	stoppedChannel = make(chan struct{})

	// Used For Reporting Kafka Metrics
	metrics = metricsServer

	// Fork A Go Routine To Process Kafka Events Asynchronously
	log.Logger().Info("Starting Kafka Producer Event Processing")
	go processProducerEvents()

	// Return Success
	log.Logger().Info("Successfully Initialized Kafka Producer")
	return nil
}

// Produce A KafkaMessage From The Specified CloudEvent To The Specified Topic And Wait For The Delivery Report
func ProduceKafkaMessage(event cloudevents.Event, channelReference eventingChannel.ChannelReference) error {

	// Validate The Kafka Producer (Must Be Pre-Initialized)
	if kafkaProducer == nil {
		log.Logger().Error("Kafka Producer Not Initialized - Unable To Produce Message")
		return errors.New("uninitialized kafka producer - unable to produce message")
	}

	// Get The Topic Name From The ChannelReference
	topicName := util.TopicName(channelReference)
	logger := log.Logger().With(zap.String("Topic", topicName))

	// Create A Kafka Message From The CloudEvent For The Appropriate Topic
	kafkaMessage, err := message.CreateKafkaMessage(event, topicName)
	if err != nil {
		logger.Error("Failed To Create Kafka Message", zap.Error(err))
		return err
	}

	// Create A DeliveryReport Channel (Producer will report results for message here)
	deliveryReportChannel := make(chan kafka.Event)

	// Produce The Kafka Message To The Kafka Topic
	logger.Debug("Producing Kafka Message", zap.Any("Headers", kafkaMessage.Headers), zap.Any("Message", kafkaMessage.Value))
	err = kafkaProducer.Produce(kafkaMessage, deliveryReportChannel)
	if err != nil {
		log.Logger().Error("Failed To Produce Kafka Message", zap.Error(err))
		return err
	}

	// Block On The DeliveryReport Channel For The Prior Message & Return Any m.TopicPartition.Errors
	select {
	case msg := <-deliveryReportChannel:
		close(deliveryReportChannel) // Close the DeliveryReport channel for safety
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

// Close The Producer (Stop Processing)
func Close() {

	// Stop Processing Success/Error Messages From Producer
	log.Logger().Info("Stopping Kafka Producer Success/Error Processing")
	close(stopChannel)
	<-stoppedChannel // Block On Stop Completion

	// Close The Kafka Producer
	log.Logger().Info("Closing Kafka Producer")
	kafkaProducer.Close()
}

// Infinite Loop For Processing Kafka Producer Events
func processProducerEvents() {
	for {
		select {
		case msg := <-kafkaProducer.Events():
			switch ev := msg.(type) {
			case *kafka.Message:
				log.Logger().Warn("Kafka Message Arrived On The Wrong Channel", zap.Any("Message", msg))
			case kafka.Error:
				log.Logger().Warn("Kafka Error", zap.Error(ev))
			case *kafka.Stats:
				// Update Kafka Prometheus Metrics
				metrics.Observe(ev.String())
			default:
				log.Logger().Info("Ignored Event", zap.String("Event", ev.String()))
			}

		case <-stopChannel:
			log.Logger().Info("Terminating Producer Event Processing")
			close(stoppedChannel) // Inform On Stop Completion & Return
			return
		}
	}
}

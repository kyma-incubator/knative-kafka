package producer

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/constants"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/test"
	kafkaproducer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/producer"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Package Variables
var _ = log.TestLogger() // Force The Use Of The TestLogger!

// TODO - Implement Test for initialize?  probably have to wrap the call to the knative createProducer()
//      - check code coverage - not great

// Test The InitializeProducer Functionality
func TestInitializeProducer(t *testing.T) {

	// Test Data
	brokers := "TestBrokers"
	username := "TestUsername"
	password := "TestPassword"
	mockProducer := test.NewMockProducer(test.TopicName)

	// Stub The Kafka Producer Creation Wrapper With Test Version Returning MockProducer
	createProducerFunctionWrapper = func(brokers string, username string, password string) (kafkaproducer.ProducerInterface, error) {
		return mockProducer, nil
	}

	// Perform The Test
	err := InitializeProducer(brokers, username, password, prometheus.NewMetricsServer("8888", "/metrics"))

	// Verify The Results
	assert.Nil(t, err)
	assert.Equal(t, mockProducer, kafkaProducer)

	// Close The Producer (Or Subsequent Tests Will Fail Because processProducerEvents() GO Routine Is Still Running)
	Close()
}

// Test The ProduceKafkaMessage() Functionality
func TestProduceKafkaMessage(t *testing.T) {

	// Test Data
	event := test.CreateCloudEvent(cloudevents.CloudEventsVersionV1)
	channelReference := test.CreateChannelReference(test.ChannelName, test.ChannelNamespace)

	// Replace The Package Singleton With A Mock Producer
	mockProducer := test.NewMockProducer(test.TopicName)
	kafkaProducer = mockProducer

	// Perform The Test & Verify Results
	err := ProduceKafkaMessage(event, channelReference)
	assert.Nil(t, err)

	// Block On The MockProducer's Channel & Verify Event Was Produced
	kafkaMessage := <-mockProducer.ProduceChannel()
	assert.NotNil(t, kafkaMessage)
	assert.Equal(t, test.PartitionKey, string(kafkaMessage.Key))
	assert.Equal(t, test.EventDataJson, kafkaMessage.Value)
	assert.Equal(t, test.TopicName, *kafkaMessage.TopicPartition.Topic)
	assert.Equal(t, kafka.PartitionAny, kafkaMessage.TopicPartition.Partition)
	assert.Equal(t, kafka.Offset(0), kafkaMessage.TopicPartition.Offset)
	assert.Nil(t, kafkaMessage.TopicPartition.Error)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeySpecVersion, cloudevents.CloudEventsVersionV1)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeyType, test.EventType)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeyId, test.EventId)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeySource, test.EventSource)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeyDataContentType, test.EventDataContentType)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeySubject, test.EventSubject)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeyDataSchema, test.EventDataSchema)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeyPartitionKey, test.PartitionKey)
}

// Test The Producer's Close() Functionality
func TestClose(t *testing.T) {

	// Replace The Package Singleton With A Mock Producer
	mockProducer := test.NewMockProducer(test.TopicName)
	kafkaProducer = mockProducer

	// Reset The Stop Channels
	stopChannel = make(chan struct{})
	stoppedChannel = make(chan struct{})

	// Block On The StopChannel & Close The StoppedChannel (Play the part of processProducerEvents())
	go func() {
		<-stopChannel
		close(stoppedChannel)
	}()

	// Perform The Test
	Close()

	// Verify The Mock Producer Was Closed
	assert.True(t, mockProducer.Closed())
}

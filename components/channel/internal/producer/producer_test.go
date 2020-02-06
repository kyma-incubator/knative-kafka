package producer

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/constants"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/test"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Package Variables
var _ = log.TestLogger() // Force The Use Of The TestLogger!

// TODO - Implement Test for initialize?  probably have to wrap the call to the knative createProducer()
//      - check code coverage - not great

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
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.HeaderKeyCeSpecVersion, cloudevents.CloudEventsVersionV1)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.HeaderKeyCeType, test.EventType)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.HeaderKeyCeId, test.EventId)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.HeaderKeyCeSource, test.EventSource)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.HeaderKeyCeDataContentType, test.EventDataContentType)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.HeaderKeyCeSubject, test.EventSubject)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.HeaderKeyCeDataSchema, test.EventDataSchema)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.HeaderKeyCePartitionKey, test.PartitionKey)
}

// Test The Producer's Close() Functionality
func TestClose(t *testing.T) {

	// Replace The Package Singleton With A Mock Producer
	mockProducer := test.NewMockProducer(test.TopicName)
	kafkaProducer = mockProducer

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

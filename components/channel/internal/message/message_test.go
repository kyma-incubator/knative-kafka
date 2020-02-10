package message

import (
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/constants"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/test"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Package Variables
var (
	_ = log.TestLogger() // Force The Use Of The TestLogger!
)

// Test All Permutations Of The CreateKafkaMessage() Functionality
func TestCreateKafkaMessage(t *testing.T) {

	// Create The Test CloudEvent
	cloudEvent := test.CreateCloudEvent(test.EventVersion)

	// Perform The Test
	kafkaMessage, err := CreateKafkaMessage(cloudEvent, test.TopicName)

	// Verify The Results
	assert.Nil(t, err)
	assert.NotNil(t, kafkaMessage)
	assert.Equal(t, test.TopicName, *kafkaMessage.TopicPartition.Topic)
	assert.Equal(t, test.TopicName, *kafkaMessage.TopicPartition.Topic)
	assert.Equal(t, []byte(test.PartitionKey), kafkaMessage.Key)
	assert.Equal(t, test.EventDataJson, kafkaMessage.Value)
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeySpecVersion, cloudEvent.SpecVersion())
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeyType, cloudEvent.Type())
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeySource, cloudEvent.Source())
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeyId, cloudEvent.ID())
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeyDataContentType, cloudEvent.DataContentType())
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeySubject, cloudEvent.Subject())
	test.ValidateKafkaMessageHeader(t, kafkaMessage.Headers, constants.CeKafkaHeaderKeyDataSchema, cloudEvent.DataSchema())
}

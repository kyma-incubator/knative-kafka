package admin

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

// Mock AdminClient Reference
var mockAdminClient AdminClientInterface

// Test The CreateAdminClient() Functionality
func TestCreateAdminClient(t *testing.T) {

	// Test Data
	namespace := "TestNamespace"
	adminClientType := Kafka
	mockAdminClient = &MockAdminClient{}

	// Test Logger
	testLogger := log.TestLogger()

	// Replace The NewAdminClient Wrapper To Provide Mock AdminClient & Defer Reset
	newAdminClientWrapperPlaceholder := NewAdminClientWrapper
	NewAdminClientWrapper = func(logger *zap.Logger, adminClientType AdminClientType, k8sNamespace string) (clientInterface AdminClientInterface, e error) {
		assert.Equal(t, testLogger, logger)
		assert.Equal(t, adminClientType, adminClientType)
		assert.Equal(t, namespace, k8sNamespace)
		return mockAdminClient, nil
	}
	defer func() { NewAdminClientWrapper = newAdminClientWrapperPlaceholder }()

	// Perform The Test
	adminClient, err := CreateAdminClient(testLogger, adminClientType, namespace)

	// Verify The Results
	assert.Nil(t, err)
	assert.NotNil(t, adminClient)
	assert.Equal(t, mockAdminClient, adminClient)
}

//
// Mock Confluent AdminClient
//

var _ AdminClientInterface = &MockAdminClient{}

type MockAdminClient struct {
	kafkaSecret string
}

func (c MockAdminClient) GetKafkaSecretName(topicName string) string {
	return c.kafkaSecret
}

func (c MockAdminClient) CreateTopics(context.Context, []kafka.TopicSpecification, ...kafka.CreateTopicsAdminOption) ([]kafka.TopicResult, error) {
	return nil, nil
}

func (c MockAdminClient) DeleteTopics(context.Context, []string, ...kafka.DeleteTopicsAdminOption) ([]kafka.TopicResult, error) {
	return nil, nil
}

func (c MockAdminClient) Close() {
	return
}

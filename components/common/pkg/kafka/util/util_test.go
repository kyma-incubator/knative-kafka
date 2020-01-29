package util

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/constants"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test The AddAuthenticationCredentials() Functionality
func TestAddAuthenticationCredentials(t *testing.T) {

	// Test Data
	username := "TestUsername"
	password := "TestPassword"

	// Test ConfigMap
	configMap := &kafka.ConfigMap{}

	// Perform The Test
	AddSaslAuthentication(configMap, constants.ConfigPropertySaslMechanismsPlain, username, password)

	// Verify The Results
	verifyConfigMapValue(t, configMap, constants.ConfigPropertySecurityProtocol, constants.ConfigPropertySecurityProtocolValue)
	verifyConfigMapValue(t, configMap, constants.ConfigPropertySaslMechanisms, constants.ConfigPropertySaslMechanismsPlain)
	verifyConfigMapValue(t, configMap, constants.ConfigPropertySaslUsername, username)
	verifyConfigMapValue(t, configMap, constants.ConfigPropertySaslPassword, password)
}

// Test The AddDebugFlags() Functionality
func TestAddDebugFlags(t *testing.T) {

	// Test Data
	flags := "TestDebugFlags"

	// Test ConfigMap
	configMap := &kafka.ConfigMap{}

	// Perform The Test
	AddDebugFlags(configMap, flags)

	// Verify The Results
	verifyConfigMapValue(t, configMap, constants.ConfigPropertyDebug, flags)
}

// Utility Function To Verify The Specified Individual ConfigMap Value
func verifyConfigMapValue(t *testing.T, configMap *kafka.ConfigMap, key string, expected kafka.ConfigValue) {
	property, err := configMap.Get(key, nil)
	assert.Nil(t, err)
	assert.Equal(t, expected, property)
}

// Test The TopicName() Functionality
func TestTopicName(t *testing.T) {

	// Test Data
	name := "TestName"
	namespace := "TestNamespace"

	// Perform The Test
	actualTopicName := TopicName(namespace, name)

	// Verify The Results
	expectedTopicName := namespace + "." + name
	assert.Equal(t, expectedTopicName, actualTopicName)
}

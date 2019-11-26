package util

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/constants"
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

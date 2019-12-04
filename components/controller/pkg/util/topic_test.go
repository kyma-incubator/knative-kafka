package util

import (
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// Test The TopicName() Functionality
func TestTopicName(t *testing.T) {

	// Test Constants
	const (
		channelName      = "TestChannelName"
		channelNamespace = "TestChannelNamespace"
		tenantId         = "TestTenantId"
		defaultTenantId  = "TestDefaultTenantId"
	)

	// Test Data
	environment := &env.Environment{DefaultTenantId: defaultTenantId}

	// Test The Default Failover - No TenantId
	channel := &kafkav1alpha1.KafkaChannel{
		ObjectMeta: metav1.ObjectMeta{Name: channelName, Namespace: channelNamespace},
	}
	expectedTopicName := defaultTenantId + "." + channelNamespace + "." + channelName
	actualTopicName := TopicName(channel, environment)
	assert.Equal(t, expectedTopicName, actualTopicName)

	// Perform The Test - Valid TenantId
	channel = &kafkav1alpha1.KafkaChannel{
		ObjectMeta: metav1.ObjectMeta{Name: channelName, Namespace: channelNamespace},
		Spec:       kafkav1alpha1.KafkaChannelSpec{TenantId: tenantId},
	}
	expectedTopicName = tenantId + "." + channelNamespace + "." + channelName
	actualTopicName = TopicName(channel, environment)
	assert.Equal(t, expectedTopicName, actualTopicName)
}

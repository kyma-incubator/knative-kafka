package util

import (
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
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
	)

	// The KafkaChannel To Test
	channel := &kafkav1alpha1.KafkaChannel{
		ObjectMeta: metav1.ObjectMeta{Name: channelName, Namespace: channelNamespace},
	}

	// Perform The Test
	actualTopicName := TopicName(channel)

	// Verify The Results
	expectedTopicName := channelNamespace + "." + channelName
	assert.Equal(t, expectedTopicName, actualTopicName)
}

package channel

import (
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/test"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	knativekafkaclientset "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned"
	fakeclientset "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

// Package Variables
var _ = log.TestLogger() // Force The Use Of The TestLogger!

// Test The InitializeKafkaChannelLister() Functionality
func TestInitializeKafkaChannelLister(t *testing.T) {

	// Stub The K8S Client Creation Wrapper With Test Version Returning The Fake KafkaClient Clientset
	getKnativeKafkaClient = func(masterUrl string, kubeconfigPath string) (knativekafkaclientset.Interface, error) {
		return fakeclientset.NewSimpleClientset(), nil
	}
	
	// Perform The Test
	err := InitializeKafkaChannelLister("", "")

	// Verify The Results
	assert.Nil(t, err)
	assert.NotNil(t, kafkaChannelLister)
}

// Test All The ValidateKafkaChannel() Functionality
func TestValidateKafkaChannel(t *testing.T) {
	channelName := "TestChannelName"
	channelNamespace := "TestChannelNamespace"

	performValidateKafkaChannelTest(t, "", channelNamespace, false, corev1.ConditionFalse, true)
	performValidateKafkaChannelTest(t, channelName, "", false, corev1.ConditionFalse, true)
	performValidateKafkaChannelTest(t, channelName, channelNamespace, true, corev1.ConditionTrue, false)
	performValidateKafkaChannelTest(t, channelName, channelNamespace, true, corev1.ConditionFalse, true)
	performValidateKafkaChannelTest(t, channelName, channelNamespace, true, corev1.ConditionTrue, true)
	performValidateKafkaChannelTest(t, channelName, channelNamespace, false, corev1.ConditionFalse, true)
}

// Utility Function To Perform A Single Instance Of The ValidateKafkaChannel Test
func performValidateKafkaChannelTest(t *testing.T, channelName string, channelNamespace string, exists bool, ready corev1.ConditionStatus, err bool) {

	// Create The Channel Reference To Test
	channelReference := test.CreateChannelReference(channelName, channelNamespace)

	// Mock The Package Level KafkaChannel Lister For The Specified Use Case
	kafkaChannelLister = test.NewMockKafkaChannelLister(channelReference.Name, channelReference.Namespace, exists, ready, err)

	// Perform The Test
	validationError := ValidateKafkaChannel(channelReference)

	// Verify The Results
	assert.Equal(t, err, validationError != nil)
}

// Test The Close() Functionality
func TestClose(t *testing.T) {

	// Test With Nil stopChan Instance
	Close()

	// Initialize The stopChan Instance
	stopChan = make(chan struct{})

	// Close In The Background
	go Close()

	// Block On The stopChan
	_, ok := <-stopChan

	// Verify stopChan Was Closed
	assert.False(t, ok)
}

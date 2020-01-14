package util

import (
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// Test Data
const (
	numPartitions            = 123
	defaultNumPartitions     = 987
	replicationFactor        = 22
	defaultReplicationFactor = 33
	retentionMillis          = int64(4444)
	defaultRetentionMillis   = int64(55555)
)

// Test The ChannelLogger Functionality
func TestChannelLogger(t *testing.T) {

	// Test Logger
	logger := log.TestLogger()

	// Test Data
	channel := &kafkav1alpha1.KafkaChannel{
		ObjectMeta: metav1.ObjectMeta{Name: "TestChannelName", Namespace: "TestChannelNamespace"},
	}

	// Perform The Test
	channelLogger := ChannelLogger(logger, channel)
	assert.NotNil(t, channelLogger)
	assert.NotEqual(t, logger, channelLogger)
	channelLogger.Info("Testing Channel Logger")
}

// Test The NewChannelControllerRef Functionality
func TestNewChannelControllerRef(t *testing.T) {

	// Test Data
	subscription := &kafkav1alpha1.KafkaChannel{
		ObjectMeta: metav1.ObjectMeta{Name: "TestName"},
	}

	// Perform The Test
	controllerRef := NewChannelControllerRef(subscription)

	// Validate Results
	assert.NotNil(t, controllerRef)
	assert.Equal(t, kafkav1alpha1.SchemeGroupVersion.String(), controllerRef.APIVersion)
	assert.Equal(t, constants.KafkaChannelKind, controllerRef.Kind)
	assert.Equal(t, subscription.ObjectMeta.Name, controllerRef.Name)
	assert.True(t, *controllerRef.Controller)
}

// Test The Channel Deployment Name Formatter / Generator
func TestChannelDeploymentName(t *testing.T) {

	// Test Data
	testChannelName := "TestChannelName"
	testChannelNamespace := "TestChannelNamespace"
	channel := &kafkav1alpha1.KafkaChannel{ObjectMeta: metav1.ObjectMeta{Name: testChannelName, Namespace: testChannelNamespace}}

	// Perform The Test
	actualChannelDeploymentName := ChannelDeploymentName(channel)

	// Verify The Results
	expectedChannelDeploymentName := testChannelName + "-" + testChannelNamespace + "-channel"
	assert.Equal(t, expectedChannelDeploymentName, actualChannelDeploymentName)
}

// Test The Channel Service Name Formatter / Generator
func TestChannelServiceName(t *testing.T) {

	// Test Data
	testChannelName := "TestChannelName"
	testChannelNamespace := "TestChannelNamespace"
	channel := &kafkav1alpha1.KafkaChannel{ObjectMeta: metav1.ObjectMeta{Name: testChannelName, Namespace: testChannelNamespace}}

	// Perform The Test
	actualChannelServiceName := ChannelServiceName(channel)

	// Verify The Results
	expectedChannelServiceName := testChannelName + "-" + testChannelNamespace + "-channel"
	assert.Equal(t, expectedChannelServiceName, actualChannelServiceName)
}

// Test The Channel Host Name Formatter / Generator
func TestChannelHostName(t *testing.T) {
	testChannelName := "TestChannelName"
	testChannelNamespace := "TestChannelNamespace"
	expectedChannelHostName := testChannelName + "." + testChannelNamespace + ".channels.cluster.local"
	actualChannelHostName := ChannelHostName(testChannelName, testChannelNamespace)
	assert.Equal(t, expectedChannelHostName, actualChannelHostName)
}

// Test The NumPartitions Accessor
func TestNumPartitions(t *testing.T) {

	// Test Logger
	logger := log.TestLogger()

	// Test Data
	environment := &env.Environment{DefaultNumPartitions: defaultNumPartitions}

	// Test The Default Failover Use Case
	channel := &kafkav1alpha1.KafkaChannel{}
	actualNumPartitions := NumPartitions(channel, environment, logger)
	assert.Equal(t, defaultNumPartitions, actualNumPartitions)

	// Test The Valid NumPartitions Use Case
	channel = &kafkav1alpha1.KafkaChannel{Spec: kafkav1alpha1.KafkaChannelSpec{NumPartitions: numPartitions}}
	actualNumPartitions = NumPartitions(channel, environment, logger)
	assert.Equal(t, numPartitions, actualNumPartitions)
}

// Test The ReplicationFactor Accessor
func TestReplicationFactor(t *testing.T) {

	// Test Logger
	logger := log.TestLogger()

	// Test Data
	environment := &env.Environment{DefaultReplicationFactor: defaultReplicationFactor}

	// Test The Default Failover Use Case
	channel := &kafkav1alpha1.KafkaChannel{}
	actualReplicationFactor := ReplicationFactor(channel, environment, logger)
	assert.Equal(t, defaultReplicationFactor, actualReplicationFactor)

	// Test The Valid ReplicationFactor Use Case
	channel = &kafkav1alpha1.KafkaChannel{Spec: kafkav1alpha1.KafkaChannelSpec{ReplicationFactor: replicationFactor}}
	actualReplicationFactor = ReplicationFactor(channel, environment, logger)
	assert.Equal(t, replicationFactor, actualReplicationFactor)
}

// Test The RetentionMillis Accessor
func TestRetentionMillis(t *testing.T) {

	// Test Logger
	logger := log.TestLogger()

	// Test Data
	environment := &env.Environment{DefaultRetentionMillis: defaultRetentionMillis}

	// Test The Default Failover Use Case
	channel := &kafkav1alpha1.KafkaChannel{}
	actualRetentionMillis := RetentionMillis(channel, environment, logger)
	assert.Equal(t, defaultRetentionMillis, actualRetentionMillis)

	// Test The Valid RetentionMillis Use Case
	channel = &kafkav1alpha1.KafkaChannel{Spec: kafkav1alpha1.KafkaChannelSpec{RetentionMillis: retentionMillis}}
	actualRetentionMillis = RetentionMillis(channel, environment, logger)
	assert.Equal(t, retentionMillis, actualRetentionMillis)
}

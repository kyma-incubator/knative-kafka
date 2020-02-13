package util

import (
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"testing"
)

// Test Data
const (
	channelName              = "testname"
	channelNamespace         = "testnamespace"
	numPartitions            = 123
	defaultNumPartitions     = 987
	replicationFactor        = 22
	defaultReplicationFactor = 33
	retentionMillis          = int64(4444)
	defaultRetentionMillis   = int64(55555)
)

// Test The ChannelLogger() Functionality
func TestChannelLogger(t *testing.T) {

	// Test Logger
	logger := log.TestLogger()

	// Test Data
	channel := &knativekafkav1alpha1.KafkaChannel{
		ObjectMeta: metav1.ObjectMeta{Name: "TestChannelName", Namespace: "TestChannelNamespace"},
	}

	// Perform The Test
	channelLogger := ChannelLogger(logger, channel)
	assert.NotNil(t, channelLogger)
	assert.NotEqual(t, logger, channelLogger)
	channelLogger.Info("Testing Channel Logger")
}

// Test The ChannelKey() Functionality
func TestChannelKey(t *testing.T) {

	// Test Data
	channel := &knativekafkav1alpha1.KafkaChannel{ObjectMeta: metav1.ObjectMeta{Name: channelName, Namespace: channelNamespace}}

	// Perform The Test
	actualResult := ChannelKey(channel)

	// Verify The Results
	expectedResult := fmt.Sprintf("%s/%s", channelNamespace, channelName)
	assert.Equal(t, expectedResult, actualResult)
}

// Test The NewChannelOwnerReference() Functionality
func TestNewChannelOwnerReference(t *testing.T) {

	// Test Data
	channel := &knativekafkav1alpha1.KafkaChannel{
		ObjectMeta: metav1.ObjectMeta{Name: channelName},
	}

	// Perform The Test
	controllerRef := NewChannelOwnerReference(channel)

	// Validate Results
	assert.NotNil(t, controllerRef)
	assert.Equal(t, knativekafkav1alpha1.SchemeGroupVersion.String(), controllerRef.APIVersion)
	assert.Equal(t, constants.KafkaChannelKind, controllerRef.Kind)
	assert.Equal(t, channel.ObjectMeta.Name, controllerRef.Name)
	assert.True(t, *controllerRef.Controller)
}

// Test The ChannelDeploymentDnsSafeName() Functionality
func TestChannelDeploymentDnsSafeName(t *testing.T) {

	// Perform The Test
	actualResult := ChannelDeploymentDnsSafeName(test.KafkaSecret)

	// Verify The Results
	expectedResult := fmt.Sprintf("%s-channel", strings.ToLower(test.KafkaSecret))
	assert.Equal(t, expectedResult, actualResult)
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
	channel := &knativekafkav1alpha1.KafkaChannel{}
	actualNumPartitions := NumPartitions(channel, environment, logger)
	assert.Equal(t, defaultNumPartitions, actualNumPartitions)

	// Test The Valid NumPartitions Use Case
	channel = &knativekafkav1alpha1.KafkaChannel{Spec: knativekafkav1alpha1.KafkaChannelSpec{NumPartitions: numPartitions}}
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
	channel := &knativekafkav1alpha1.KafkaChannel{}
	actualReplicationFactor := ReplicationFactor(channel, environment, logger)
	assert.Equal(t, defaultReplicationFactor, actualReplicationFactor)

	// Test The Valid ReplicationFactor Use Case
	channel = &knativekafkav1alpha1.KafkaChannel{Spec: knativekafkav1alpha1.KafkaChannelSpec{ReplicationFactor: replicationFactor}}
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
	channel := &knativekafkav1alpha1.KafkaChannel{}
	actualRetentionMillis := RetentionMillis(channel, environment, logger)
	assert.Equal(t, defaultRetentionMillis, actualRetentionMillis)

	// Test The Valid RetentionMillis Use Case
	channel = &knativekafkav1alpha1.KafkaChannel{Spec: knativekafkav1alpha1.KafkaChannelSpec{RetentionMillis: retentionMillis}}
	actualRetentionMillis = RetentionMillis(channel, environment, logger)
	assert.Equal(t, retentionMillis, actualRetentionMillis)
}

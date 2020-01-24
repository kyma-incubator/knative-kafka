package util

import (
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/eventing/pkg/utils"
)

// Get A Logger With Channel Info
func ChannelLogger(logger *zap.Logger, channel *knativekafkav1alpha1.KafkaChannel) *zap.Logger {
	return logger.With(zap.String("Namespace", channel.Namespace), zap.String("Name", channel.Name))
}

// Create A New ControllerReference Model For The Specified Channel
func NewChannelControllerRef(channel *knativekafkav1alpha1.KafkaChannel) metav1.OwnerReference {
	return *metav1.NewControllerRef(channel, schema.GroupVersionKind{
		Group:   knativekafkav1alpha1.SchemeGroupVersion.Group,
		Version: knativekafkav1alpha1.SchemeGroupVersion.Version,
		Kind:    constants.KafkaChannelKind,
	})
}

// Create A Knative Reconciler "Key" Formatted Representation Of The Specified Channel
func ChannelKey(channel *knativekafkav1alpha1.KafkaChannel) string {
	return fmt.Sprintf("%s/%s", channel.Namespace, channel.Name)
}

// Create A DNS Safe Name For The Specified KafkaChannel Suitable For Use With K8S Services
func ChannelDnsSafeName(channel *knativekafkav1alpha1.KafkaChannel) string {

	// In order for the resulting name to be a valid DNS component is 63 characters.  We are appending 9 characters to separate
	// the components and to indicate this is a Channel which reduces the available length to 54.  We will allocate 30 characters
	// to the channel, and 20 to the namespace, leaving some extra buffer.
	safeChannelName := GenerateValidDnsName(channel.Name, 30)
	safeChannelNamespace := GenerateValidDnsName(channel.Namespace, 20)
	return fmt.Sprintf("%s-%s-channel", safeChannelName, safeChannelNamespace)
}

// Channel Host Naming Utility
func ChannelHostName(channelName, channelNamespace string) string {
	return fmt.Sprintf("%s.%s.channels.%s", channelName, channelNamespace, utils.GetClusterDomainName())
}

// Utility Function To Get The NumPartitions - First From Channel Spec And Then From Environment
func NumPartitions(channel *knativekafkav1alpha1.KafkaChannel, environment *env.Environment, logger *zap.Logger) int {
	value := channel.Spec.NumPartitions
	if value <= 0 {
		logger.Warn("Kafka Channel Spec 'NumPartitions' Not Specified - Using Default", zap.Int("Value", environment.DefaultNumPartitions))
		value = environment.DefaultNumPartitions
	}
	return value
}

// Utility Function To Get The ReplicationFactor - First From Channel Spec And Then From Environment
func ReplicationFactor(channel *knativekafkav1alpha1.KafkaChannel, environment *env.Environment, logger *zap.Logger) int {
	value := channel.Spec.ReplicationFactor
	if value <= 0 {
		logger.Warn("Kafka Channel Spec 'ReplicationFactor' Not Specified - Using Default", zap.Int("Value", environment.DefaultReplicationFactor))
		value = environment.DefaultReplicationFactor
	}
	return value
}

// Utility Function To Get The RetentionMillis - First From Channel Spec And Then From Environment
func RetentionMillis(channel *knativekafkav1alpha1.KafkaChannel, environment *env.Environment, logger *zap.Logger) int64 {
	value := channel.Spec.RetentionMillis
	if value <= 0 {
		logger.Warn("Kafka Channel Spec 'RetentionMillis' Not Specified - Using Default", zap.Int64("Value", environment.DefaultRetentionMillis))
		value = environment.DefaultRetentionMillis
	}
	return value
}

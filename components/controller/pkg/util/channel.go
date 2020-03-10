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

// Create A Knative Reconciler "Key" Formatted Representation Of The Specified Channel
func ChannelKey(channel *knativekafkav1alpha1.KafkaChannel) string {
	return fmt.Sprintf("%s/%s", channel.Namespace, channel.Name)
}

// Create A New OwnerReference For The Specified KafkaChannel (Controller)
func NewChannelOwnerReference(channel *knativekafkav1alpha1.KafkaChannel) metav1.OwnerReference {

	kafkaChannelGroupVersion := schema.GroupVersion{
		Group:   knativekafkav1alpha1.SchemeGroupVersion.Group,
		Version: knativekafkav1alpha1.SchemeGroupVersion.Version,
	}

	blockOwnerDeletion := true
	controller := true

	return metav1.OwnerReference{
		APIVersion:         kafkaChannelGroupVersion.String(),
		Kind:               constants.KafkaChannelKind,
		Name:               channel.GetName(),
		UID:                channel.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &controller,
	}
}

//
// Create A DNS Safe Name For The Channel Deployment Using The Specified Kafka Secret
//
// Note - The current implementation creates a single Channel Deployment for each
//        Kafka Authentication (K8S Secrets) instance.
//
func ChannelDnsSafeName(kafkaSecretName string) string {

	// In order for the resulting name to be a valid DNS component it's length must be no more than 63 characters.
	// We are consuming 9 chars for the component separators, and the Channel suffix, which reduces the available
	// length to 54. We will allocate 50 characters to the kafka secret name leaving an extra buffer.
	safeSecretName := GenerateValidDnsName(kafkaSecretName, 50, true, false)
	return fmt.Sprintf("%s-channel", safeSecretName)
}

// Channel Host Naming Utility
func ChannelHostName(channelName, channelNamespace string) string {
	return fmt.Sprintf("%s.%s.channels.%s", channelName, channelNamespace, utils.GetClusterDomainName())
}

// Utility Function To Get The NumPartitions - First From Channel Spec And Then From Environment
func NumPartitions(channel *knativekafkav1alpha1.KafkaChannel, environment *env.Environment, logger *zap.Logger) int {
	value := channel.Spec.NumPartitions
	if value <= 0 {
		logger.Debug("Kafka Channel Spec 'NumPartitions' Not Specified - Using Default", zap.Int("Value", environment.DefaultNumPartitions))
		value = environment.DefaultNumPartitions
	}
	return value
}

// Utility Function To Get The ReplicationFactor - First From Channel Spec And Then From Environment
func ReplicationFactor(channel *knativekafkav1alpha1.KafkaChannel, environment *env.Environment, logger *zap.Logger) int {
	value := channel.Spec.ReplicationFactor
	if value <= 0 {
		logger.Debug("Kafka Channel Spec 'ReplicationFactor' Not Specified - Using Default", zap.Int("Value", environment.DefaultReplicationFactor))
		value = environment.DefaultReplicationFactor
	}
	return value
}

// Utility Function To Get The RetentionMillis - First From Channel Spec And Then From Environment
func RetentionMillis(channel *knativekafkav1alpha1.KafkaChannel, environment *env.Environment, logger *zap.Logger) int64 {
	value := channel.Spec.RetentionMillis
	if value <= 0 {
		logger.Debug("Kafka Channel Spec 'RetentionMillis' Not Specified - Using Default", zap.Int64("Value", environment.DefaultRetentionMillis))
		value = environment.DefaultRetentionMillis
	}
	return value
}

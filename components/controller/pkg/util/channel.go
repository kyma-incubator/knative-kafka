package util

import (
	"fmt"
	"github.com/knative/eventing/pkg/utils"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Get A Logger With Channel Info
func ChannelLogger(logger *zap.Logger, channel *kafkav1alpha1.KafkaChannel) *zap.Logger {
	return logger.With(zap.String("Namespace", channel.Namespace), zap.String("Name", channel.Name))
}

// Create A New ControllerReference Model For The Specified Channel
func NewChannelControllerRef(channel *kafkav1alpha1.KafkaChannel) metav1.OwnerReference {
	return *metav1.NewControllerRef(channel, schema.GroupVersionKind{
		Group:   kafkav1alpha1.SchemeGroupVersion.Group,
		Version: kafkav1alpha1.SchemeGroupVersion.Version,
		Kind:    constants.KafkaChannelKind,
	})
}

//
// Knative Eventing Channel Utility Functions
//
// The following functions were copied from the Knative Eventing implementation and
// altered slightly for use in Knative-Kafka (knative/eventing/pkg/provisioners/channel_util.go).
// They are duplicated here because they've been made private and we're not using
// their public utilities for creating Services / VirtualServices etc.  We're also
// not using GeneratedNames as of yet.  We do however want to align with their
// naming convention as much as possible.
//

// Channel VirtualService Naming Utility
func ChannelVirtualServiceName(channelName string) string {
	return fmt.Sprintf("%s-channel", channelName)
}

// Channel Service Naming Utility
func ChannelServiceName(channelName string) string {
	return fmt.Sprintf("%s-channel", channelName)
}

// Channel Host Naming Utility
func ChannelHostName(channelName, channelNamespace string) string {
	return fmt.Sprintf("%s.%s.channels.%s", channelName, channelNamespace, utils.GetClusterDomainName())
}

// Utility Function To Get The NumPartitions - First From Channel Spec And Then From Environment
func NumPartitions(channel *kafkav1alpha1.KafkaChannel, environment *env.Environment, logger *zap.Logger) int {
	value := channel.Spec.NumPartitions
	if value <= 0 {
		logger.Warn("Kafka Channel Spec 'NumPartitions' Not Specified - Using Default", zap.Int("Value", environment.DefaultNumPartitions))
		value = environment.DefaultNumPartitions
	}
	return value
}

// Utility Function To Get The ReplicationFactor - First From Channel Spec And Then From Environment
func ReplicationFactor(channel *kafkav1alpha1.KafkaChannel, environment *env.Environment, logger *zap.Logger) int {
	value := channel.Spec.ReplicationFactor
	if value <= 0 {
		logger.Warn("Kafka Channel Spec 'ReplicationFactor' Not Specified - Using Default", zap.Int("Value", environment.DefaultReplicationFactor))
		value = environment.DefaultReplicationFactor
	}
	return value
}

// Utility Function To Get The RetentionMillis - First From Channel Spec And Then From Environment
func RetentionMillis(channel *kafkav1alpha1.KafkaChannel, environment *env.Environment, logger *zap.Logger) int64 {
	value := channel.Spec.RetentionMillis
	if value <= 0 {
		logger.Warn("Kafka Channel Spec 'RetentionMillis' Not Specified - Using Default", zap.Int64("Value", environment.DefaultRetentionMillis))
		value = environment.DefaultRetentionMillis
	}
	return value
}

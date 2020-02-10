package util

import (
	commonkafkautil "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/util"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
)

// Get The TopicName For Specified Channel (ChannelNamespace.ChannelName)
func TopicName(channel *kafkav1alpha1.KafkaChannel) string {
	return commonkafkautil.TopicName(channel.Namespace, channel.Name)
}

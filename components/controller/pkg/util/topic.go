package util

import (
	"fmt"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
)

// Get The TopicName For Specified Channel (TenantId.ChannelNamespace.ChannelName)
func TopicName(channel *kafkav1alpha1.KafkaChannel, environment *env.Environment) string {

	// Get The Topic Name Components For Specified Channel
	tenantId := channel.Spec.TenantId
	channelNamespace := channel.Namespace
	channelName := channel.Name

	// Default The Tenant Id
	if len(tenantId) <= 0 {
		tenantId = environment.DefaultTenantId
	}

	// Format The Topic Name And Return
	return fmt.Sprintf("%s.%s.%s", tenantId, channelNamespace, channelName)
}

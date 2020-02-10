package test

import (
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingChannel "knative.dev/eventing/pkg/channel"
	knativeapis "knative.dev/pkg/apis"
	apiduckv1 "knative.dev/pkg/apis/duck/v1"
)

// Utility Function For Creating A Test ChannelReference (Knative)
func CreateChannelReference(name string, namespace string) eventingChannel.ChannelReference {
	return eventingChannel.ChannelReference{
		Name:      name,
		Namespace: namespace,
	}
}

// Utility Function For Creating A Test KafkaChannel (Knative-Kafka)
func CreateKafkaChannel(name string, namespace string, ready corev1.ConditionStatus) *knativekafkav1alpha1.KafkaChannel {
	return &knativekafkav1alpha1.KafkaChannel{
		TypeMeta: v1.TypeMeta{
			Kind:       constants.KafkaChannelKind,
			APIVersion: knativekafkav1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: knativekafkav1alpha1.KafkaChannelSpec{
			TenantId:          "",
			NumPartitions:     0,
			ReplicationFactor: 0,
			RetentionMillis:   0,
			Subscribable:      nil,
		},
		Status: knativekafkav1alpha1.KafkaChannelStatus{
			Status: apiduckv1.Status{
				Conditions: []knativeapis.Condition{
					{Type: knativeapis.ConditionReady, Status: ready},
				},
			},
		},
	}
}

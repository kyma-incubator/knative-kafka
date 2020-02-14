package testing

import (
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/eventing/pkg/apis/duck/v1beta1"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

// KafkaChannelOption enables further configuration of a KafkaChannel.
type KafkaChannelOption func(*v1alpha1.KafkaChannel)

// NewKafkaChannel creates an KafkaChannel with KafkaChannelOptions.
func NewKafkaChannel(name, namespace string, ncopt ...KafkaChannelOption) *v1alpha1.KafkaChannel {
	nc := &v1alpha1.KafkaChannel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.KafkaChannelSpec{},
	}
	for _, opt := range ncopt {
		opt(nc)
	}
	return nc
}

func WithInitKafkaChannelConditions(nc *v1alpha1.KafkaChannel) {
	nc.Status.InitializeConditions()
}

func WithKafkaChannelReady(nc *v1alpha1.KafkaChannel) {
	nc.Status.MarkTopicTrue()
	nc.Status.MarkChannelServiceTrue()
	nc.Status.MarkChannelDeploymentTrue()
	nc.Status.MarkChannelDeploymentServiceTrue()
	nc.Status.MarkDispatcherDeploymentTrue()
}

func WithKafkaChannelDeleted(nc *v1alpha1.KafkaChannel) {
	deleteTime := metav1.NewTime(time.Unix(1e9, 0))
	nc.ObjectMeta.SetDeletionTimestamp(&deleteTime)
}

func WithKafkaChannelAddress(a string) KafkaChannelOption {
	return func(nc *v1alpha1.KafkaChannel) {
		nc.Status.SetAddress(&apis.URL{
			Scheme: "http",
			Host:   a,
		})
	}
}

func WithSubscriber(uid types.UID, uri string) KafkaChannelOption {
	return func(nc *v1alpha1.KafkaChannel) {
		if nc.Spec.Subscribers == nil {
			nc.Spec.Subscribers = []v1beta1.SubscriberSpec{}
		}

		url, _ := apis.ParseURL(uri)
		nc.Spec.Subscribers = append(nc.Spec.Subscribers, v1beta1.SubscriberSpec{
			UID:           uid,
			SubscriberURI: url,
		})
	}
}

func WithSubscriberReady(uid types.UID) KafkaChannelOption {
	return func(nc *v1alpha1.KafkaChannel) {
		if nc.Status.Subscribers == nil {
			nc.Status.Subscribers = []v1beta1.SubscriberStatus{}
		}

		nc.Status.SubscribableStatus.Subscribers = append(nc.Status.SubscribableStatus.Subscribers, v1beta1.SubscriberStatus{
			Ready: v1.ConditionTrue,
			UID:   uid,
		})
	}
}

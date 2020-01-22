package testing

import (
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	v1alpha12 "knative.dev/eventing/pkg/apis/duck/v1alpha1"
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
	nc.Status.MarkChannelServiceTrue()
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
		if nc.Spec.Subscribable == nil {
			nc.Spec.Subscribable = &v1alpha12.Subscribable{}
		}

		nc.Spec.Subscribable.Subscribers = append(nc.Spec.Subscribable.Subscribers, v1alpha12.SubscriberSpec{
			UID:           uid,
			SubscriberURI: uri,
		})
	}
}

func WithSubscriberReady(uid types.UID) KafkaChannelOption {
	return func(nc *v1alpha1.KafkaChannel) {
		if nc.Status.SubscribableStatus == nil {
			nc.Status.SubscribableStatus = &v1alpha12.SubscribableStatus{}
		}

		nc.Status.SubscribableStatus.Subscribers = append(nc.Status.SubscribableStatus.Subscribers, v1alpha12.SubscriberStatus{
			Ready: v1.ConditionTrue,
			UID:   uid,
		})
	}
}

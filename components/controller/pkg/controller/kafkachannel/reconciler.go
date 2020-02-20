package kafkachannel

import (
	"context"
	"fmt"
	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	kafkachannelreconciler "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkachannel"
	listers "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/listers/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	eventingreconciler "knative.dev/eventing/pkg/reconciler"
)

// Reconciler implements controller.Reconciler for KafkaChannel resources.
type Reconciler struct {
	*eventingreconciler.Base

	adminClient          kafkaadmin.AdminClientInterface
	environment          *env.Environment
	kafkachannelLister   listers.KafkaChannelLister
	kafkachannelInformer cache.SharedIndexInformer
	deploymentLister     appsv1listers.DeploymentLister
	serviceLister        corev1listers.ServiceLister
}

// Check that our Reconciler implements Interface
var _ kafkachannelreconciler.Interface = (*Reconciler)(nil)

// Perform The Actual Channel Reconciliation (In Parallel)
func (r *Reconciler) reconcile(ctx context.Context, channel *kafkav1alpha1.KafkaChannel) error {

	// NOTE - The sequential order of reconciliation must be "Topic" then "Channel / Dispatcher" in order for the
	//        EventHub Cache to know the dynamically determined EventHub Namespace / Kafka Secret selected for the topic.

	// Reconcile The KafkaChannel's Kafka Topic
	err := r.reconcileTopic(ctx, channel)
	if err != nil {
		return fmt.Errorf("reconciliation failed")
	}

	// Reconcile The KafkaChannel's Channel & Dispatcher Deployment/Service
	channelError := r.reconcileChannel(channel)
	dispatcherError := r.reconcileDispatcher(channel)
	if channelError != nil || dispatcherError != nil {
		return fmt.Errorf("reconciliation failed")
	}

	// Return Success
	return nil
}

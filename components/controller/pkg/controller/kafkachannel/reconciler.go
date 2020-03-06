package kafkachannel

import (
	"context"
	"fmt"
	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	knativekafkaclientset "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned"
	kafkachannelreconciler "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkachannel"
	knativekafkalisters "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/listers/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	eventingreconciler "knative.dev/eventing/pkg/reconciler"
)

// Reconciler Implements controller.Reconciler for KafkaChannel Resources
type Reconciler struct {
	*eventingreconciler.Base
	knativekafkaClientSet knativekafkaclientset.Interface
	adminClient           kafkaadmin.AdminClientInterface
	environment           *env.Environment
	kafkachannelLister    knativekafkalisters.KafkaChannelLister
	kafkachannelInformer  cache.SharedIndexInformer
	deploymentLister      appsv1listers.DeploymentLister
	serviceLister         corev1listers.ServiceLister
}

// Verify The KafkaChannel Reconciler Implements It's Interface
var _ kafkachannelreconciler.Interface = (*Reconciler)(nil)

// Perform The Actual Channel Reconciliation
func (r *Reconciler) reconcile(ctx context.Context, channel *kafkav1alpha1.KafkaChannel) error {

	// NOTE - The sequential order of reconciliation must be "Topic" then "Channel / Dispatcher" in order for the
	//        EventHub Cache to know the dynamically determined EventHub Namespace / Kafka Secret selected for the topic.

	// Reconcile The KafkaChannel's Kafka Topic
	err := r.reconcileTopic(ctx, channel)
	if err != nil {
		return fmt.Errorf(constants.ReconciliationFailedError)
	}

	// Reconcile The KafkaChannel's Channel & Dispatcher Deployment/Service
	channelError := r.reconcileChannel(channel)
	dispatcherError := r.reconcileDispatcher(channel)
	if channelError != nil || dispatcherError != nil {
		return fmt.Errorf(constants.ReconciliationFailedError)
	}

	// Add Labels To KafkaChannel
	err = r.reconcileKafkaChannelLabels(channel)
	if err != nil {
		return fmt.Errorf(constants.ReconciliationFailedError)
	}

	// Return Success
	return nil
}

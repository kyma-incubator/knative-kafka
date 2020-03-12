package kafkachannel

import (
	"context"
	"fmt"
	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	knativekafkaclientset "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkachannel"
	knativekafkalisters "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/listers/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/util"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	eventingreconciler "knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/reconciler"
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

var (
	_ kafkachannel.Interface = (*Reconciler)(nil) // Verify Reconciler Implements Interface
	_ kafkachannel.Finalizer = (*Reconciler)(nil) // Verify Reconciler Implements Finalizer
)

// ReconcileKind Implements The Reconciler Interface & Is Responsible For Performing The Reconciliation (Creation)
func (r *Reconciler) ReconcileKind(ctx context.Context, channel *knativekafkav1alpha1.KafkaChannel) reconciler.Event {

	r.Logger.Debug("<==========  START KAFKA-CHANNEL RECONCILIATION  ==========>")

	// Reset The Channel's Status Conditions To Unknown (Addressable, Topic, Service, Deployment, etc...)
	channel.Status.InitializeConditions()

	// Perform The KafkaChannel Reconciliation & Handle Error Response
	r.Logger.Info("Channel Owned By Controller - Reconciling", zap.Any("Channel.Spec", channel.Spec))
	err := r.reconcile(ctx, channel)
	if err != nil {
		r.Logger.Error("Failed To Reconcile KafkaChannel", zap.Any("Channel", channel), zap.Error(err))
		return err
	}

	// Return Success
	r.Logger.Info("Successfully Reconciled KafkaChannel", zap.Any("Channel", channel))
	channel.Status.ObservedGeneration = channel.Generation
	return reconciler.NewEvent(corev1.EventTypeNormal, event.KafkaChannelReconciled.String(), "KafkaChannel Reconciled Successfully: \"%s/%s\"", channel.Namespace, channel.Name)
}

// ReconcileKind Implements The Finalizer Interface & Is Responsible For Performing The Finalization (Topic Deletion)
func (r *Reconciler) FinalizeKind(ctx context.Context, channel *knativekafkav1alpha1.KafkaChannel) reconciler.Event {

	r.Logger.Debug("<==========  START KAFKA-CHANNEL FINALIZATION  ==========>")

	// Get The Kafka Topic Name For Specified Channel
	topicName := util.TopicName(channel)

	// Delete The Kafka Topic & Handle Error Response
	err := r.deleteTopic(ctx, topicName)
	if err != nil {
		r.Logger.Error("Failed To Finalize KafkaChannel", zap.Any("Channel", channel), zap.Error(err))
		return err
	}

	// Return Success
	r.Logger.Info("Successfully Finalized KafkaChannel", zap.Any("Channel", channel))
	return reconciler.NewEvent(corev1.EventTypeNormal, event.KafkaChannelFinalized.String(), "KafkaChannel Finalized Successfully: \"%s/%s\"", channel.Namespace, channel.Name)
}

// Perform The Actual Channel Reconciliation
func (r *Reconciler) reconcile(ctx context.Context, channel *knativekafkav1alpha1.KafkaChannel) error {

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
	err = r.reconcileKafkaChannel(channel)
	if err != nil {
		return fmt.Errorf(constants.ReconciliationFailedError)
	}

	// Return Success
	return nil
}

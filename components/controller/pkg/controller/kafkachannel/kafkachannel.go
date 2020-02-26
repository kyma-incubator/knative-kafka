package kafkachannel

import (
	"context"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkachannel"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/util"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/reconciler"
)

var (
	_ kafkachannel.Interface = (*Reconciler)(nil) // Verify Reconciler Implements Interface
	_ kafkachannel.Finalizer = (*Reconciler)(nil) // Verify Reconciler Implements Finalizer
)

// ReconcileKind Implements The Reconciler Interface & Is Responsible For Performing The Reconciliation (Creation)
func (r *Reconciler) ReconcileKind(ctx context.Context, kafkachannel *v1alpha1.KafkaChannel) reconciler.Event {

	r.Logger.Debug("<==========  START CHANNEL RECONCILIATION  ==========>")

	// Reset The Channel's Status Conditions To Unknown (Addressable, Topic, Service, Deployment, etc...)
	kafkachannel.Status.InitializeConditions()

	// Perform The Channel Reconciliation & Handle Error Response
	r.Logger.Info("Channel Owned By Controller - Reconciling", zap.Any("Channel.Spec", kafkachannel.Spec))
	err := r.reconcile(ctx, kafkachannel)
	if err != nil {
		r.Logger.Error("Failed To Reconcile KafkaChannel", zap.Any("Channel", kafkachannel), zap.Error(err))
		return err
	}

	// Return Success
	r.Logger.Info("Successfully Reconciled KafkaChannel", zap.Any("Channel", kafkachannel))
	kafkachannel.Status.ObservedGeneration = kafkachannel.Generation
	return reconciler.NewEvent(corev1.EventTypeNormal, event.KafkaChannelReconciled.String(), "KafkaChannel Reconciled Successfully: \"%s/%s\"", kafkachannel.Namespace, kafkachannel.Name)
}

// ReconcileKind Implements The Finalizer Interface & Is Responsible For Performing The Finalization (Deletion)
func (r *Reconciler) FinalizeKind(ctx context.Context, kafkachannel *v1alpha1.KafkaChannel) reconciler.Event {

	// Get The Kafka Topic Name For Specified Channel
	topicName := util.TopicName(kafkachannel)

	// Delete The Kafka Topic & Handle Error Response
	err := r.deleteTopic(ctx, topicName)
	if err != nil {
		r.Logger.Error("Failed To Finalize KafkaChannel", zap.Any("Channel", kafkachannel), zap.Error(err))
		return err
	}

	// Return Success
	r.Logger.Info("Successfully Finalized KafkaChannel", zap.Any("Channel", kafkachannel))
	return reconciler.NewEvent(corev1.EventTypeNormal, event.KafkaChannelFinalized.String(), "KafkaChannel Finalized Successfully: \"%s/%s\"", kafkachannel.Namespace, kafkachannel.Name)
}

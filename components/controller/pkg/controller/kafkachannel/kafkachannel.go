package kafkachannel

import (
	"context"

	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkachannel"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/util"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/reconciler"
)

// newReconciledNormal makes a new reconciler event with event type Normal, and reason KafkaChannelReconciled.
func newReconciledNormal(namespace, name string) reconciler.Event {
	return reconciler.NewEvent(v1.EventTypeNormal, "KafkaChannelReconciled", "KafkaChannel reconciled: \"%s/%s\"", namespace, name)
}

// Check that our Reconciler implements Interface
var _ kafkachannel.Interface = (*Reconciler)(nil)

// Optionally check that our Reconciler implements Finalizer
var _ kafkachannel.Finalizer = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, kafkachannel *v1alpha1.KafkaChannel) reconciler.Event {

	r.Logger.Debug("<==========  START CHANNEL RECONCILIATION  ==========>")

	// Reset The Channel's Status Conditions To Unknown (Addressable, Topic, Service, Deployment, etc...)
	kafkachannel.Status.InitializeConditions()

	// Perform The Channel Reconciliation
	r.Logger.Info("Channel Owned By Controller - Reconciling", zap.Any("Channel.Spec", kafkachannel.Spec))
	err := r.reconcile(ctx, kafkachannel)
	if err != nil {
		r.Logger.Error("Failed To Reconcile Channel", zap.Error(err))
		// Note: Do NOT Return Error Here In Order To Ensure Status Update
	} else {
		r.Logger.Info("Successfully Reconciled Channel", zap.Any("Channel", kafkachannel))
	}

	kafkachannel.Status.ObservedGeneration = kafkachannel.Generation
	return newReconciledNormal(kafkachannel.Namespace, kafkachannel.Name)
}

func (r *Reconciler) FinalizeKind(ctx context.Context, kafkachannel *v1alpha1.KafkaChannel) reconciler.Event {
	// Get The TopicName For Specified Channel
	topicName := util.TopicName(kafkachannel)
	err := r.deleteTopic(ctx, topicName)
	if err != nil {
		return err
	}
	return reconciler.NewEvent(v1.EventTypeNormal, "KafkaChannel Finalized", "Topic %s deleted during finalization", topicName)
}

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

// ReconcileKind Implements The Reconciler Interface & Is Responsible For Performing The Reconciliation (Creation)
func (r *Reconciler) ReconcileKind(ctx context.Context, kafkachannel *v1alpha1.KafkaChannel) reconciler.Event {

	r.Logger.Debug("<==========  START CHANNEL RECONCILIATION  ==========>")

	// Reset The Channel's Status Conditions To Unknown (Addressable, Topic, Service, Deployment, etc...)
	kafkachannel.Status.InitializeConditions()

	// Perform The Channel Reconciliation
	r.Logger.Info("Channel Owned By Controller - Reconciling", zap.Any("Channel.Spec", kafkachannel.Spec))
	err := r.reconcile(ctx, kafkachannel)
	if err != nil {
		r.Logger.Error("Failed To Reconcile KafkaChannel", zap.Any("Channel", kafkachannel), zap.Error(err))
		// TODO - should we return errors? or log an event or something ??? - look at the eventing-contrib implementations to see???
	} else {
		r.Logger.Info("Successfully Reconciled KafkaChannel", zap.Any("Channel", kafkachannel))
	}

	kafkachannel.Status.ObservedGeneration = kafkachannel.Generation
	return newReconciledNormal(kafkachannel.Namespace, kafkachannel.Name)
}

// ReconcileKind Implements The Finalizer Interface & Is Responsible For Performing The Finalization (Deletion)
func (r *Reconciler) FinalizeKind(ctx context.Context, kafkachannel *v1alpha1.KafkaChannel) reconciler.Event {

	// Get The Kafka Topic Name For Specified Channel
	topicName := util.TopicName(kafkachannel)

	// Delete The Kafka Topic
	err := r.deleteTopic(ctx, topicName)
	if err != nil {
		// TODO - create event here??? same as above - what's the standard
		r.Logger.Error("Failed To Finalize KafkaChannel", zap.Any("Channel", kafkachannel), zap.Error(err))
		return err
	} else {
		r.Logger.Info("Successfully Finalized KafkaChannel", zap.Any("Channel", kafkachannel))
		// TODO - convert this event creation to our pattern and maybe create event functions to simplify?  and/or look at the type of events we should be creating now ?
		// TODO - they have their own reconciler event struct for returning here - our other stuff logs/creates all in one (using the client-go stuff, etc...   sigh
		return reconciler.NewEvent(v1.EventTypeNormal, "KafkaChannel Finalized", "Topic %s deleted during finalization", topicName)
		// TODO - r.Recorder.Eventf(channel, corev1.EventTypeWarning, event.ChannelServiceReconciliationFailed.String(), "Failed To Reconcile KafkaChannel Service For Channel: %v", channelServiceErr)
	}
}

package kafkachannel

import (
	"context"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
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

// TODO - do these belong here or in the reconcile.go ?  would like to move them ; )

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

// Add Labels To KafkaChannel (Call After KafkaChannel Reconciliation !)
func (r *Reconciler) reconcileKafkaChannelLabels(channel *knativekafkav1alpha1.KafkaChannel) error {

	// Get The KafkaChannel's Current Labels
	labels := channel.Labels

	// Initialize The Labels Map If Empty
	if labels == nil {
		labels = make(map[string]string)
	}

	// Track Modified Status
	modified := false

	// Add Kafka Topic Label If Missing
	topicName := util.TopicName(channel)
	if labels[constants.KafkaTopicLabel] != topicName {
		labels[constants.KafkaTopicLabel] = topicName
		modified = true
	}

	// Add Kafka Secret Label If Missing
	secretName := r.kafkaSecretName(channel) // Can Only Be Called AFTER Topic Reconciliation !!!
	if labels[constants.KafkaSecretLabel] != secretName {
		labels[constants.KafkaSecretLabel] = secretName
		modified = true
	}

	// If The KafkaChannel's Labels Were Modified
	if modified {

		// Then Update The KafkaChannel's Labels
		channel.Labels = labels
		_, err := r.knativekafkaClientSet.KnativekafkaV1alpha1().KafkaChannels(channel.Namespace).Update(channel)
		if err != nil {
			r.Logger.Error("Failed To Update KafkaChannel Labels", zap.Error(err))
			return err
		} else {
			r.Logger.Info("Successfully Updated KafkaChannel Labels")
			return nil
		}
	} else {

		// Otherwise Nothing To Do
		r.Logger.Info("Successfully Verified KafkaChannel Labels")
		return nil
	}
}

package kafkasecret

import (
	"context"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkasecret"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/reconciler"
)

var (
	_ kafkasecret.Interface = (*Reconciler)(nil) // Verify Reconciler Implements Interface
	_ kafkasecret.Finalizer = (*Reconciler)(nil) // Verify Reconciler Implements Finalizer
)

// ReconcileKind Implements The Reconciler Interface & Is Responsible For Performing The Reconciliation (Creation)
func (r *Reconciler) ReconcileKind(ctx context.Context, secret *corev1.Secret) reconciler.Event {

	// Setup Logger & Debug Log Separator
	r.Logger.Debug("<==========  START KAFKA-SECRET RECONCILIATION  ==========>")
	logger := r.Logger.With(zap.String("Secret", secret.Name))

	// Perform The Secret Reconciliation & Handle Error Response
	logger.Info("Secret Owned By Controller - Reconciling", zap.String("Secret", secret.Name))
	err := r.reconcile(ctx, secret)
	if err != nil {
		logger.Error("Failed To Reconcile Kafka Secret", zap.String("Secret", secret.Name), zap.Error(err))
		return err
	}

	// Return Success
	logger.Info("Successfully Reconciled Kafka Secret", zap.Any("Channel", secret.Name))
	return reconciler.NewEvent(corev1.EventTypeNormal, event.KafkaSecretReconciled.String(), "Kafka Secret Reconciled Successfully: \"%s/%s\"", secret.Namespace, secret.Name)
}

// ReconcileKind Implements The Finalizer Interface & Is Responsible For Performing The Finalization (KafkaChannel Status)
func (r *Reconciler) FinalizeKind(ctx context.Context, secret *corev1.Secret) reconciler.Event {

	// Setup Logger & Debug Log Separator
	r.Logger.Debug("<==========  START KAFKA-SECRET FINALIZATION  ==========>")
	logger := r.Logger.With(zap.String("Secret", secret.Name))

	// Reconcile The Affected KafkaChannel Status To Indicate The Channel Service/Deployment Are Not Longer Available
	err := r.reconcileKafkaChannelStatus(secret, false, false)
	if err != nil {
		logger.Error("Failed To Finalize Kafka Secret - KafkaChannel Status Update Failed", zap.Error(err))
		return err
	}

	// Return Success
	logger.Info("Successfully Finalized Kafka Secret", zap.Any("Secret", secret.Name))
	return reconciler.NewEvent(corev1.EventTypeNormal, event.KafkaSecretFinalized.String(), "Kafka Secret Finalized Successfully: \"%s/%s\"", secret.Namespace, secret.Name)
}

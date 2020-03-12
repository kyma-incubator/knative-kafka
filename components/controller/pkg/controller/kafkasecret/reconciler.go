package kafkasecret

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkasecret"
	knativekafkalisters "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/listers/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	eventingreconciler "knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/reconciler"
)

// Reconciler Implements controller.Reconciler for K8S Secrets Containing Kafka Auth (Labelled)
type Reconciler struct {
	*eventingreconciler.Base
	environment        *env.Environment
	kafkaChannelClient versioned.Interface
	kafkachannelLister knativekafkalisters.KafkaChannelLister
	deploymentLister   appsv1listers.DeploymentLister
	serviceLister      corev1listers.ServiceLister
}

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
	err := r.reconcileKafkaChannelStatus(secret,
		false, "ChannelServiceUnavailable", "Kafka Auth Secret Finalized",
		false, "ChannelDeploymentUnavailable", "Kafka Auth Secret Finalized")
	if err != nil {
		logger.Error("Failed To Finalize Kafka Secret - KafkaChannel Status Update Failed", zap.Error(err))
		return err
	}

	// Return Success
	logger.Info("Successfully Finalized Kafka Secret", zap.Any("Secret", secret.Name))
	return reconciler.NewEvent(corev1.EventTypeNormal, event.KafkaSecretFinalized.String(), "Kafka Secret Finalized Successfully: \"%s/%s\"", secret.Namespace, secret.Name)
}

// Perform The Actual Secret Reconciliation
func (r *Reconciler) reconcile(ctx context.Context, secret *corev1.Secret) error {

	// Perform The Kafka Secret Reconciliation
	err := r.reconcileChannel(secret)
	if err != nil {
		return fmt.Errorf(constants.ReconciliationFailedError)
	}

	// Return Success
	return nil
}

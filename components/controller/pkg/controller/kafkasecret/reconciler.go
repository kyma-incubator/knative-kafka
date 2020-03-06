package kafkasecret

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned"
	kafkasecretreconciler "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkasecret"
	knativekafkalisters "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/listers/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	corev1 "k8s.io/api/core/v1"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	eventingreconciler "knative.dev/eventing/pkg/reconciler"
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

// Verify The KafkaSecret Reconciler Implements It's Interface
var _ kafkasecretreconciler.Interface = (*Reconciler)(nil)

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

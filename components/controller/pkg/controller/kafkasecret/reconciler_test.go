package kafkasecret

import (
	"context"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	fakeknativekafkaclient "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/client/fake"
	kafkasecretreconciler "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkasecret"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/eventing/pkg/reconciler"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	logtesting "knative.dev/pkg/logging/testing"
	. "knative.dev/pkg/reconciler/testing"
	"testing"
)

// Initialization - Add types to scheme
func init() {
	_ = knativekafkav1alpha1.AddToScheme(scheme.Scheme)
	_ = duckv1alpha1.AddToScheme(scheme.Scheme)
}

// Test The Reconcile Functionality
func TestReconcile(t *testing.T) {

	// Clear Logs After Testing
	defer logtesting.ClearAll()

	//
	// Define The KafkaChannel Reconciler Test Cases
	//
	// Note - Knative testing framework assumes ALL actions will be in the same Namespace
	//        as the Key so we have to set SkipNamespaceValidation in all tests!
	//
	// Note - Knative reconciler framework expects Events (not errors) from ReconcileKind()
	//        so WantErr is only for higher level failures in the injected Reconcile() function.
	//
	tableTest := TableTest{

		//
		// Top Level Use Cases
		//

		{
			Name: "Bad Secret Key",
			Key:  "too/many/parts",
		},
		{
			Name: "Secret Key Not Found",
			Key:  "foo/not-found",
		},

		//
		// Full Reconciliation
		//

		{
			Name: "Complete Reconciliation Without KafkaChannel",
			Key:  test.KafkaSecretKey,
			Objects: []runtime.Object{
				test.NewKafkaSecret(),
			},
			WantCreates: []runtime.Object{
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelChannelDeployment(),
			},
			WantPatches: []clientgotesting.PatchActionImpl{test.NewKafkaSecretFinalizerPatchActionImpl()},
			WantEvents: []string{
				test.NewKafkaSecretFinalizerUpdateEvent(),
				test.NewKafkaSecretSuccessfulReconciliationEvent(),
			},
		},
		{
			Name: "Complete Reconciliation With KafkaChannel",
			Key:  test.KafkaSecretKey,
			Objects: []runtime.Object{
				test.NewKafkaSecret(),
				test.NewKnativeKafkaChannel(),
			},
			WantCreates: []runtime.Object{
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelChannelDeployment(),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: test.NewKnativeKafkaChannel(
						test.WithKafkaChannelChannelServiceReady,
						test.WithKafkaChannelChannelDeploymentReady,
					),
				},
			},
			WantPatches: []clientgotesting.PatchActionImpl{test.NewKafkaSecretFinalizerPatchActionImpl()},
			WantEvents: []string{
				test.NewKafkaSecretFinalizerUpdateEvent(),
				test.NewKafkaSecretSuccessfulReconciliationEvent(),
			},
		},

		//
		// KafkaChannel Secret Deletion (Finalizer)
		//

		{
			Name: "Finalize Deleted Secret",
			Key:  test.KafkaSecretKey,
			Objects: []runtime.Object{
				test.NewKafkaSecret(test.WithKafkaSecretDeleted),
				test.NewKnativeKafkaChannel(
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
				),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: test.NewKnativeKafkaChannel(
						test.WithKafkaChannelChannelServiceFinalized,
						test.WithKafkaChannelChannelDeploymentFinalized,
					),
				},
			},
			WantEvents: []string{
				test.NewKafkaSecretSuccessfulFinalizedEvent(),
			},
		},

		//
		// KafkaChannel Channel Service
		//

		{
			Name: "Reconcile Missing Channel Service Success",
			Key:  test.KafkaSecretKey,
			Objects: []runtime.Object{
				test.NewKafkaSecret(test.WithKafkaSecretFinalizer),
				test.NewKnativeKafkaChannel(
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
				),
				test.NewKafkaChannelChannelDeployment(),
			},
			WantCreates: []runtime.Object{test.NewKafkaChannelChannelService()},
			WantEvents: []string{
				test.NewKafkaSecretSuccessfulReconciliationEvent(),
			},
		},
		{
			Name: "Reconcile Missing Channel Service Error(Create)",
			Key:  test.KafkaSecretKey,
			Objects: []runtime.Object{
				test.NewKafkaSecret(test.WithKafkaSecretFinalizer),
				test.NewKnativeKafkaChannel(
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
				),
				test.NewKafkaChannelService(),
				test.NewKafkaChannelChannelDeployment(),
				test.NewKafkaChannelDispatcherService(),
				test.NewKafkaChannelDispatcherDeployment(),
			},
			WithReactors: []clientgotesting.ReactionFunc{InduceFailure("create", "services")},
			WantErr:      true,
			WantCreates:  []runtime.Object{test.NewKafkaChannelChannelService()},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: test.NewKnativeKafkaChannel(
						test.WithKafkaChannelChannelServiceFailed,
						test.WithKafkaChannelChannelDeploymentReady,
					),
				},
			},
			WantEvents: []string{
				Eventf(corev1.EventTypeWarning, event.ChannelServiceReconciliationFailed.String(), "Failed To Reconcile Channel Service: inducing failure for create services"),
				test.NewKafkaSecretFailedReconciliationEvent(),
			},
		},

		//
		// KafkaChannel Channel Deployment
		//

		{
			Name: "Reconcile Missing Channel Deployment Success",
			Key:  test.KafkaSecretKey,
			Objects: []runtime.Object{
				test.NewKafkaSecret(test.WithKafkaSecretFinalizer),
				test.NewKnativeKafkaChannel(
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
				),
				test.NewKafkaChannelService(),
				test.NewKafkaChannelChannelService(),
			},
			WantCreates: []runtime.Object{test.NewKafkaChannelChannelDeployment()},
			WantEvents:  []string{test.NewKafkaSecretSuccessfulReconciliationEvent()},
		},
		{
			Name: "Reconcile Missing Channel Deployment Error(Create)",
			Key:  test.KafkaSecretKey,
			Objects: []runtime.Object{
				test.NewKafkaSecret(test.WithKafkaSecretFinalizer),
				test.NewKnativeKafkaChannel(
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
				),
				test.NewKafkaChannelService(),
				test.NewKafkaChannelChannelService(),
			},
			WithReactors: []clientgotesting.ReactionFunc{InduceFailure("create", "deployments")},
			WantErr:      true,
			WantCreates:  []runtime.Object{test.NewKafkaChannelChannelDeployment()},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: test.NewKnativeKafkaChannel(
						test.WithKafkaChannelChannelServiceReady,
						test.WithKafkaChannelChannelDeploymentFailed,
					),
				},
			},
			WantEvents: []string{
				Eventf(corev1.EventTypeWarning, event.ChannelDeploymentReconciliationFailed.String(), "Failed To Reconcile Channel Deployment: inducing failure for create deployments"),
				test.NewKafkaSecretFailedReconciliationEvent(),
			},
		},
	}

	// Run The TableTest Using The KafkaChannel Reconciler Provided By The Factory
	logger := logtesting.TestLogger(t)
	tableTest.Test(t, test.MakeFactory(func(ctx context.Context, listers *test.Listers, cmw configmap.Watcher) controller.Reconciler {
		r := &Reconciler{
			Base:               reconciler.NewBase(ctx, constants.KafkaChannelControllerAgentName, cmw),
			environment:        test.NewEnvironment(),
			kafkaChannelClient: fakeknativekafkaclient.Get(ctx),
			kafkachannelLister: listers.GetKafkaChannelLister(),
			deploymentLister:   listers.GetDeploymentLister(),
			serviceLister:      listers.GetServiceLister(),
		}
		return kafkasecretreconciler.NewReconciler(ctx, r.Logger, r.KubeClientSet.CoreV1(), listers.GetSecretLister(), r.Recorder, r)
	}, logger.Desugar()))
}

package kafkachannel

import (
	"context"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	fakeknativekafkaclient "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/client/fake"
	kafkachannelreconciler "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkachannel"
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

// Test Initialization - Add types to scheme
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
			Name: "Bad KafkaChannel Key",
			Key:  "too/many/parts",
		},
		{
			Name: "KafkaChannel Key Not Found",
			Key:  "foo/not-found",
		},
		{
			Name: "Delete KafkaChannel",
			Key:  test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(test.WithInitKafkaChannelConditions, test.WithKafkaChannelDeleted),
			},
			WantEvents: []string{
				Eventf(corev1.EventTypeNormal, event.KafkaChannelFinalized.String(), "KafkaChannel Finalized Successfully: \"%s/%s\"", test.KafkaChannelNamespace, test.KafkaChannelName),
			},
		},

		//
		// Full Reconciliation
		//

		{
			Name:                    "Complete Reconciliation Success",
			SkipNamespaceValidation: true,
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(test.WithInitKafkaChannelConditions),
			},
			WantCreates: []runtime.Object{
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelDeploymentService(),
				test.NewKafkaChannelChannelDeployment(),
				test.NewKafkaChannelDispatcherService(),
				test.NewKafkaChannelDispatcherDeployment(),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: test.NewKnativeKafkaChannel(
						test.WithKafkaChannelAddress,
						test.WithInitKafkaChannelConditions,
						test.WithKafkaChannelChannelServiceReady,
						test.WithKafkaChannelDeploymentServiceReady,
						test.WithKafkaChannelChannelDeploymentReady,
						test.WithKafkaChannelDispatcherDeploymentReady,
						test.WithKafkaChannelTopicReady,
					),
				},
			},
			WantPatches: []clientgotesting.PatchActionImpl{test.NewFinalizerPatchActionImpl()},
			WantEvents: []string{
				Eventf(corev1.EventTypeNormal, "FinalizerUpdate", `Updated "%s" finalizers`, test.KafkaChannelName),
				test.NewKafkaChannelSuccessfulReconciliationEvent(),
			},
		},

		//
		// KafkaChannel Channel Channel Service
		//

		{
			Name:                    "Reconcile Missing Channel Channel Service Success",
			SkipNamespaceValidation: true,
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(
					test.WithFinalizer,
					test.WithKafkaChannelAddress,
					test.WithInitKafkaChannelConditions,
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelDeploymentServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
					test.WithKafkaChannelDispatcherDeploymentReady,
					test.WithKafkaChannelTopicReady,
				),
				test.NewKafkaChannelDeploymentService(),
				test.NewKafkaChannelChannelDeployment(),
				test.NewKafkaChannelDispatcherService(),
				test.NewKafkaChannelDispatcherDeployment(),
			},
			WantCreates: []runtime.Object{test.NewKafkaChannelChannelService()},
			WantEvents:  []string{test.NewKafkaChannelSuccessfulReconciliationEvent()},
		},
		{
			Name:                    "Reconcile Missing Channel Channel Service Error(Create)",
			SkipNamespaceValidation: true,
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(
					test.WithFinalizer,
					test.WithKafkaChannelAddress,
					test.WithInitKafkaChannelConditions,
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelDeploymentServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
					test.WithKafkaChannelDispatcherDeploymentReady,
					test.WithKafkaChannelTopicReady,
				),
				test.NewKafkaChannelDeploymentService(),
				test.NewKafkaChannelChannelDeployment(),
				test.NewKafkaChannelDispatcherService(),
				test.NewKafkaChannelDispatcherDeployment(),
			},
			WithReactors: []clientgotesting.ReactionFunc{InduceFailure("create", "Services")},
			WantErr:      true,
			WantCreates:  []runtime.Object{test.NewKafkaChannelChannelService()},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: test.NewKnativeKafkaChannel(
						test.WithFinalizer,
						test.WithKafkaChannelAddress,
						test.WithInitKafkaChannelConditions,
						test.WithKafkaChannelChannelServiceFailed,
						test.WithKafkaChannelDeploymentServiceReady,
						test.WithKafkaChannelChannelDeploymentReady,
						test.WithKafkaChannelDispatcherDeploymentReady,
						test.WithKafkaChannelTopicReady,
					),
				},
			},
			WantEvents: []string{
				Eventf(corev1.EventTypeWarning, event.ChannelServiceReconciliationFailed.String(), "Failed To Reconcile Channel Service (Channel): inducing failure for create services"),
				test.NewKafkaChannelFailedReconciliationEvent(),
			},
		},

		//
		// KafkaChannel Channel Deployment Service
		//

		{
			Name:                    "Reconcile Missing Channel Deployment Service Success",
			SkipNamespaceValidation: true,
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(
					test.WithFinalizer,
					test.WithKafkaChannelAddress,
					test.WithInitKafkaChannelConditions,
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelDeploymentServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
					test.WithKafkaChannelDispatcherDeploymentReady,
					test.WithKafkaChannelTopicReady,
				),
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelChannelDeployment(),
				test.NewKafkaChannelDispatcherService(),
				test.NewKafkaChannelDispatcherDeployment(),
			},
			WantCreates: []runtime.Object{test.NewKafkaChannelDeploymentService()},
			WantEvents:  []string{test.NewKafkaChannelSuccessfulReconciliationEvent()},
		},
		{
			Name:                    "Reconcile Missing Channel Deployment Service Error(Create)",
			SkipNamespaceValidation: true,
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(
					test.WithFinalizer,
					test.WithKafkaChannelAddress,
					test.WithInitKafkaChannelConditions,
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelDeploymentServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
					test.WithKafkaChannelDispatcherDeploymentReady,
					test.WithKafkaChannelTopicReady,
				),
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelChannelDeployment(),
				test.NewKafkaChannelDispatcherService(),
				test.NewKafkaChannelDispatcherDeployment(),
			},
			WithReactors: []clientgotesting.ReactionFunc{InduceFailure("create", "services")},
			WantErr:      true,
			WantCreates:  []runtime.Object{test.NewKafkaChannelDeploymentService()},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: test.NewKnativeKafkaChannel(
						test.WithFinalizer,
						test.WithKafkaChannelAddress,
						test.WithInitKafkaChannelConditions,
						test.WithKafkaChannelChannelServiceReady,
						test.WithKafkaChannelDeploymentServiceFailed,
						test.WithKafkaChannelChannelDeploymentReady,
						test.WithKafkaChannelDispatcherDeploymentReady,
						test.WithKafkaChannelTopicReady,
					),
				},
			},
			WantEvents: []string{
				Eventf(corev1.EventTypeWarning, event.ChannelServiceReconciliationFailed.String(), "Failed To Reconcile Channel Service (Deployment): inducing failure for create services"),
				test.NewKafkaChannelFailedReconciliationEvent(),
			},
		},

		//
		// KafkaChannel Channel Deployment
		//

		{
			Name:                    "Reconcile Missing Channel Deployment Success",
			SkipNamespaceValidation: true,
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(
					test.WithFinalizer,
					test.WithKafkaChannelAddress,
					test.WithInitKafkaChannelConditions,
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelDeploymentServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
					test.WithKafkaChannelDispatcherDeploymentReady,
					test.WithKafkaChannelTopicReady,
				),
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelDeploymentService(),
				test.NewKafkaChannelDispatcherService(),
				test.NewKafkaChannelDispatcherDeployment(),
			},
			WantCreates: []runtime.Object{test.NewKafkaChannelChannelDeployment()},
			WantEvents:  []string{test.NewKafkaChannelSuccessfulReconciliationEvent()},
		},
		{
			Name:                    "Reconcile Missing Channel Deployment Error(Create)",
			SkipNamespaceValidation: true,
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(
					test.WithFinalizer,
					test.WithKafkaChannelAddress,
					test.WithInitKafkaChannelConditions,
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelDeploymentServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
					test.WithKafkaChannelDispatcherDeploymentReady,
					test.WithKafkaChannelTopicReady,
				),
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelDeploymentService(),
				test.NewKafkaChannelDispatcherService(),
				test.NewKafkaChannelDispatcherDeployment(),
			},
			WithReactors: []clientgotesting.ReactionFunc{InduceFailure("create", "deployments")},
			WantErr:      true,
			WantCreates:  []runtime.Object{test.NewKafkaChannelChannelDeployment()},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: test.NewKnativeKafkaChannel(
						test.WithFinalizer,
						test.WithKafkaChannelAddress,
						test.WithInitKafkaChannelConditions,
						test.WithKafkaChannelChannelServiceReady,
						test.WithKafkaChannelDeploymentServiceReady,
						test.WithKafkaChannelChannelDeploymentFailed,
						test.WithKafkaChannelDispatcherDeploymentReady,
						test.WithKafkaChannelTopicReady,
					),
				},
			},
			WantEvents: []string{
				Eventf(corev1.EventTypeWarning, event.ChannelDeploymentReconciliationFailed.String(), "Failed To Reconcile Channel Deployment: inducing failure for create deployments"),
				test.NewKafkaChannelFailedReconciliationEvent(),
			},
		},

		//
		// KafkaChannel Dispatcher Service
		//

		{
			Name:                    "Reconcile Missing Dispatcher Service Success",
			SkipNamespaceValidation: true,
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(
					test.WithFinalizer,
					test.WithKafkaChannelAddress,
					test.WithInitKafkaChannelConditions,
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelDeploymentServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
					test.WithKafkaChannelDispatcherDeploymentReady,
					test.WithKafkaChannelTopicReady,
				),
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelDeploymentService(),
				test.NewKafkaChannelChannelDeployment(),
				test.NewKafkaChannelDispatcherDeployment(),
			},
			WantCreates: []runtime.Object{test.NewKafkaChannelDispatcherService()},
			WantEvents:  []string{test.NewKafkaChannelSuccessfulReconciliationEvent()},
		},
		{
			Name:                    "Reconcile Missing Dispatcher Service Error(Create)",
			SkipNamespaceValidation: true,
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(
					test.WithFinalizer,
					test.WithKafkaChannelAddress,
					test.WithInitKafkaChannelConditions,
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelDeploymentServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
					test.WithKafkaChannelDispatcherDeploymentReady,
					test.WithKafkaChannelTopicReady,
				),
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelDeploymentService(),
				test.NewKafkaChannelChannelDeployment(),
				test.NewKafkaChannelDispatcherDeployment(),
			},
			WithReactors: []clientgotesting.ReactionFunc{InduceFailure("create", "services")},
			WantErr:      true,
			WantCreates:  []runtime.Object{test.NewKafkaChannelDispatcherService()},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				// Note - Not currently tracking status for the Dispatcher Service since it is only for Prometheus
			},
			WantEvents: []string{
				Eventf(corev1.EventTypeWarning, event.DispatcherServiceReconciliationFailed.String(), "Failed To Reconcile Dispatcher Service: inducing failure for create services"),
				test.NewKafkaChannelFailedReconciliationEvent(),
			},
		},

		//
		// KafkaChannel Dispatcher Deployment
		//

		{
			Name:                    "Reconcile Missing Dispatcher Deployment Success",
			SkipNamespaceValidation: true,
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(
					test.WithFinalizer,
					test.WithKafkaChannelAddress,
					test.WithInitKafkaChannelConditions,
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelDeploymentServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
					test.WithKafkaChannelDispatcherDeploymentReady,
					test.WithKafkaChannelTopicReady,
				),
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelDeploymentService(),
				test.NewKafkaChannelChannelDeployment(),
				test.NewKafkaChannelDispatcherService(),
			},
			WantCreates: []runtime.Object{test.NewKafkaChannelDispatcherDeployment()},
			WantEvents:  []string{test.NewKafkaChannelSuccessfulReconciliationEvent()},
		},
		{
			Name:                    "Reconcile Missing Dispatcher Deployment Error(Create)",
			SkipNamespaceValidation: true,
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(
					test.WithFinalizer,
					test.WithKafkaChannelAddress,
					test.WithInitKafkaChannelConditions,
					test.WithKafkaChannelChannelServiceReady,
					test.WithKafkaChannelDeploymentServiceReady,
					test.WithKafkaChannelChannelDeploymentReady,
					test.WithKafkaChannelDispatcherDeploymentReady,
					test.WithKafkaChannelTopicReady,
				),
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelDeploymentService(),
				test.NewKafkaChannelChannelDeployment(),
				test.NewKafkaChannelDispatcherService(),
			},
			WithReactors: []clientgotesting.ReactionFunc{InduceFailure("create", "deployments")},
			WantErr:      true,
			WantCreates:  []runtime.Object{test.NewKafkaChannelDispatcherDeployment()},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: test.NewKnativeKafkaChannel(
						test.WithFinalizer,
						test.WithKafkaChannelAddress,
						test.WithInitKafkaChannelConditions,
						test.WithKafkaChannelChannelServiceReady,
						test.WithKafkaChannelDeploymentServiceReady,
						test.WithKafkaChannelChannelDeploymentReady,
						test.WithKafkaChannelDispatcherDeploymentFailed,
						test.WithKafkaChannelTopicReady,
					),
				},
			},
			WantEvents: []string{
				Eventf(corev1.EventTypeWarning, event.DispatcherDeploymentReconciliationFailed.String(), "Failed To Reconcile Dispatcher Deployment: inducing failure for create deployments"),
				test.NewKafkaChannelFailedReconciliationEvent(),
			},
		},
	}

	// Run The TableTest Using The KafkaChannel Reconciler Provided By The Factory
	logger := logtesting.TestLogger(t)
	tableTest.Test(t, test.MakeFactory(func(ctx context.Context, listers *test.Listers, cmw configmap.Watcher) controller.Reconciler {
		r := &Reconciler{
			Base:                  reconciler.NewBase(ctx, constants.KafkaChannelControllerAgentName, cmw),
			adminClient:           &test.MockAdminClient{},
			environment:           test.NewEnvironment(),
			kafkachannelLister:    listers.GetKafkaChannelLister(),
			kafkachannelInformer:  nil, // TODO - Fix ???
			deploymentLister:      listers.GetDeploymentLister(),
			serviceLister:         listers.GetServiceLister(),
			knativekafkaClientSet: fakeknativekafkaclient.Get(ctx),
		}
		return kafkachannelreconciler.NewReconciler(ctx, r.Logger, r.knativekafkaClientSet, listers.GetKafkaChannelLister(), r.Recorder, r)
	}, logger.Desugar()))
}

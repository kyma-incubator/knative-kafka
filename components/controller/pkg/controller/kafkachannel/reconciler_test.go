package kafkachannel

import (
	"context"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	fakeknativekafkaclient "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/client/fake"
	kafkachannelreconciler "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkachannel"
	"github.com/kyma-incubator/knative-kafka/components/controller/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/eventing/pkg/reconciler"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	logtesting "knative.dev/pkg/logging/testing"

	"k8s.io/client-go/kubernetes/scheme"
	"testing"

	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"

	//eventingreconcilertesting "knative.dev/eventing/pkg/reconciler/testing"
	// TODO - from the eventing in-memory test
	. "knative.dev/pkg/reconciler/testing"
	//. "knative.dev/pkg/reconciler/testing"

	// TODO - from the eventing-contrib kafkachannel test
	//	reconcilekafkatesting "knative.dev/eventing-contrib/kafka/channel/pkg/reconciler/testing"
	//	reconcilertesting "knative.dev/eventing-contrib/kafka/channel/pkg/reconciler/testing"

)

/* TODO - inmemory test imports
import (
	"context"
	"fmt"
	"testing"

	"knative.dev/eventing/pkg/client/injection/reconciler/messaging/v1alpha1/inmemorychannel"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	clientgotesting "k8s.io/client-go/testing"
	eventingduckv1alpha1 "knative.dev/eventing/pkg/apis/duck/v1alpha1"
	"knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/eventing/pkg/reconciler/inmemorychannel/controller/resources"
	. "knative.dev/eventing/pkg/reconciler/testing"
	reconciletesting "knative.dev/eventing/pkg/reconciler/testing"
	"knative.dev/eventing/pkg/utils"
	"knative.dev/pkg/apis"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmeta"
	logtesting "knative.dev/pkg/logging/testing"
	. "knative.dev/pkg/reconciler/testing"
)

*/

// TODO - eventing-contrib implementations (kafkachannel, natss, etc have their own copy of makefactory

func init() {
	// Add types to scheme
	_ = knativekafkav1alpha1.AddToScheme(scheme.Scheme)
	//_ = v1alpha1.AddToScheme(scheme.Scheme) // tODO - do we need this ?    "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	_ = duckv1alpha1.AddToScheme(scheme.Scheme)
}

/* TODO
// Test Initialization
func init() {

	// Get Test Logger
	logger := log.TestLogger()

	// Add 3rd Party Types To The Scheme
	err := kafkav1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		logger.Fatal("Failed To Add Kafka Channel Eventing Scheme", zap.Error(err))
	}
}

// Test The InjectClient() Functionality`
func TestInjectClient(t *testing.T) {
	r := &Reconciler{}
	origClient := r.client
	newClient := fake.NewFakeClient()
	assert.NotEqual(t, origClient, newClient)
	err := r.InjectClient(newClient)
	assert.Nil(t, err)
	assert.Equal(t, newClient, r.client)
}

// Test The NewReconciler Functionality
func TestNewReconciler(t *testing.T) {

	// Test Data
	testCase := &test.TestCase{}
	recorder := testCase.GetEventRecorder()
	logger := log.TestLogger()
	client := &test.MockAdminClient{}
	environment := &env.Environment{}

	// Perform The Test
	reconciler := NewReconciler(recorder, logger, client, environment)

	// Validate Results
	assert.NotNil(t, reconciler)
	assert.Equal(t, recorder, reconciler.recorder)
	assert.Equal(t, logger, reconciler.logger)
	assert.Equal(t, client, reconciler.adminClient)
	assert.Equal(t, environment, reconciler.environment)
}
*/

// TODO - NEW !!!
// Test The Reconcile Functionality
func TestReconcile(t *testing.T) {

	// Clear Logs After Testing
	defer logtesting.ClearAll()

	// Define The KafkaChannel Reconciler Test Cases
	tableTest := TableTest{
		//{
		//	Name:    "Bad KafkaChannel Key",
		//	Key:     "too/many/parts",
		//	WantErr: false,
		//},
		//{
		//	Name:    "KafkaChannel Key Not Found",
		//	Key:     "foo/not-found",
		//	WantErr: false,
		//},
		//{
		//	Name: "Delete KafkaChannel",
		//	Key:  test.KafkaChannelKey,
		//	Objects: []runtime.Object{
		//		test.NewKnativeKafkaChannel(test.WithInitKafkaChannelConditions, test.WithKafkaChannelDeleted),
		//	},
		//	WantErr: false,
		//	WantEvents: []string{
		//		// TODO - see notes in FinalizeKind() about this mess (at a minimum add event constants)
		//		Eventf(corev1.EventTypeNormal, "KafkaChannel Finalized", "Topic %s.%s deleted during finalization", test.KafkaChannelNamespace, test.KafkaChannelName),
		//	},
		//},
		{
			Name:                    "Complete Reconciliation Success",
			SkipNamespaceValidation: true, // TODO - Knative Testing Framework Assumes ALL Actions Will Be In The Same Namespace As The Key !!!
			Key:                     test.KafkaChannelKey,
			Objects: []runtime.Object{
				test.NewKnativeKafkaChannel(test.WithInitKafkaChannelConditions),
			},
			WantErr: false,
			WantCreates: []runtime.Object{
				test.NewKafkaChannelChannelService(),
				test.NewKafkaChannelDeploymentService(),
				test.NewKafkaChannelDeployment(),
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
			// TODO - Why doesn't the in-memory test have/need this?  this comes from adding the finalizer!
			// TODO - In memory doesn't have this because they don't implement FinalizeKind() and thus don't need to add a finalizer !!!!!
			WantPatches: []clientgotesting.PatchActionImpl{
				{
					ActionImpl: clientgotesting.ActionImpl{
						Namespace:   test.KafkaChannelNamespace,
						Verb:        "patch",
						Resource:    schema.GroupVersionResource{Group: knativekafkav1alpha1.SchemeGroupVersion.Group, Version: knativekafkav1alpha1.SchemeGroupVersion.Version, Resource: "kafkachannels"},
						Subresource: "",
					},
					Name:      test.KafkaChannelName,
					PatchType: "application/merge-patch+json",
					Patch:     []byte(`{"metadata":{"finalizers":["kafkachannels.knativekafka.kyma-project.io"],"resourceVersion":""}}`),
				},
			},
			WantEvents: []string{
				Eventf(corev1.EventTypeNormal, "FinalizerUpdate", `Updated "%s" finalizers`, test.KafkaChannelName),
				Eventf(corev1.EventTypeNormal, "KafkaChannelReconciled", `KafkaChannel reconciled: "%s/%s"`, test.KafkaChannelNamespace, test.KafkaChannelName),
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

/* TODO - Compiler Error above...

 Cannot use 'r.EventingClientSet' (type "github.com/kyma-incubator/knative-kafka/components/controller/vendor/knative.dev/eventing/pkg/client/clientset/versioned".Interface) as type
                                        "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned".Interface
 Type does not implement 'versioned.Interface' as some methods are missing: KnativekafkaV1alpha1() knativekafkav1alpha1.KnativekafkaV1alpha1Interface

... it wants to add this function to the knative.dev/eventing/pkg/client/clientset/versioned/clientset.go  file !!!

func (i Interface) KnativekafkaV1alpha1() v1alpha1.KnativekafkaV1alpha1Interface {
	panic("implement me")
}

... so why is NewBase() creating the vendor/knative.dev/eventing clientset instead of ours ?   I


sooo... the in-memory implementation does it this way above (calling NewReconciler) BUT the kafka/natss eventing-contrib tests don't
        they just create the inline Reconciler.  The difference is that those structs have their own Reconcile() function whereas ours and
        the in-memory implementation get their Reconcile function from the injection package which has a reconcilerImpl{} struct
        the natss/kafka don't even HAVE an injection/reconciler package - only injection/client and injection/informers

i think we need our own copy of NewBase that sets the right clientset ???  natss has it's own NewBase() implementation which sets a NatssClientSet instead of an EventingClientSet
but kafka doesnt have one?

natss has also defined it's own Base{} struct containing the NatssClientSet

both natss and kafka have their own stats_reporter.go implementations - for now i just copied in the kafkachannel one

OR... instead of creating our own Base/NewBase() we could use the one we're using and add a new smaller struct with the knativekafka clientset? and/or not call the injected NewReconciler() - we don't call from from the controller.go???
		(basically add the clientset directly to our reconciler?)

The kafka implementation adds their clientset directly in their reconciler definition and uses NewBase (but not the generated/injection code - has it's own Reconcile())

*!!! the reason the in-memory one works is that it is a part of the Messaging clientset and treated like channels/subscriptions (which is odd/bad - not what we're doing for eventing-contrib inclusion)

so we're kind of in this weird trifecta of implementations and unsure which one to pattern after...

	- do we want to NOT generate the injected reconciler and write our own Reconcile() ?
	- do we want to write our own Base/NewBase()
*/

/* eventing-contrib kafka example
tableTest.Test(t, test.MakeFactory(func(ctx context.Context, listers *test.Listers, cmw configmap.Watcher) controller.Reconciler {

	return &Reconciler{
		Base: reconciler.NewBase(ctx, constants.KafkaChannelControllerAgentName, cmw),
		//dispatcherNamespace:      testNS,
		//dispatcherDeploymentName: testDispatcherDeploymentName,
		//dispatcherServiceName:    testDispatcherServiceName,
		//dispatcherImage:          testDispatcherImage,
		//kafkaConfig: &KafkaConfig{
		//	Brokers: []string{brokerName},
		//},
		//kafkachannelLister: listers.GetKafkaChannelLister(),
		//// TODO fix
		//kafkachannelInformer: nil,
		deploymentLister: listers.GetDeploymentLister(),
		serviceLister:    listers.GetServiceLister(),
		//endpointsLister:      listers.GetEndpointsLister(), // TODO - don't have one of these yet
		//kafkaClusterAdmin:    &mockClusterAdmin{},
		//kafkaClientSet:       fakekafkaclient.Get(ctx),
	}
}, zap.L()))
*/

/*
// Test The Reconcile Functionality
func TestReconcile(t *testing.T) {

	// Define All TestCases
	testCases := []test.TestCase{
		{
			Name:           "Channel Not Found",
			InitialState:   []runtime.Object{},
			ExpectedResult: reconcile.Result{},
			ExpectedError:  nil,
			ExpectedAbsent: []runtime.Object{
				test.GetNewChannel(true, true, false),
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "Channel Lookup Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, false),
			},
			Mocks: test.Mocks{
				GetFns: []test.MockGetFn{test.MockGetFnKafkaChannelError},
			},
			ExpectedResult: reconcile.Result{},
			ExpectedError:  errors.New(test.MockGetFnKafkaChannelErrorMessage),
			ExpectedAbsent: []runtime.Object{
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannel(true, true, false),
			},
		},
		{
			Name: "Deleted Channel",
			InitialState: []runtime.Object{
				test.GetNewChannelDeleted(true, true, true),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedError:  nil,
			ExpectedAbsent: []runtime.Object{
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelDeleted(false, true, true),
			},
		},
		{
			Name: "New Channel",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedError:  nil,
			ExpectedAbsent: []runtime.Object{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusReady()),
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "New Channel Without Finalizer",
			InitialState: []runtime.Object{
				test.GetNewChannel(false, true, false),
			},
			ExpectedResult: reconcile.Result{
				Requeue: true,
			},
			ExpectedAbsent: []runtime.Object{
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, false, false, test.GetChannelStatusReady()),
			},
		},
		{
			Name: "New Channel Without Spec Properties",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, false, true),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, false, true, true, test.GetChannelStatusReady()),
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.DefaultTopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.DefaultTopicName),
			},
		},
		{
			Name: "New Channel Without Subscribers",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, false),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, false, true, test.GetChannelStatusReady()),
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "New Channel Filtering Other Deployments",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
				test.GetNewChannelWithName("OtherChannel", true, true, true),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusReady()),
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
				test.GetNewChannelWithName("OtherChannel", true, true, true),
			},
		},
		{
			Name: "Update Channel",
			InitialState: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusReady()),
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusReady()),
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "Update Channel (Finalizer) Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(false, true, true),
			},
			Mocks: test.Mocks{
				UpdateFns: []test.MockUpdateFn{test.MockUpdateFnKafkaChannelError},
			},
			ExpectedError:  errors.New(test.MockUpdateFnKafkaChannelErrorMessage),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelUpdateFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannel(false, true, true),
			},
		},
		{
			Name: "Update Channel (Status) Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks: test.Mocks{
				StatusUpdateFns: []test.MockStatusUpdateFn{test.MockStatusUpdateFnChannelError},
			},
			ExpectedError:  errors.New(test.MockStatusUpdateFnChannelErrorMessage),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelUpdateFailed.String()}},
			ExpectedAbsent: []runtime.Object{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannel(true, true, true),
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "Create Channel Deployment Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnChannelDeploymentError},
			},
			ExpectedError:  errors.New("reconciliation failed"),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelDeploymentReconciliationFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SChannelDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusChannelDeploymentFailed()),
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "Create KafkaChannel Service Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnKafkaChannelServiceError},
			},
			ExpectedError:  errors.New("reconciliation failed"),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelServiceReconciliationFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewKafkaChannelService(),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusChannelServiceFailed()),
				test.GetNewChannelDeploymentService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "Create Channel Deployment Service Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnChannelDeploymentServiceError},
			},
			ExpectedError:  errors.New("reconciliation failed"),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelServiceReconciliationFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewChannelDeploymentService(),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusChannelDeploymentServiceFailed()),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "Create Dispatcher Deployment Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnDispatcherDeploymentError},
			},
			ExpectedError:  errors.New("reconciliation failed"),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.DispatcherDeploymentReconciliationFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusDispatcherDeploymentFailed()),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SDispatcherService(),
			},
		},
		{
			Name: "Create Dispatcher Service Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnDispatcherServiceError},
			},
			ExpectedError:  errors.New("reconciliation failed"),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.DispatcherServiceReconciliationFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusReady()),
				test.GetNewChannelDeploymentService(),
				test.GetNewKafkaChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
	}

	//
	// Run All The TestCases Against The DataFlow Reconciler
	//
	for _, testCase := range test.FilterTestCases(testCases) {

		// Default The TestCase Name / Namespace If Not Specified
		if len(testCase.ReconcileName) == 0 {
			testCase.ReconcileName = test.ChannelName
		}
		if len(testCase.ReconcileNamespace) == 0 {
			testCase.ReconcileNamespace = test.NamespaceName
		}

		// Create The Reconciler With Initialized Fake Client
		reconciler := &Reconciler{
			client:      testCase.GetClient(),
			recorder:    testCase.GetEventRecorder(),
			logger:      log.TestLogger(),
			adminClient: &test.MockAdminClient{},
			environment: test.NewEnvironment(),
		}

		// Run The TestCase
		testStarted := t.Run(testCase.Name, testCase.Runner(reconciler))
		assert.True(t, testStarted)
	}
}

//
// Test The Kafka Topic Reconciliation
//
// Ideally the Knative Eventing test runner implementation would have provided a hook for additional
// channel-type-specific (ie kafka, nats, etc) validation, but unfortunately it is solely
// focused on the K8S objects existing/not.  Therefore we're left to test the actual Topic handling
// separately.
//
func TestReconcileTopic(t *testing.T) {

	// Define The Topic TestCase
	type topicTestCase struct {
		Name                   string
		Channel                *kafkav1alpha1.KafkaChannel
		WantTopicSpecification kafka.TopicSpecification
		MockErrorCode          kafka.ErrorCode
		WantError              string
		WantCreate             bool
		WantDelete             bool
	}

	// Define & Initialize The Topic TestCases
	topicTestCases := []topicTestCase{
		{
			Name:       "Create New Topic",
			Channel:    test.GetNewChannel(true, true, false),
			WantCreate: true,
			WantDelete: false,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
		},
		{
			Name:       "Create Preexisting Topic",
			Channel:    test.GetNewChannel(true, true, false),
			WantCreate: true,
			WantDelete: false,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
			MockErrorCode: kafka.ErrTopicAlreadyExists,
		},
		{
			Name:       "Error Creating Topic",
			Channel:    test.GetNewChannel(true, true, false),
			WantCreate: true,
			WantDelete: false,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
			MockErrorCode: kafka.ErrAllBrokersDown,
			WantError:     test.ErrorString,
		},
		{
			Name:       "Delete Existing Topic",
			Channel:    test.GetNewChannelDeleted(true, true, false),
			WantCreate: false,
			WantDelete: true,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
		},
		{
			Name:       "Delete Nonexistent Topic",
			Channel:    test.GetNewChannelDeleted(true, true, false),
			WantCreate: false,
			WantDelete: true,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
			MockErrorCode: kafka.ErrUnknownTopic,
		},
		{
			Name:       "Error Deleting Topic",
			Channel:    test.GetNewChannelDeleted(true, true, false),
			WantCreate: false,
			WantDelete: true,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
			MockErrorCode: kafka.ErrAllBrokersDown,
			WantError:     test.ErrorString,
		},
	}

	// Create A New Recorder & Logger For Testing
	recorder := record.NewBroadcaster().NewRecorder(scheme.Scheme, corev1.EventSource{Component: constants.KafkaChannelControllerAgentName})
	logger := log.TestLogger()

	// Loop Over All TestCases - Setup The Reconciler & Run Test Via Custom Runner
	for _, tc := range topicTestCases {

		// Setup Desired Mock ClusterAdmin Behavior
		mockAdminClient := &test.MockAdminClient{
			// Mock CreateTopic Behavior - Validate Parameters & Return MockError
			MockCreateTopicFunc: func(ctx context.Context, topics []kafka.TopicSpecification, options ...kafka.CreateTopicsAdminOption) (result []kafka.TopicResult, err error) {
				if !tc.WantCreate {
					t.Errorf("Unexpected CreateTopics() Call")
				}
				if ctx == nil {
					t.Error("expected non nil context")
				}
				if len(topics) != 1 {
					t.Errorf("expected one TopicSpecification but received %d", len(topics))
				}
				if diff := cmp.Diff(tc.WantTopicSpecification, topics[0]); diff != "" {
					t.Errorf("expected TopicSpecification: %+v", diff)
				}
				if options != nil {
					t.Error("expected nil options")
				}
				var topicResults []kafka.TopicResult
				if tc.MockErrorCode != 0 {
					topicResults = []kafka.TopicResult{
						{
							Topic: tc.WantTopicSpecification.Topic,
							Error: kafka.NewError(tc.MockErrorCode, test.ErrorString, false),
						},
					}
				}
				return topicResults, nil
			},
			//Mock DeleteTopic Behavior - Validate Parameters & Return MockError
			MockDeleteTopicFunc: func(ctx context.Context, topics []string, options ...kafka.DeleteTopicsAdminOption) (result []kafka.TopicResult, err error) {
				if !tc.WantDelete {
					t.Errorf("Unexpected DeleteTopics() Call")
				}
				if ctx == nil {
					t.Error("expected non nil context")
				}
				if len(topics) != 1 {
					t.Errorf("expected one TopicSpecification but received %d", len(topics))
				}
				if diff := cmp.Diff(tc.WantTopicSpecification.Topic, topics[0]); diff != "" {
					t.Errorf("expected TopicSpecification: %+v", diff)
				}
				if options != nil {
					t.Error("expected nil options")
				}
				var topicResults []kafka.TopicResult
				if tc.MockErrorCode != 0 {
					topicResults = []kafka.TopicResult{
						{
							Topic: tc.WantTopicSpecification.Topic,
							Error: kafka.NewError(tc.MockErrorCode, test.ErrorString, false),
						},
					}
				}
				return topicResults, nil
			},
		}

		// Initialize The TestCases Reconciler
		r := &Reconciler{
			recorder:    recorder,
			logger:      logger,
			adminClient: mockAdminClient,
			environment: test.NewEnvironment(),
		}

		// Log The Test & Run With Custom Inline Test Runner
		t.Logf("running test %s", tc.Name)
		t.Run(tc.Name, func(t *testing.T, tc topicTestCase, r *Reconciler) func(t *testing.T) {
			return func(t *testing.T) {

				// Perform The Test - Reconcile The Topic For The TestChannel
				err := r.reconcileTopic(context.TODO(), tc.Channel)

				// Verify Create/Delete Called
				if tc.WantCreate != mockAdminClient.CreateTopicsCalled() {
					t.Errorf("expected CreateTopics() called to be %t", tc.WantCreate)
				}
				if tc.WantDelete != mockAdminClient.DeleteTopicsCalled() {
					t.Errorf("expected DeleteTopics() called to be %t", tc.WantCreate)
				}

				// Validate TestCase Expected Error State
				var got string
				if err != nil {
					got = err.Error()
				}
				if diff := cmp.Diff(tc.WantError, got); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
			}
		}(t, tc, r))
	}
}
*/

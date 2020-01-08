package kafkasubscription

import (
	"errors"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

// Test Initialization
func init() {

	// Get Test Logger
	logger := log.TestLogger()

	// Add 3rd Party Types To The Scheme
	err := kafkav1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		logger.Fatal("Failed To Add Kafka Channel Schema", zap.Error(err))
	}
	err = messagingv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		logger.Fatal("Failed To Add Knative Messaging Schema", zap.Error(err))
	}
}

// Test The InjectClient() Functionality
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

// Test The Reconcile Functionality
func TestReconcile(t *testing.T) {

	//
	// Define All TestCases
	//
	// Note - The "Direct / Indirect" naming refers to whether the Knative Subscription refers
	//        directly to a KafkaChannel or indirectly via a Knative Messaging Channel.
	//
	testCases := []test.TestCase{
		{
			Name:           "Subscription Not Found",
			InitialState:   []runtime.Object{},
			ExpectedResult: reconcile.Result{},
			ExpectedAbsent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, false, test.EventStartTime),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "Subscription Lookup Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, false),
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, false, test.EventStartTime),
			},
			Mocks: test.Mocks{
				GetFns: []test.MockGetFn{test.MockGetFnSubscriptionError},
			},
			ExpectedError: errors.New(test.MockGetFnSubscriptionErrorMessage),
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannel(true, true, false),
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, false, test.EventStartTime),
			},
		},
		{
			Name: "Channel Not Found (Direct Deletion)",
			InitialState: []runtime.Object{
				test.GetNewSubscriptionDeleted(test.NamespaceName, test.SubscriberName, true, true),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscriptionDeleted(test.NamespaceName, test.SubscriberName, true, false),
			},
		},
		{
			Name: "Channel Not Found (Direct Non-Deletion)",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
			},
			ExpectedResult: reconcile.Result{Requeue: true},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
			},
		},
		{
			Name: "Channel Not Found (Indirect Deletion)",
			InitialState: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime, true),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, false, test.EventStartTime, true),
			},
		},
		{
			Name: "Channel Not Found (Indirect Non-Deletion)",
			InitialState: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime, false),
			},
			ExpectedResult: reconcile.Result{Requeue: true},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime, false),
			},
		},
		{
			Name: "Channel Lookup Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, false),
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
			},
			Mocks: test.Mocks{
				GetFns: []test.MockGetFn{test.MockGetFnKafkaChannelError},
			},
			ExpectedError: errors.New(test.MockGetFnKafkaChannelErrorMessage),
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannel(true, true, false),
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
			},
		},
		{
			Name: "Missing Finalizer",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, false, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			ExpectedResult: reconcile.Result{Requeue: true},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
		},
		{
			Name: "Subscription Deletion",
			InitialState: []runtime.Object{
				test.GetNewSubscriptionDeleted(test.NamespaceName, test.SubscriberName, true, true),
				test.GetNewChannel(true, true, false),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscriptionDeleted(test.NamespaceName, test.SubscriberName, true, false),
				test.GetNewChannel(true, true, false),
			},
		},
		{
			Name: "Dispatcher Service Lookup Error",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			Mocks: test.Mocks{
				GetFns: []test.MockGetFn{test.MockGetFnServiceError},
			},
			ExpectedError: errors.New("reconciliation failed"),
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedEvents: []corev1.Event{{Reason: event.DispatcherK8sServiceReconciliationFailed.String(), Type: corev1.EventTypeWarning}},
		},
		{
			Name: "Dispatcher Service Creation Error",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnServiceError},
			},
			ExpectedError: errors.New("reconciliation failed"),
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "Dispatcher Deployment Lookup Error",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			Mocks: test.Mocks{
				GetFns: []test.MockGetFn{test.MockGetFnDeploymentError},
			},
			ExpectedError: errors.New("reconciliation failed"),
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherService(),
			},
			ExpectedEvents: []corev1.Event{{Reason: event.DispatcherDeploymentReconciliationFailed.String(), Type: corev1.EventTypeWarning}},
		},
		{
			Name: "Dispatcher Deployment Creation Error",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnDeploymentError},
			},
			ExpectedError: errors.New("reconciliation failed"),
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherService(),
			},
			ExpectedEvents: []corev1.Event{{Reason: event.DispatcherDeploymentReconciliationFailed.String(), Type: corev1.EventTypeWarning}},
		},
		{
			Name: "Full Reconciliation (Direct Success)",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "Full Reconciliation (Indirect Success)",
			InitialState: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime, false),
				test.GetNewKnativeMessagingChannel(kafkav1alpha1.SchemeGroupVersion.String(), constants.KafkaChannelKind),
				test.GetNewChannel(true, true, false),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime, false),
				test.GetNewKnativeMessagingChannel(kafkav1alpha1.SchemeGroupVersion.String(), constants.KafkaChannelKind),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
		},
		{
			Name: "Full Reconciliation Without Annotations (Success)",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, false, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, false, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeploymentWithoutAnnotations(test.TopicName),
			},
		},

		// TODO - Disabled Until Replay/EventStartTime Is Re-Enabled (Will probably require some love to catch up with test framework changes)
		//{
		//	Name: "Test Event Start Time Created (Success)",
		//	InitialState: []runtime.Object{
		//		test.GetNewClusterChannelProvisioner(test.ClusterChannelProvisionerName, true),
		//		test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, "2017-06-05T12:00:00Z"),
		//		test.GetNewChannel(test.ChannelName, test.ClusterChannelProvisionerName, true, true, false),
		//	},
		//	ExpectedResult: reconcile.Result{},
		//	ExpectedPresent: []runtime.Object{
		//		test.GetNewClusterChannelProvisioner(test.ClusterChannelProvisionerName, true),
		//		test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, "2017-06-05T12:00:00Z"),
		//		test.GetNewChannel(test.ChannelName, test.ClusterChannelProvisionerName, true, true, false),
		//		test.GetNewK8SDispatcherService(),
		//		test.GetNewK8SDispatcherDeploymentWithEventStartTime(test.TopicName, "2017-06-05T12:00:00Z"),
		//	},
		//},
		//{
		//	Name: "Test Event Start Time Updated (Success)",
		//	InitialState: []runtime.Object{
		//		test.GetNewClusterChannelProvisioner(test.ClusterChannelProvisionerName, true),
		//		test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, "2017-06-05T12:00:00Z"),
		//		test.GetNewChannel(test.ChannelName, test.ClusterChannelProvisionerName, true, true, false),
		//		test.GetNewK8SDispatcherService(),
		//		test.GetNewK8SDispatcherDeploymentWithEventStartTime(test.TopicName, "2015-06-05T12:00:00Z"),
		//	},
		//	ExpectedResult: reconcile.Result{},
		//	ExpectedPresent: []runtime.Object{
		//		test.GetNewClusterChannelProvisioner(test.ClusterChannelProvisionerName, true),
		//		test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, "2017-06-05T12:00:00Z"),
		//		test.GetNewChannel(test.ChannelName, test.ClusterChannelProvisionerName, true, true, false),
		//		test.GetNewK8SDispatcherService(),
		//		test.GetNewK8SDispatcherDeploymentWithEventStartTime(test.TopicName, "2017-06-05T12:00:00Z"),
		//	},
		//},
	}

	//
	// Run All The TestCases Against The DataFlow Reconciler
	//
	for _, testCase := range test.FilterTestCases(testCases) {

		// Default The TestCase Name / Namespace If Not Specified
		if len(testCase.ReconcileName) == 0 {
			testCase.ReconcileName = test.SubscriberName
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

package kafkasubscription

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	"github.com/knative/eventing/pkg/provisioners"
	controllertesting "github.com/knative/eventing/pkg/reconciler/testing"
	kafkaconsumer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/consumer"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

// Test Initialization
func init() {

	// Add 3rd Party Types To The Scheme
	err := kafkav1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("Failed To Add Kafka Channel Eventing Schema: %+v", err)
	}
	err = eventingv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("Failed To Add Knative Eventing Schema: %+v", err)
	}
	err = messagingv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("Failed To Add Knative Messaging Schema: %+v", err)
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

// Test The Reconcile Functionality
func TestReconcile(t *testing.T) {

	//
	// Define All TestCases
	//
	// Note - The "Direct / Indirect" naming refers to whether the Knative Subscription refers
	//        directly to a KafkaChannel or indirectly via a Knative Messaging Channel.
	//
	testCases := []controllertesting.TestCase{
		{
			Name:         "Subscription Not Found",
			InitialState: []runtime.Object{},
			WantResult:   reconcile.Result{},
			WantAbsent: []runtime.Object{
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
			Mocks:      controllertesting.Mocks{MockGets: test.SubscriptionMockGetsError},
			WantErrMsg: test.SubscriptionMockGetsErrorMessage,
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewChannel(true, true, false),
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, false, test.EventStartTime),
			},
		},
		{
			Name: "Channel Not Found (Direct Deletion)",
			InitialState: []runtime.Object{
				test.GetNewSubscriptionDeleted(test.NamespaceName, test.SubscriberName, true, true),
			},
			WantResult: reconcile.Result{},
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewSubscriptionDeleted(test.NamespaceName, test.SubscriberName, true, false),
			},
		},
		{
			Name: "Channel Not Found (Direct Non-Deletion)",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
			},
			WantResult: reconcile.Result{Requeue: true},
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
			},
		},
		{
			Name: "Channel Not Found (Indirect Deletion)",
			InitialState: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime, true),
			},
			WantResult: reconcile.Result{},
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, false, test.EventStartTime, true),
			},
		},
		{
			Name: "Channel Not Found (Indirect Non-Deletion)",
			InitialState: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime, false),
			},
			WantResult: reconcile.Result{Requeue: true},
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime, false),
			},
		},
		{
			Name: "Channel Lookup Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, false),
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
			},
			Mocks:      controllertesting.Mocks{MockGets: test.ChannelMockGetsError},
			WantErrMsg: test.ChannelMockGetsErrorMessage,
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
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
			WantResult: reconcile.Result{Requeue: true},
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
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
			WantResult: reconcile.Result{},
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
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
			Mocks:      controllertesting.Mocks{MockGets: test.ServiceMockGetsError},
			WantErrMsg: "reconciliation failed",
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
			},
			WantPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			OtherTestData: map[string]interface{}{
				"dispatcherDeployment": test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			AdditionalVerification: []func(t *testing.T, tc *controllertesting.TestCase){test.VerifyWantDeploymentPresent("dispatcherDeployment")},
			WantEvent:              []corev1.Event{{Reason: event.DispatcherK8sServiceReconciliationFailed.String(), Type: corev1.EventTypeWarning}},
		},
		{
			Name: "Dispatcher Service Creation Error",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			Mocks:      controllertesting.Mocks{MockCreates: test.ServiceMockCreatesError},
			WantErrMsg: "reconciliation failed",
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(),
			},
			WantPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			OtherTestData: map[string]interface{}{
				"dispatcherDeployment": test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			AdditionalVerification: []func(t *testing.T, tc *controllertesting.TestCase){test.VerifyWantDeploymentPresent("dispatcherDeployment")},
			WantEvent:              []corev1.Event{{Reason: event.DispatcherK8sServiceReconciliationFailed.String(), Type: corev1.EventTypeWarning}},
		},
		{
			Name: "Dispatcher Deployment Lookup Error",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			Mocks:      controllertesting.Mocks{MockGets: test.DeploymentMockGetsError},
			WantErrMsg: "reconciliation failed",
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherService(),
			},
			WantEvent: []corev1.Event{{Reason: event.DispatcherDeploymentReconciliationFailed.String(), Type: corev1.EventTypeWarning}},
		},
		{
			Name: "Dispatcher Deployment Creation Error",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			Mocks:      controllertesting.Mocks{MockCreates: test.DeploymentMockCreatesError},
			WantErrMsg: "reconciliation failed",
			WantAbsent: []runtime.Object{
				test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherService(),
			},
			WantEvent: []corev1.Event{{Reason: event.DispatcherDeploymentReconciliationFailed.String(), Type: corev1.EventTypeWarning}},
		},
		{
			Name: "Full Reconciliation (Direct Success)",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			WantResult: reconcile.Result{},
			WantPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherService(),
			},
			OtherTestData: map[string]interface{}{
				"dispatcherDeployment": test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			AdditionalVerification: []func(t *testing.T, tc *controllertesting.TestCase){test.VerifyWantDeploymentPresent("dispatcherDeployment")},
		},
		{
			Name: "Full Reconciliation (Indirect Success)",
			InitialState: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime, false),
				test.GetNewKnativeMessagingChannel(kafkav1alpha1.SchemeGroupVersion.String(), constants.KafkaChannelKind),
				test.GetNewChannel(true, true, false),
			},
			WantResult: reconcile.Result{},
			WantPresent: []runtime.Object{
				test.GetNewSubscriptionIndirectChannel(test.NamespaceName, test.SubscriberName, true, true, test.EventStartTime, false),
				test.GetNewKnativeMessagingChannel(kafkav1alpha1.SchemeGroupVersion.String(), constants.KafkaChannelKind),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherService(),
			},
			OtherTestData: map[string]interface{}{
				"dispatcherDeployment": test.GetNewK8SDispatcherDeployment(test.TopicName),
			},
			AdditionalVerification: []func(t *testing.T, tc *controllertesting.TestCase){test.VerifyWantDeploymentPresent("dispatcherDeployment")},
		},
		{
			Name: "Full Reconciliation Without Annotations (Success)",
			InitialState: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, false, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
			},
			WantResult: reconcile.Result{},
			WantPresent: []runtime.Object{
				test.GetNewSubscription(test.NamespaceName, test.SubscriberName, false, true, test.EventStartTime),
				test.GetNewChannel(true, true, false),
				test.GetNewK8SDispatcherService(),
			},
			OtherTestData: map[string]interface{}{
				"dispatcherDeployment": test.GetNewK8SDispatcherDeploymentWithoutAnnotations(test.TopicName),
			},
			AdditionalVerification: []func(t *testing.T, tc *controllertesting.TestCase){test.VerifyWantDeploymentPresent("dispatcherDeployment")},
		},

		// TODO - Disabled Until Replay/EventStartTime Is Re-Enabled
		//{
		//	Name: "Test Event Start Time Created (Success)",
		//	InitialState: []runtime.Object{
		//		test.GetNewClusterChannelProvisioner(test.ClusterChannelProvisionerName, true),
		//		test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, "2017-06-05T12:00:00Z"),
		//		test.GetNewChannel(test.ChannelName, test.ClusterChannelProvisionerName, true, true, false),
		//	},
		//	WantResult: reconcile.Result{},
		//	WantPresent: []runtime.Object{
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
		//	WantResult: reconcile.Result{},
		//	WantPresent: []runtime.Object{
		//		test.GetNewClusterChannelProvisioner(test.ClusterChannelProvisionerName, true),
		//		test.GetNewSubscription(test.NamespaceName, test.SubscriberName, true, "2017-06-05T12:00:00Z"),
		//		test.GetNewChannel(test.ChannelName, test.ClusterChannelProvisionerName, true, true, false),
		//		test.GetNewK8SDispatcherService(),
		//		test.GetNewK8SDispatcherDeploymentWithEventStartTime(test.TopicName, "2017-06-05T12:00:00Z"),
		//	},
		//},
	}

	// Create A New Recorder & Logger For Testing
	logger := provisioners.NewProvisionerLoggerFromConfig(provisioners.NewLoggingConfig())

	// Replace The NewClient Wrapper To Provide Mock Consumer & Defer Reset
	newConsumerWrapperPlaceholder := kafkaconsumer.NewConsumerWrapper
	kafkaconsumer.NewConsumerWrapper = func(configMap *kafka.ConfigMap) (kafkaconsumer.ConsumerInterface, error) {
		return test.NewMockConsumer(), nil
	}
	defer func() { kafkaconsumer.NewConsumerWrapper = newConsumerWrapperPlaceholder }()

	// Loop Over All TestCases - Setup The Reconciler & Run Test Via Knative Eventing TestCase Runner
	for _, testCase := range testCases {
		recorder := testCase.GetEventRecorder()
		c := testCase.GetClient()
		test.MockClient = c
		r := &Reconciler{
			client:      c,
			recorder:    recorder,
			logger:      logger.Desugar(),
			adminClient: &test.MockAdminClient{},
			environment: test.NewEnvironment(),
		}
		testCase.ReconcileKey = fmt.Sprintf("%s/%s", test.NamespaceName, test.SubscriberName)
		testCase.IgnoreTimes = true
		t.Logf("Running TestCase %s", testCase.Name)
		t.Run(testCase.Name, testCase.Runner(t, r, c, recorder))
	}
}

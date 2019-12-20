package kafkachannel

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/go-cmp/cmp"
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
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
	"testing"
)

// TODO - ORIG
// controllertesting "github.com/knative/eventing/pkg/reconciler/testing"
import (
	"context"
	"k8s.io/client-go/kubernetes/scheme"
	controllertesting "knative.dev/eventing/pkg/reconciler/testing"
	istiov1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
	logtesting "knative.dev/pkg/logging/testing"
	reconcilertesting "knative.dev/pkg/reconciler/testing"
)

// TODO reconcilertesting "knative.dev/eventing/pkg/reconciler/testing"

// Test Initialization
func init() {

	// Get Test Logger
	logger := log.TestLogger()

	// Add 3rd Party Types To The Scheme
	err := kafkav1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		logger.Fatal("Failed To Add Kafka Channel Eventing Scheme", zap.Error(err))
	}
	err = istiov1alpha3.AddToScheme(scheme.Scheme)
	if err != nil {
		logger.Fatal("Failed To Add Istio Scheme", zap.Error(err))
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
	// TODO testCase := &controllertesting.TestCase{}
	testCase := &reconcilertesting.TableTest{}
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

	// Define All TestCases
	// TODO testCases := []controllertesting.TestCase{
	// TODO - see the inmemorychannel_test.go for usage example
	// TODO - see the table.go file for supported TableRow{} data
	// TODO - changes...
	//      - InitialState is now Objects
	//      - WantErrMsg string is now WantErr bool
	//      - ReconcileKey is now Key (add to every test instead of once below)
	tableTest := reconcilertesting.TableTest{
		{
			Name:       "Channel Not Found",
			Objects:    []runtime.Object{},
			WantErr:    false,
			WantResult: reconcile.Result{},
			WantAbsent: []runtime.Object{
				test.GetNewChannel(true, true, false),
				test.GetNewK8sChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
			},
		},
		{
			Name: "Channel Lookup Error",
			Objects: []runtime.Object{
				test.GetNewChannel(true, true, false),
			},
			Mocks: controllertesting.Mocks{MockGets: test.ChannelMockGetsError},
			// TODO WantErrMsg: test.ChannelMockGetsErrorMessage,
			WantErr: true,
			WantAbsent: []runtime.Object{
				test.GetNewK8sChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewChannel(true, true, false),
			},
		},
		{
			Name: "Deleted Channel",
			Objects: []runtime.Object{
				test.GetNewChannelDeleted(true, true, true),
			},
			WantErr:    false,
			WantResult: reconcile.Result{},
			WantAbsent: []runtime.Object{
				test.GetNewK8sChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewChannelDeleted(false, true, true),
			},
		},
		{
			Name: "New Channel",
			Objects: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			WantErr:    false,
			WantResult: reconcile.Result{},
			WantAbsent: []runtime.Object{},
			WantPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, true),
				test.GetNewK8sChannelService(),
			},
			OtherTestData: map[string]interface{}{
				"channelDeployment": test.GetNewK8SChannelDeployment(test.TopicName),
			},
			AdditionalVerification: []func(t *testing.T, tc *controllertesting.TestCase){test.VerifyWantDeploymentPresent("channelDeployment")},
		},
		{
			Name: "New Channel Without Finalizer",
			Objects: []runtime.Object{
				test.GetNewChannel(false, true, false),
			},
			WantErr: false,
			WantResult: reconcile.Result{
				Requeue: true,
			},
			WantAbsent: []runtime.Object{
				test.GetNewK8sChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, false, false, false),
			},
		},
		{
			Name: "New Channel Without Spec Properties",
			Objects: []runtime.Object{
				test.GetNewChannel(true, false, true),
			},
			WantErr:    false,
			WantResult: reconcile.Result{},
			WantPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, false, true, true, true),
				test.GetNewK8sChannelService(),
			},
			OtherTestData: map[string]interface{}{
				"channelDeployment": test.GetNewK8SChannelDeployment(test.DefaultTopicName),
			},
			AdditionalVerification: []func(t *testing.T, tc *controllertesting.TestCase){test.VerifyWantDeploymentPresent("channelDeployment")},
		},
		{
			Name: "New Channel Without Subscribers",
			Objects: []runtime.Object{
				test.GetNewChannel(true, true, false),
			},
			WantErr:    false,
			WantResult: reconcile.Result{},
			WantPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, false, true, true),
				test.GetNewK8sChannelService(),
			},
			OtherTestData: map[string]interface{}{
				"channelDeployment": test.GetNewK8SChannelDeployment(test.TopicName),
			},
			AdditionalVerification: []func(t *testing.T, tc *controllertesting.TestCase){test.VerifyWantDeploymentPresent("channelDeployment")},
		},
		{
			Name: "New Channel Filtering Other Deployments",
			Objects: []runtime.Object{
				test.GetNewChannel(true, true, true),
				test.GetNewChannelWithName("OtherChannel", true, true, true),
			},
			WantErr:    false,
			WantResult: reconcile.Result{},
			WantPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, true),
				test.GetNewK8sChannelService(),
				test.GetNewChannelWithName("OtherChannel", true, true, true),
			},
			OtherTestData: map[string]interface{}{
				"channelDeployment": test.GetNewK8SChannelDeployment(test.TopicName),
			},
			AdditionalVerification: []func(t *testing.T, tc *controllertesting.TestCase){test.VerifyWantDeploymentPresent("channelDeployment")},
		},
		{
			Name: "Update Channel (Finalizer) Error",
			Objects: []runtime.Object{
				test.GetNewChannel(false, true, true),
			},
			Mocks: controllertesting.Mocks{MockUpdates: test.ChannelMockUpdatesError},
			// TODO WantErrMsg: test.ChannelMockUpdatesErrorMessage,
			WantErr:   true,
			WantEvent: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelUpdateFailed.String()}},
			WantAbsent: []runtime.Object{
				test.GetNewK8sChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewChannel(false, true, true),
			},
		},
		{
			Name: "Update Channel (Status) Error",
			Objects: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks: controllertesting.Mocks{MockStatusUpdates: test.ChannelMockStatusUpdatesError},
			// TODO WantErrMsg: test.ChannelMockStatusUpdatesErrorMessage,
			WantErr:    true,
			WantEvent:  []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelUpdateFailed.String()}},
			WantAbsent: []runtime.Object{},
			WantPresent: []runtime.Object{
				test.GetNewChannel(true, true, true),
				test.GetNewK8sChannelService(),
			},
			OtherTestData: map[string]interface{}{
				"channelDeployment": test.GetNewK8SChannelDeployment(test.TopicName),
			},
			AdditionalVerification: []func(t *testing.T, tc *controllertesting.TestCase){test.VerifyWantDeploymentPresent("channelDeployment")},
		},
		{
			Name: "Create Channel Deployment Error",
			Objects: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks: controllertesting.Mocks{MockCreates: test.DeploymentMockCreatesError},
			// TODO WantErrMsg: "reconciliation failed",
			WantErr:   true,
			WantEvent: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelDeploymentReconciliationFailed.String()}},
			WantAbsent: []runtime.Object{
				test.GetNewK8SChannelDeployment(test.TopicName),
			},
			WantPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, true),
				test.GetNewK8sChannelService(),
			},
		},
		{
			Name: "Create Channel Service Error",
			Objects: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks: controllertesting.Mocks{MockCreates: test.ServiceMockCreatesError},
			// TODO WantErrMsg: "reconciliation failed",
			WantErr:   true,
			WantEvent: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelServiceReconciliationFailed.String()}},
			WantAbsent: []runtime.Object{
				test.GetNewK8sChannelService(),
			},
			WantPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, false),
			},
			OtherTestData: map[string]interface{}{
				"channelDeployment": test.GetNewK8SChannelDeployment(test.TopicName),
			},
			AdditionalVerification: []func(t *testing.T, tc *controllertesting.TestCase){test.VerifyWantDeploymentPresent("channelDeployment")},
		},
	}

	/* TODO - Original Implementation

	// Create A New Recorder & Logger For Testing
	logger := logging.NewLoggerFromConfig(logging.NewLoggingConfig())

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
		testCase.ReconcileKey = fmt.Sprintf("%s/%s", test.NamespaceName, test.ChannelName)
		testCase.IgnoreTimes = true
		t.Logf("Running TestCase %s", testCase.Name)
		t.Run(testCase.Name, testCase.Runner(t, r, c, recorder))
	}

	*/

	// TODO - New Implementation
	logger := logtesting.TestLogger(t)
	factory := controllerrtesting.MakeFactory()
	tableTest.Test(t, factory)

	/* TODO - New Example Test Execution looks like this...

		logger := logtesting.TestLogger(t)
	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		return &Reconciler{
			Base:                     reconciler.NewBase(ctx, controllerAgentName, cmw),
			dispatcherNamespace:      testNS,
			dispatcherDeploymentName: testDispatcherDeploymentName,
			dispatcherServiceName:    testDispatcherServiceName,
			inmemorychannelLister:    listers.GetInMemoryChannelLister(),
			// TODO: FIx
			inmemorychannelInformer: nil,
			deploymentLister:        listers.GetDeploymentLister(),
			serviceLister:           listers.GetServiceLister(),
			endpointsLister:         listers.GetEndpointsLister(),
		}
	},
		false,
		logger,
	))

	*/
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
	logger := logging.NewLoggerFromConfig(logging.NewLoggingConfig())

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
			logger:      logger.Desugar(),
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

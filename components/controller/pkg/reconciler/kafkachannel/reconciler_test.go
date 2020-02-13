package kafkachannel

import (
	"context"
	"errors"
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
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
	"testing"
)

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
				test.GetNewChannel(true, true, false, 1),
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
		},
		{
			Name: "Channel Lookup Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, false, 1),
			},
			Mocks: test.Mocks{
				GetFns: []test.MockGetFn{test.MockGetFnKafkaChannelError},
			},
			ExpectedResult: reconcile.Result{},
			ExpectedError:  errors.New(test.MockGetFnKafkaChannelErrorMessage),
			ExpectedAbsent: []runtime.Object{
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannel(true, true, false, 1),
			},
		},
		{
			Name: "Deleted Channel",
			InitialState: []runtime.Object{
				test.GetNewChannelDeleted(true, true, true, 1),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedError:  nil,
			ExpectedAbsent: []runtime.Object{
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelDeleted(false, true, true, 2),
			},
		},
		{
			Name: "New Channel",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true, 1),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedError:  nil,
			ExpectedAbsent: []runtime.Object{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusReady(), 2),
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
		},
		{
			Name: "New Channel Without Finalizer",
			InitialState: []runtime.Object{
				test.GetNewChannel(false, true, false, 0),
			},
			ExpectedResult: reconcile.Result{
				Requeue: true,
			},
			ExpectedAbsent: []runtime.Object{
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, false, false, test.GetChannelStatusReady(), 1),
			},
		},
		{
			Name: "New Channel Without Spec Properties",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, false, true, 0),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, false, true, true, test.GetChannelStatusReady(), 1),
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.DefaultTopicName, 1),
			},
		},
		{
			Name: "New Channel Without Subscribers",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, false, 0),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, false, true, test.GetChannelStatusReady(), 1),
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
		},
		{
			Name: "New Channel Filtering Other Deployments",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true, 0),
				test.GetNewChannelWithName("OtherChannel", true, true, true, 1),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusReady(), 1),
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
				test.GetNewChannelWithName("OtherChannel", true, true, true, 1),
			},
		},
		{
			Name: "Update Channel",
			InitialState: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusReady(), 1),
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
			ExpectedResult: reconcile.Result{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusReady(), 1),
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
		},
		{
			Name: "Update Channel (Finalizer) Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(false, true, true, 1),
			},
			Mocks: test.Mocks{
				UpdateFns: []test.MockUpdateFn{test.MockUpdateFnKafkaChannelError},
			},
			ExpectedError:  errors.New(test.MockUpdateFnKafkaChannelErrorMessage),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelUpdateFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannel(false, true, true, 1),
			},
		},
		{
			Name: "Update Channel (Status) Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true, 1),
			},
			Mocks: test.Mocks{
				StatusUpdateFns: []test.MockStatusUpdateFn{test.MockStatusUpdateFnChannelError},
			},
			ExpectedError:  errors.New(test.MockStatusUpdateFnChannelErrorMessage),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelUpdateFailed.String()}},
			ExpectedAbsent: []runtime.Object{},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannel(true, true, true, 1),
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
		},
		{
			Name: "Create Channel Deployment Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true, 1),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnChannelDeploymentError},
			},
			ExpectedError:  errors.New("reconciliation failed"),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelDeploymentReconciliationFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SChannelDeployment(1),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusChannelDeploymentFailed(), 2),
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
		},
		{
			Name: "Create KafkaChannel Service Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true, 1),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnKafkaChannelServiceError},
			},
			ExpectedError:  errors.New("reconciliation failed"),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelServiceReconciliationFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewKafkaChannelService(1),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusChannelServiceFailed(), 2),
				test.GetNewChannelDeploymentService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
		},
		{
			Name: "Create Channel Deployment Service Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true, 1),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnChannelDeploymentServiceError},
			},
			ExpectedError:  errors.New("reconciliation failed"),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelServiceReconciliationFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewChannelDeploymentService(1),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusChannelDeploymentServiceFailed(), 2),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherService(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
		},
		{
			Name: "Create Dispatcher Deployment Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true, 1),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnDispatcherDeploymentError},
			},
			ExpectedError:  errors.New("reconciliation failed"),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.DispatcherDeploymentReconciliationFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusDispatcherDeploymentFailed(), 2),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SDispatcherService(1),
			},
		},
		{
			Name: "Create Dispatcher Service Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true, 1),
			},
			Mocks: test.Mocks{
				CreateFns: []test.MockCreateFn{test.MockCreateFnDispatcherServiceError},
			},
			ExpectedError:  errors.New("reconciliation failed"),
			ExpectedEvents: []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.DispatcherServiceReconciliationFailed.String()}},
			ExpectedAbsent: []runtime.Object{
				test.GetNewK8SDispatcherService(1),
			},
			ExpectedPresent: []runtime.Object{
				test.GetNewChannelWithProvisionedStatus(true, true, true, true, test.GetChannelStatusReady(), 2),
				test.GetNewChannelDeploymentService(1),
				test.GetNewKafkaChannelService(1),
				test.GetNewK8SChannelDeployment(1),
				test.GetNewK8SDispatcherDeployment(test.TopicName, 1),
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
			Channel:    test.GetNewChannel(true, true, false, 1),
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
			Channel:    test.GetNewChannel(true, true, false, 1),
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
			Channel:    test.GetNewChannel(true, true, false, 1),
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
			Channel:    test.GetNewChannelDeleted(true, true, false, 1),
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
			Channel:    test.GetNewChannelDeleted(true, true, false, 1),
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
			Channel:    test.GetNewChannelDeleted(true, true, false, 1),
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

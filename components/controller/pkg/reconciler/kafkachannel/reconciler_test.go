package kafkachannel

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/go-cmp/cmp"
	"github.com/knative/eventing/pkg/provisioners"
	controllertesting "github.com/knative/eventing/pkg/reconciler/testing"
	istiov1alpha3 "github.com/knative/pkg/apis/istio/v1alpha3"
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
	testCase := &controllertesting.TestCase{}
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
	testCases := []controllertesting.TestCase{
		{
			Name:         "Channel Not Found",
			InitialState: []runtime.Object{},
			WantResult:   reconcile.Result{},
			WantAbsent: []runtime.Object{
				test.GetNewChannel(true, true, false),
				test.GetNewK8sChannelService(),
				test.GetNewK8SChannelDeployment(test.TopicName),
			},
		},
		{
			Name: "Channel Lookup Error",
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, false),
			},
			Mocks:      controllertesting.Mocks{MockGets: test.ChannelMockGetsError},
			WantErrMsg: test.ChannelMockGetsErrorMessage,
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
			InitialState: []runtime.Object{
				test.GetNewChannelDeleted(true, true, true),
			},
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
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
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
			InitialState: []runtime.Object{
				test.GetNewChannel(false, true, false),
			},
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
			InitialState: []runtime.Object{
				test.GetNewChannel(true, false, true),
			},
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
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, false),
			},
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
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
				test.GetNewChannelWithName("OtherChannel", true, true, true),
			},
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
			InitialState: []runtime.Object{
				test.GetNewChannel(false, true, true),
			},
			Mocks:      controllertesting.Mocks{MockUpdates: test.ChannelMockUpdatesError},
			WantErrMsg: test.ChannelMockUpdatesErrorMessage,
			WantEvent:  []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelUpdateFailed.String()}},
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
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks:      controllertesting.Mocks{MockStatusUpdates: test.ChannelMockStatusUpdatesError},
			WantErrMsg: test.ChannelMockStatusUpdatesErrorMessage,
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
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks:      controllertesting.Mocks{MockCreates: test.DeploymentMockCreatesError},
			WantErrMsg: "reconciliation failed",
			WantEvent:  []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelDeploymentReconciliationFailed.String()}},
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
			InitialState: []runtime.Object{
				test.GetNewChannel(true, true, true),
			},
			Mocks:      controllertesting.Mocks{MockCreates: test.ServiceMockCreatesError},
			WantErrMsg: "reconciliation failed",
			WantEvent:  []corev1.Event{{Type: corev1.EventTypeWarning, Reason: event.ChannelServiceReconciliationFailed.String()}},
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

	// Create A New Recorder & Logger For Testing
	logger := provisioners.NewProvisionerLoggerFromConfig(provisioners.NewLoggingConfig())

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
	logger := provisioners.NewProvisionerLoggerFromConfig(provisioners.NewLoggingConfig())

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

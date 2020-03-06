package kafkachannel

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/go-cmp/cmp"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"knative.dev/eventing/pkg/reconciler"
	logtesting "knative.dev/pkg/logging/testing"
	"strconv"
	"testing"
)

// Define The Topic TestCase Type
type TopicTestCase struct {
	Name                   string
	Channel                *knativekafkav1alpha1.KafkaChannel
	WantTopicSpecification kafka.TopicSpecification
	MockErrorCode          kafka.ErrorCode
	WantError              string
	WantCreate             bool
	WantDelete             bool
}

//
// Test The Kafka Topic Reconciliation
//
// Ideally the Knative Eventing test runner implementation would have provided a hook for additional
// channel-type-specific (ie kafka, nats, etc) validation, but unfortunately it is solely focused
// on the K8S objects existing/not.  Therefore we're left to test the actual Topic handling separately.
//
func TestReconcileTopic(t *testing.T) {

	// Define & Initialize The TopicTestCases
	topicTestCases := []TopicTestCase{
		{
			Name: "Skip Deleted Topic",
			Channel: test.NewKnativeKafkaChannel(
				test.WithFinalizer,
				test.WithKafkaChannelDeleted,
				test.WithKafkaChannelAddress,
				test.WithInitKafkaChannelConditions,
				test.WithKafkaChannelChannelServiceReady,
				test.WithKafkaChannelDeploymentServiceReady,
				test.WithKafkaChannelChannelDeploymentReady,
				test.WithKafkaChannelDispatcherDeploymentReady,
			),
			WantCreate: false,
			WantDelete: false,
		},
		{
			Name: "Create New Topic",
			Channel: test.NewKnativeKafkaChannel(
				test.WithFinalizer,
				test.WithKafkaChannelAddress,
				test.WithInitKafkaChannelConditions,
				test.WithKafkaChannelChannelServiceReady,
				test.WithKafkaChannelDeploymentServiceReady,
				test.WithKafkaChannelChannelDeploymentReady,
				test.WithKafkaChannelDispatcherDeploymentReady,
			),
			WantCreate: true,
			WantDelete: false,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					constants.KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
		},
		{
			Name: "Create Preexisting Topic",
			Channel: test.NewKnativeKafkaChannel(
				test.WithFinalizer,
				test.WithKafkaChannelAddress,
				test.WithInitKafkaChannelConditions,
				test.WithKafkaChannelChannelServiceReady,
				test.WithKafkaChannelDeploymentServiceReady,
				test.WithKafkaChannelChannelDeploymentReady,
				test.WithKafkaChannelDispatcherDeploymentReady,
			),
			WantCreate: true,
			WantDelete: false,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					constants.KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
			MockErrorCode: kafka.ErrTopicAlreadyExists,
		},
		{
			Name: "Error Creating Topic",
			Channel: test.NewKnativeKafkaChannel(
				test.WithFinalizer,
				test.WithKafkaChannelAddress,
				test.WithInitKafkaChannelConditions,
				test.WithKafkaChannelChannelServiceReady,
				test.WithKafkaChannelDeploymentServiceReady,
				test.WithKafkaChannelChannelDeploymentReady,
				test.WithKafkaChannelDispatcherDeploymentReady,
			),
			WantCreate: true,
			WantDelete: false,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					constants.KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
			MockErrorCode: kafka.ErrAllBrokersDown,
			WantError:     test.ErrorString,
		},
		{
			Name: "Delete Existing Topic",
			Channel: test.NewKnativeKafkaChannel(
				test.WithFinalizer,
				test.WithKafkaChannelAddress,
				test.WithInitKafkaChannelConditions,
				test.WithKafkaChannelChannelServiceReady,
				test.WithKafkaChannelDeploymentServiceReady,
				test.WithKafkaChannelChannelDeploymentReady,
				test.WithKafkaChannelDispatcherDeploymentReady,
			),
			WantCreate: false,
			WantDelete: true,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					constants.KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
		},
		{
			Name: "Delete Nonexistent Topic",
			Channel: test.NewKnativeKafkaChannel(
				test.WithFinalizer,
				test.WithKafkaChannelAddress,
				test.WithInitKafkaChannelConditions,
				test.WithKafkaChannelChannelServiceReady,
				test.WithKafkaChannelDeploymentServiceReady,
				test.WithKafkaChannelChannelDeploymentReady,
				test.WithKafkaChannelDispatcherDeploymentReady,
			),
			WantCreate: false,
			WantDelete: true,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					constants.KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
			MockErrorCode: kafka.ErrUnknownTopic,
		},
		{
			Name: "Error Deleting Topic",
			Channel: test.NewKnativeKafkaChannel(
				test.WithFinalizer,
				test.WithKafkaChannelAddress,
				test.WithInitKafkaChannelConditions,
				test.WithKafkaChannelChannelServiceReady,
				test.WithKafkaChannelDeploymentServiceReady,
				test.WithKafkaChannelChannelDeploymentReady,
				test.WithKafkaChannelDispatcherDeploymentReady,
			),
			WantCreate: false,
			WantDelete: true,
			WantTopicSpecification: kafka.TopicSpecification{
				Topic:             test.TopicName,
				NumPartitions:     test.NumPartitions,
				ReplicationFactor: test.ReplicationFactor,
				Config: map[string]string{
					constants.KafkaTopicConfigRetentionMs: strconv.FormatInt(test.RetentionMillis, 10),
				},
			},
			MockErrorCode: kafka.ErrAllBrokersDown,
			WantError:     test.ErrorString,
		},
	}

	// Run All The TopicTestCases
	for _, tc := range topicTestCases {
		t.Run(tc.Name, topicTestCaseFactory(tc))
	}
}

// Factory For Creating A Go Test Function For The Specified TopicTestCase
func topicTestCaseFactory(tc TopicTestCase) func(t *testing.T) {
	return func(t *testing.T) {

		// Create A New Recorder & Logger For Testing
		recorder := record.NewBroadcaster().NewRecorder(scheme.Scheme, corev1.EventSource{Component: constants.KafkaChannelControllerAgentName})

		// Create A Mock Kafka AdminClient For Current TopicTestCase
		mockAdminClient := createMockAdminClientForTestCase(t, tc)

		// Initialize The Reconciler For The Current TopicTestCase
		r := &Reconciler{
			Base: &reconciler.Base{
				Logger:   logtesting.TestLogger(t),
				Recorder: recorder,
			},
			adminClient: mockAdminClient,
			environment: test.NewEnvironment(),
		}

		// Track Any Error Responses
		var err error

		// Perform The Test (Create) - Normal Topic Reconciliation Called Indirectly From ReconcileKind()
		if tc.WantCreate {
			err = r.reconcileTopic(context.TODO(), tc.Channel)
			if !mockAdminClient.CreateTopicsCalled() {
				t.Errorf("expected CreateTopics() called to be %t", tc.WantCreate)
			}
		}

		// Perform The Test (Delete) - Called By Knative FinalizeKind() Directly
		if tc.WantDelete {
			err = r.deleteTopic(context.TODO(), test.TopicName)
			if !mockAdminClient.DeleteTopicsCalled() {
				t.Errorf("expected DeleteTopics() called to be %t", tc.WantCreate)
			}
		}

		// Validate TestCase Expected Error State
		var errorString string
		if err != nil {
			errorString = err.Error()
		}
		if diff := cmp.Diff(tc.WantError, errorString); diff != "" {
			t.Errorf("unexpected error (-want, +got) = %v", diff)
		}
	}
}

// Create A Mock Kafka AdminClient For The Specified TopicTestCase
func createMockAdminClientForTestCase(t *testing.T, tc TopicTestCase) *test.MockAdminClient {

	// Setup Desired Mock ClusterAdmin Behavior From TopicTestCase
	return &test.MockAdminClient{

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
}

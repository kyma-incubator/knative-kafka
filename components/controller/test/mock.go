package test

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	controllertesting "github.com/knative/eventing/pkg/reconciler/testing"
	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	kafkaconsumer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/consumer"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Mock Constants
const (
	ServiceMockGetsErrorMessage          = "mock get service error"
	ServiceMockCreatesErrorMessage       = "mock create service error"
	DeploymentMockGetsErrorMessage       = "mock get deployment error"
	DeploymentMockCreatesErrorMessage    = "mock create deployment error"
	ChannelMockGetsErrorMessage          = "mock get channel error"
	ChannelMockUpdatesErrorMessage       = "mock update channel error"
	ChannelMockStatusUpdatesErrorMessage = "mock statusupdate channel error"
	SubscriptionMockGetsErrorMessage     = "mock get subscription error"
)

//
// TestCase Mocks
//

// Mock The Gets Of Services To Return Error
var ServiceMockGetsError = []controllertesting.MockGet{
	func(innerClient client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) (controllertesting.MockHandled, error) {
		if _, ok := obj.(*corev1.Service); ok {
			err := fmt.Errorf(ServiceMockGetsErrorMessage)
			return controllertesting.Handled, err
		}
		return controllertesting.Unhandled, nil
	},
}

// Mock The Creates Of Services To Return Error
var ServiceMockCreatesError = []controllertesting.MockCreate{
	func(innerClient client.Client, ctx context.Context, obj runtime.Object) (controllertesting.MockHandled, error) {
		if _, ok := obj.(*corev1.Service); ok {
			err := fmt.Errorf(ServiceMockCreatesErrorMessage)
			return controllertesting.Handled, err
		}
		return controllertesting.Unhandled, nil
	},
}

// Mock The Gets Of Deployments To Return Error
var DeploymentMockGetsError = []controllertesting.MockGet{
	func(innerClient client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) (controllertesting.MockHandled, error) {
		if _, ok := obj.(*appsv1.Deployment); ok {
			err := fmt.Errorf(DeploymentMockGetsErrorMessage)
			return controllertesting.Handled, err
		}
		return controllertesting.Unhandled, nil
	},
}

// Mock The Creates Of Deployments To Return Error
var DeploymentMockCreatesError = []controllertesting.MockCreate{
	func(innerClient client.Client, ctx context.Context, obj runtime.Object) (controllertesting.MockHandled, error) {
		if _, ok := obj.(*appsv1.Deployment); ok {
			err := fmt.Errorf(DeploymentMockCreatesErrorMessage)
			return controllertesting.Handled, err
		}
		return controllertesting.Unhandled, nil
	},
}

// Mock The Gets Of Channel To Return Error
var ChannelMockGetsError = []controllertesting.MockGet{
	func(innerClient client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) (controllertesting.MockHandled, error) {
		if _, ok := obj.(*kafkav1alpha1.KafkaChannel); ok {
			err := fmt.Errorf(ChannelMockGetsErrorMessage)
			return controllertesting.Handled, err
		}
		return controllertesting.Unhandled, nil
	},
}

// Mock The Updates Of Channel To Return Error
var ChannelMockUpdatesError = []controllertesting.MockUpdate{
	func(innerClient client.Client, ctx context.Context, obj runtime.Object) (controllertesting.MockHandled, error) {
		if _, ok := obj.(*kafkav1alpha1.KafkaChannel); ok {
			err := fmt.Errorf(ChannelMockUpdatesErrorMessage)
			return controllertesting.Handled, err
		} else {
			return controllertesting.Unhandled, nil
		}
	},
}

// Mock The StatusUpdates Of Channel To Return Error
var ChannelMockStatusUpdatesError = []controllertesting.MockStatusUpdate{
	func(innerClient client.Client, ctx context.Context, obj runtime.Object) (handled controllertesting.MockHandled, e error) {
		if _, ok := obj.(*kafkav1alpha1.KafkaChannel); ok {
			err := fmt.Errorf(ChannelMockStatusUpdatesErrorMessage)
			return controllertesting.Handled, err
		} else {
			return controllertesting.Unhandled, nil
		}
	},
}

// Mock The Gets Of Subscription To Return Error
var SubscriptionMockGetsError = []controllertesting.MockGet{
	func(innerClient client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) (controllertesting.MockHandled, error) {
		if _, ok := obj.(*eventingv1alpha1.Subscription); ok {
			err := fmt.Errorf(SubscriptionMockGetsErrorMessage)
			return controllertesting.Handled, err
		}
		return controllertesting.Unhandled, nil
	},
}

//
// Mock Confluent AdminClient
//

// Verify The Mock AdminClient Implements The KafkaAdminClient Interface
var _ kafkaadmin.AdminClientInterface = &MockAdminClient{}

// Mock Kafka AdminClient Implementation
type MockAdminClient struct {
	createTopicsCalled  bool
	deleteTopicsCalled  bool
	MockCreateTopicFunc func(context.Context, []kafka.TopicSpecification, ...kafka.CreateTopicsAdminOption) (result []kafka.TopicResult, err error)
	MockDeleteTopicFunc func(context.Context, []string, ...kafka.DeleteTopicsAdminOption) (result []kafka.TopicResult, err error)
}

// Mock Kafka AdminClient CreateTopics Function - Calls Custom CreateTopics If Specified, Otherwise Returns Success
func (m *MockAdminClient) CreateTopics(ctx context.Context, topics []kafka.TopicSpecification, options ...kafka.CreateTopicsAdminOption) (result []kafka.TopicResult, err error) {
	m.createTopicsCalled = true
	if m.MockCreateTopicFunc != nil {
		return m.MockCreateTopicFunc(ctx, topics, options...)
	}
	return []kafka.TopicResult{}, nil
}

// Check On Calls To CreateTopics()
func (m *MockAdminClient) CreateTopicsCalled() bool {
	return m.createTopicsCalled
}

// Mock Kafka AdminClient DeleteTopics Function - Calls Custom DeleteTopics If Specified, Otherwise Returns Success
func (m *MockAdminClient) DeleteTopics(ctx context.Context, topics []string, options ...kafka.DeleteTopicsAdminOption) (result []kafka.TopicResult, err error) {
	m.deleteTopicsCalled = true
	if m.MockDeleteTopicFunc != nil {
		return m.MockDeleteTopicFunc(ctx, topics, options...)
	}
	return []kafka.TopicResult{}, nil
}

// Check On Calls To DeleteTopics()
func (m *MockAdminClient) DeleteTopicsCalled() bool {
	return m.deleteTopicsCalled
}

// Mock Kafka AdminClient Close Function - NoOp
func (m *MockAdminClient) Close() {
	return
}

// Mock Kafka Secret Name Function - Return Test Data
func (m *MockAdminClient) GetKafkaSecretName(topicName string) string {
	return KafkaSecret
}

//
// Mock ConsumerInterface Implementation
//

var _ kafkaconsumer.ConsumerInterface = &MockConsumer{}

func NewMockConsumer() *MockConsumer {
	return &MockConsumer{
		events: make(chan kafka.Event),
		closed: false,
	}
}

type MockConsumer struct {
	events             chan kafka.Event
	eventsChanEnable   bool
	readerTermChan     chan bool
	rebalanceCb        kafka.RebalanceCb
	appReassigned      bool
	appRebalanceEnable bool // config setting
	closed             bool
}

func (mc *MockConsumer) StoreOffsets(offsets []kafka.TopicPartition) (storedOffsets []kafka.TopicPartition, err error) {
	return nil, nil
}

func (mc *MockConsumer) Commit() ([]kafka.TopicPartition, error) {
	return nil, nil
}

func (mc *MockConsumer) QueryWatermarkOffsets(topic string, partition int32, timeoutMs int) (low, high int64, err error) {
	return 0, 0, nil
}

func (mc *MockConsumer) OffsetsForTimes(times []kafka.TopicPartition, timeoutMs int) (offsets []kafka.TopicPartition, err error) {
	return nil, nil
}

func (mc *MockConsumer) GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error) {
	metadata := kafka.Metadata{
		Topics: map[string]kafka.TopicMetadata{
			TopicName: {
				Partitions: []kafka.PartitionMetadata{{}, {}},
			},
		},
	}
	return &metadata, nil
}

func (mc *MockConsumer) CommitOffsets(offsets []kafka.TopicPartition) ([]kafka.TopicPartition, error) {
	return nil, nil
}

func (mc *MockConsumer) Poll(timeout int) kafka.Event {
	return <-mc.events
}

func (mc *MockConsumer) CommitMessage(*kafka.Message) ([]kafka.TopicPartition, error) {
	return nil, nil
}

func (mc *MockConsumer) Close() error {
	mc.closed = true
	close(mc.events)
	return nil
}

func (mc *MockConsumer) Subscribe(topic string, rebalanceCb kafka.RebalanceCb) error {
	return nil
}

func (mc *MockConsumer) sendMessage(message kafka.Event) {
	mc.events <- message
}



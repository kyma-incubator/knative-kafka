package test

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	kafkaconsumer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/consumer"
	kafkautil "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/util"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// Mock Constants
const (
	MockGetFnKafkaChannelErrorMessage = "mock Get() KafkaChannel error"
	MockGetFnSubscriptionErrorMessage = "mock Get() Subscription error"
	MockGetFnServiceErrorMessage      = "mock Get() Service error"
	MockGetFnDeploymentErrorMessage   = "mock Get() Deployment error"

	MockCreateFnServiceErrorMessage    = "mock Create() Service error"
	MockCreateFnDeploymentErrorMessage = "mock Create() Deployment error"

	MockUpdateFnKafkaChannelErrorMessage = "mock Update() KafkaChannel error"

	MockStatusUpdateFnChannelErrorMessage = "mock Status().Update() KafkaChannel error"
)

//
// Mock Client
//
// A thin wrapper around the specified client (likely the fake k8s client for testing) which allows
// for injecting Mocks for the Client calls (Get, List, Create, etc...) as a way of customizing
// test behavior (e.g - returning errors).
//
// The array of mocks for a specific Client function will be executed until one of them indicates
// that it has "handled" the request.  If there are no mocks, or none of them "handle" the request,
// then the inner client will be called.
//

// MockHandled "Enum" Type
type MockHandled int

// MockHandled "Enum" Values
const (
	Handled MockHandled = iota
	Unhandled
)

// Define The Mock Function Types
type MockGetFn func(innerClient client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) (MockHandled, error)
type MockListFn func(innerClient client.Client, ctx context.Context, list runtime.Object, opts ...client.ListOption) (MockHandled, error)
type MockCreateFn func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.CreateOption) (MockHandled, error)
type MockDeleteFn func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) (MockHandled, error)
type MockUpdateFn func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) (MockHandled, error)
type MockPatchFn func(innerClient client.Client, ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) (MockHandled, error)
type MockDeleteAllOf func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) (MockHandled, error)
type MockStatusUpdateFn func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) (MockHandled, error)
type MockStatusPatchFn func(innerClient client.Client, ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) (MockHandled, error)

// The Set Of Mocks To Include When Testing The Controller
type Mocks struct {
	GetFns          []MockGetFn
	ListFns         []MockListFn
	CreateFns       []MockCreateFn
	DeleteFns       []MockDeleteFn
	UpdateFns       []MockUpdateFn
	PatchFns        []MockPatchFn
	DeleteAllOfFns  []MockDeleteAllOf
	StatusUpdateFns []MockStatusUpdateFn
	StatusPatchFns  []MockStatusPatchFn
}

// The MockClient Type - Wraps Actual (Fake) Client For Mocking
type MockClient struct {
	innerClient  client.Client
	statusWriter client.StatusWriter
	mocks        Mocks
}

// MockClient Constructor
func NewMockClient(innerClient client.Client, mocks Mocks) *MockClient {
	return &MockClient{
		innerClient:  innerClient,
		statusWriter: NewMockStatusWriter(innerClient, StatusWriterMocks{UpdateFns: mocks.StatusUpdateFns, PatchFns: mocks.StatusPatchFns}),
		mocks:        mocks,
	}
}

// Verify MockClient Implements The Client Interface
var _ client.Client = &MockClient{}

// Call Mock Get() Until Handled Or Delegate To Inner Client
func (mc *MockClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	for _, mockGet := range mc.mocks.GetFns {
		handled, err := mockGet(mc.innerClient, ctx, key, obj)
		if handled == Handled {
			return err
		}
	}
	return mc.innerClient.Get(ctx, key, obj)
}

// Call Mock List() Until Handled Or Delegate To Inner Client
func (mc *MockClient) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	for _, mockList := range mc.mocks.ListFns {
		handled, err := mockList(mc.innerClient, ctx, list, opts...)
		if handled == Handled {
			return err
		}
	}
	return mc.innerClient.List(ctx, list, opts...)
}

// Call Mock Create() Until Handled Or Delegate To Inner Client
func (mc *MockClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	for _, mockCreate := range mc.mocks.CreateFns {
		handled, err := mockCreate(mc.innerClient, ctx, obj, opts...)
		if handled == Handled {
			return err
		}
	}
	return mc.innerClient.Create(ctx, obj)
}

// Call Mock Delete() Until Handled Or Delegate To Inner Client
func (mc *MockClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	for _, mockDelete := range mc.mocks.DeleteFns {
		handled, err := mockDelete(mc.innerClient, ctx, obj, opts...)
		if handled == Handled {
			return err
		}
	}
	return mc.innerClient.Delete(ctx, obj, opts...)
}

// Call Mock Update() Until Handled Or Delegate To Inner Client
func (mc *MockClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	for _, mockUpdate := range mc.mocks.UpdateFns {
		handled, err := mockUpdate(mc.innerClient, ctx, obj, opts...)
		if handled == Handled {
			return err
		}
	}
	return mc.innerClient.Update(ctx, obj, opts...)
}

// Call Mock Patch() Until Handled Or Delegate To Inner Client
func (mc *MockClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	for _, mockPatch := range mc.mocks.PatchFns {
		handled, err := mockPatch(mc.innerClient, ctx, obj, patch, opts...)
		if handled == Handled {
			return err
		}
	}
	return mc.innerClient.Patch(ctx, obj, patch, opts...)
}

// Call Mock DeleteAllOf() Until Handled Or Delegate To Inner Client
func (mc *MockClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	for _, mockDeleteAllOf := range mc.mocks.DeleteAllOfFns {
		handled, err := mockDeleteAllOf(mc.innerClient, ctx, obj, opts...)
		if handled == Handled {
			return err
		}
	}
	return mc.innerClient.DeleteAllOf(ctx, obj, opts...)
}

// Return The MockClient's MockStatusWriter
func (mc *MockClient) Status() client.StatusWriter {
	return mc.statusWriter
}

// Stop Mocking By Clearing All The Mocks
func (mc *MockClient) StopMocking() {
	mc.mocks = Mocks{}
}

//
// K8S Service Mock Functions
//

// MockGetFn Which Handles Any Kubernetes Service By Returning An Error
var MockGetFnServiceError MockGetFn = func(innerClient client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) (MockHandled, error) {
	if _, ok := obj.(*corev1.Service); ok {
		err := errors.New(MockGetFnServiceErrorMessage)
		return Handled, err
	}
	return Unhandled, nil
}

// Handle Any Kubernetes Services By Returning An Error
var MockCreateFnServiceError MockCreateFn = func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.CreateOption) (MockHandled, error) {
	if _, ok := obj.(*corev1.Service); ok {
		err := errors.New(MockCreateFnServiceErrorMessage)
		return Handled, err
	}
	return Unhandled, nil
}

// Handle Any KafkaChannel Services By Returning An Error
var MockCreateFnKafkaChannelServiceError MockCreateFn = func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.CreateOption) (MockHandled, error) {
	if _, ok := obj.(*corev1.Service); ok {
		if obj.(*corev1.Service).Name == kafkautil.AppendKafkaChannelServiceNameSuffix(ChannelName) {
			err := errors.New(MockCreateFnServiceErrorMessage)
			return Handled, err
		}
	}
	return Unhandled, nil
}

// Handle Any Channel Deployment Services By Returning An Error
var MockCreateFnChannelDeploymentServiceError MockCreateFn = func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.CreateOption) (MockHandled, error) {
	if _, ok := obj.(*corev1.Service); ok {
		if obj.(*corev1.Service).Name == ChannelDeploymentName {
			err := errors.New(MockCreateFnServiceErrorMessage)
			return Handled, err
		}
	}
	return Unhandled, nil
}

// Handle Any Dispatcher Services By Returning An Error
var MockCreateFnDispatcherServiceError MockCreateFn = func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.CreateOption) (MockHandled, error) {
	if _, ok := obj.(*corev1.Service); ok {
		if strings.HasSuffix(obj.(*corev1.Service).Name, "-dispatcher") {
			err := errors.New(MockCreateFnServiceErrorMessage)
			return Handled, err
		}
	}
	return Unhandled, nil
}

//
// K8S Deployment Mock Functions
//

// MockGetFn Which Handles Any Kubernetes Deployment By Returning An Error
var MockGetFnDeploymentError MockGetFn = func(innerClient client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) (MockHandled, error) {
	if _, ok := obj.(*appsv1.Deployment); ok {
		err := errors.New(MockGetFnDeploymentErrorMessage)
		return Handled, err
	}
	return Unhandled, nil
}

// Handle Any Kubernetes Deployment By Returning An Error
var MockCreateFnDeploymentError MockCreateFn = func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.CreateOption) (MockHandled, error) {
	if _, ok := obj.(*appsv1.Deployment); ok {
		err := errors.New(MockCreateFnDeploymentErrorMessage)
		return Handled, err
	}
	return Unhandled, nil
}

// Handle Any Channel Deployment By Returning An Error
var MockCreateFnChannelDeploymentError MockCreateFn = func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.CreateOption) (MockHandled, error) {
	if _, ok := obj.(*appsv1.Deployment); ok {
		if obj.(*appsv1.Deployment).Name == ChannelDeploymentName {
			err := errors.New(MockCreateFnDeploymentErrorMessage)
			return Handled, err
		}
	}
	return Unhandled, nil
}

// Handle Any Dispatcher Deployment By Returning An Error
var MockCreateFnDispatcherDeploymentError MockCreateFn = func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.CreateOption) (MockHandled, error) {
	if _, ok := obj.(*appsv1.Deployment); ok {
		if strings.HasSuffix(obj.(*appsv1.Deployment).Name, "-dispatcher") {
			err := errors.New(MockCreateFnDeploymentErrorMessage)
			return Handled, err
		}
	}
	return Unhandled, nil
}

//
// Knative Channel Mock Functions
//

// MockGetFn Which Handles Any Knative KafkaChannel By Returning An Error
var MockGetFnKafkaChannelError MockGetFn = func(innerClient client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) (MockHandled, error) {
	if _, ok := obj.(*kafkav1alpha1.KafkaChannel); ok {
		err := errors.New(MockGetFnKafkaChannelErrorMessage)
		return Handled, err
	}
	return Unhandled, nil
}

// MockUpdateFn Which Handles Any Knative KafkaChannel By Returning An Error
var MockUpdateFnKafkaChannelError MockUpdateFn = func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) (MockHandled, error) {
	if _, ok := obj.(*kafkav1alpha1.KafkaChannel); ok {
		err := errors.New(MockUpdateFnKafkaChannelErrorMessage)
		return Handled, err
	}
	return Unhandled, nil
}

//
// Knative Subscription Mock Functions
//

// MockGetFn Which Handles Any Knative Subscription By Returning An Error
var MockGetFnSubscriptionError MockGetFn = func(innerClient client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) (MockHandled, error) {
	if _, ok := obj.(*messagingv1alpha1.Subscription); ok {
		err := errors.New(MockGetFnSubscriptionErrorMessage)
		return Handled, err
	}
	return Unhandled, nil
}

//
// Mock ControllerRuntime StatusWriter
//

// The Set Of Mocks To Include When Testing The Controller
type StatusWriterMocks struct {
	UpdateFns []MockStatusUpdateFn
	PatchFns  []MockStatusPatchFn
}

// Define The Mock StatusWriter
type MockStatusWriter struct {
	innerClient client.Client
	mocks       StatusWriterMocks
}

// Create A New MockStatusWriter
func NewMockStatusWriter(innerClient client.Client, statusWriterMocks StatusWriterMocks) client.StatusWriter {
	return MockStatusWriter{
		innerClient: innerClient,
		mocks:       statusWriterMocks,
	}
}

// Verify The MockStatusWriter Implements The ControllerRuntime's StatusWriter Interface
var _ client.StatusWriter = &MockStatusWriter{}

// Mock The StatusWriter's Update() Function
func (msw MockStatusWriter) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	for _, mockUpdate := range msw.mocks.UpdateFns {
		handled, err := mockUpdate(msw.innerClient, ctx, obj, opts...)
		if handled == Handled {
			return err
		}
	}
	return msw.innerClient.Status().Update(ctx, obj, opts...)
}

// Mock The StatusWriter's Patch() Function
func (msw MockStatusWriter) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	for _, mockPatch := range msw.mocks.PatchFns {
		handled, err := mockPatch(msw.innerClient, ctx, obj, patch, opts...)
		if handled == Handled {
			return err
		}
	}
	return msw.innerClient.Status().Patch(ctx, obj, patch, opts...)
}

//
// Knative KafkaChannel Mock Functions
//

// MockStatusUpdateFn Which Handles Any Knative KafkaChannel By Returning An Error
var MockStatusUpdateFnChannelError MockStatusUpdateFn = func(innerClient client.Client, ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) (MockHandled, error) {
	if _, ok := obj.(*kafkav1alpha1.KafkaChannel); ok {
		err := errors.New(MockStatusUpdateFnChannelErrorMessage)
		return Handled, err
	}
	return Unhandled, nil
}

//
// Mock EventRecorder
//

// "k8s.io/client-go/tools/record"

// Define The Mock EventRecorder
type MockEventRecorder struct {
	events []corev1.Event
}

// Verify The MockEventRecorder Implements The ControllerRuntime's EventRecorder Interface
var _ record.EventRecorder = &MockEventRecorder{}

// Create A New MockEventRecorder
func NewMockEventRecorder() *MockEventRecorder {
	return &MockEventRecorder{
		events: make([]corev1.Event, 0),
	}
}

func (mer *MockEventRecorder) Event(object runtime.Object, eventType, reason, message string) {
	mer.recordEvent(mer.createEvent(eventType, reason, message))
}

func (mer *MockEventRecorder) Eventf(object runtime.Object, eventType, reason, messageFmt string, args ...interface{}) {
	mer.recordEvent(mer.createEvent(eventType, reason, fmt.Sprintf(messageFmt, args...)))
}

func (mer *MockEventRecorder) PastEventf(object runtime.Object, timestamp v1.Time, eventType, reason, messageFmt string, args ...interface{}) {
	mer.recordEvent(mer.createEvent(eventType, reason, fmt.Sprintf(messageFmt, args...)))
}

func (mer *MockEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventType, reason, messageFmt string, args ...interface{}) {
	mer.recordEvent(mer.createEvent(eventType, reason, fmt.Sprintf(messageFmt, args...)))
}

// Create A Sparse Event (Currently Only Validating The Type & Reason)
func (mer *MockEventRecorder) createEvent(eventType string, reason string, message string) corev1.Event {
	return corev1.Event{
		Reason:  reason,
		Message: message,
		Type:    eventType,
	}
}

// Record The Specified Event
func (mer *MockEventRecorder) recordEvent(event corev1.Event) {
	mer.events = append(mer.events, event)
}

// Remove Event With Specified Index From Tracking
func (mer *MockEventRecorder) deleteEvent(eventIndex int) {
	copy(mer.events[eventIndex:], mer.events[eventIndex+1:])
	mer.events = mer.events[:len(mer.events)-1]
}

// Determine Whether The Specified Event Was Recorded - If So Remove From Tracking To Facilitate Duplicate Events
func (mer *MockEventRecorder) EventRecorded(expectedEvent corev1.Event) bool {
	for eventIndex, actualEvent := range mer.events {
		if actualEvent.Type == expectedEvent.Type && actualEvent.Reason == expectedEvent.Reason {
			mer.deleteEvent(eventIndex)
			return true
		}
	}
	return false
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
	MockCreateTopicFunc func(context.Context, []kafka.TopicSpecification, ...kafka.CreateTopicsAdminOption) ([]kafka.TopicResult, error)
	MockDeleteTopicFunc func(context.Context, []string, ...kafka.DeleteTopicsAdminOption) ([]kafka.TopicResult, error)
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

package test

import (
	"errors"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaproducer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/producer"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	knativekafkalisters "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/listers/knativekafka/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

//
// Mock Confluent Producer
//

var _ kafkaproducer.ProducerInterface = &MockProducer{}

type MockProducer struct {
	config              *kafka.ConfigMap
	produce             chan *kafka.Message
	events              chan kafka.Event
	testResponseMessage *kafka.Message
	closed              bool
}

func NewMockProducer(topicName string) *MockProducer {

	testResponseMessage := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topicName,
			Partition: 1,
			Offset:    -1,
			Error:     nil,
		},
	}

	return &MockProducer{
		config:              nil,
		produce:             make(chan *kafka.Message, 1),
		events:              make(chan kafka.Event, 1),
		testResponseMessage: testResponseMessage,
		closed:              false,
	}
}

// String returns a human readable name for a Producer instance
func (p *MockProducer) String() string {
	return ""
}

func (p *MockProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	p.produce <- msg

	// Write back to the deliveryChan has to be done in a separate goroutine
	go func() {
		deliveryChan <- p.testResponseMessage
	}()

	return nil
}

func (p *MockProducer) Events() chan kafka.Event {
	return p.events
}

func (p *MockProducer) ProduceChannel() chan *kafka.Message {
	return p.produce
}

func (p *MockProducer) Len() int {
	return 0
}

func (p *MockProducer) Flush(timeoutMs int) int {
	return 0
}

func (p *MockProducer) Close() {
	close(p.events)
	close(p.produce)
	p.closed = true
}

func (p *MockProducer) Closed() bool {
	return p.closed
}

//
// Mock KafkaChannel Lister
//

var _ knativekafkalisters.KafkaChannelLister = &MockKafkaChannelLister{}

type MockKafkaChannelLister struct {
	name      string
	namespace string
	exists    bool
	ready     corev1.ConditionStatus
	err       bool
}

func NewMockKafkaChannelLister(name string, namespace string, exists bool, ready corev1.ConditionStatus, err bool) MockKafkaChannelLister {
	return MockKafkaChannelLister{
		name:      name,
		namespace: namespace,
		exists:    exists,
		ready:     ready,
		err:       err,
	}
}

func (m MockKafkaChannelLister) List(selector labels.Selector) (ret []*knativekafkav1alpha1.KafkaChannel, err error) {
	panic("implement me")
}

func (m MockKafkaChannelLister) KafkaChannels(namespace string) knativekafkalisters.KafkaChannelNamespaceLister {
	return NewMockKafkaChannelNamespaceLister(m.name, namespace, m.exists, m.ready, m.err)
}

//
// Mock KafkaChannel NamespaceLister
//

var _ knativekafkalisters.KafkaChannelNamespaceLister = &MockKafkaChannelNamespaceLister{}

type MockKafkaChannelNamespaceLister struct {
	name      string
	namespace string
	exists    bool
	ready     corev1.ConditionStatus
	err       bool
}

func NewMockKafkaChannelNamespaceLister(name string, namespace string, exists bool, ready corev1.ConditionStatus, err bool) MockKafkaChannelNamespaceLister {
	return MockKafkaChannelNamespaceLister{
		name:      name,
		namespace: namespace,
		exists:    exists,
		ready:     ready,
		err:       err,
	}
}

func (m MockKafkaChannelNamespaceLister) List(selector labels.Selector) (ret []*knativekafkav1alpha1.KafkaChannel, err error) {
	panic("implement me")
}

func (m MockKafkaChannelNamespaceLister) Get(name string) (*knativekafkav1alpha1.KafkaChannel, error) {
	if m.err {
		return nil, k8serrors.NewInternalError(errors.New("expected Unit Test error from MockKafkaChannelNamespaceLister"))
	} else if m.exists {
		return CreateKafkaChannel(m.name, m.namespace, m.ready), nil
	} else {
		return nil, k8serrors.NewNotFound(knativekafkav1alpha1.Resource("KafkaChannel"), name)
	}
}

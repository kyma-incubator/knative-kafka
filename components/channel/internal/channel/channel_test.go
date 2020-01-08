package channel

import (
	"errors"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaproducer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/producer"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test Constants
const (
	testBrokers      = "TestBrokers"
	testTopic        = "TestTopic"
	testClientId     = "TestClientId"
	testPartitionKey = "TestPartitionKey"
	testMessage      = "{ \"TestMessage\": \"TestMessage\"}"
	testUsername     = "TestUsername"
	testPassword     = "TestPassword"
)

var testCloudEvent cloudevents.Event

func init() {
	testCloudEvent = cloudevents.NewEvent(cloudevents.VersionV03)
	testCloudEvent.SetID("ABC-123")
	testCloudEvent.SetType("com.cloudevents.readme.sent")
	testCloudEvent.SetSource("http://localhost:8080/")
	testCloudEvent.SetDataContentType("application/json")
	testCloudEvent.SetData(testMessage)
}

// Setup A Test Logger (Ignore Unused Warning - This Updates The log.Logger Reference!)
var logger = log.TestLogger()

// Test The Whole Shebang - Create New Channel, Send Messages & Process Responses
func TestChannel(t *testing.T) {

	produceChannel := make(chan *kafka.Message, 1)
	eventsChannel := make(chan kafka.Event, 1)

	// String pointer hell
	topicName := testTopic
	testResponseMessage := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topicName,
			Partition: 1,
			Offset:    1,
			Error:     nil,
		},
	}

	var mockProducer kafkaproducer.ProducerInterface

	// Replace The NewProducer Wrapper To Provide Mock Producer & Defer Reset
	newProducerWrapperPlaceholder := kafkaproducer.NewProducerWrapper
	kafkaproducer.NewProducerWrapper = func(configMap *kafka.ConfigMap) (kafkaproducer.ProducerInterface, error) {
		verifyKafkaProducerConfig(t, configMap)
		mockProducer = MockProducer{configMap, produceChannel, eventsChannel, testResponseMessage}
		return mockProducer, nil
	}
	defer func() { kafkaproducer.NewProducerWrapper = newProducerWrapperPlaceholder }()

	// Create A New Channel
	config := ChannelConfig{
		Brokers:       testBrokers,
		Topic:         testTopic,
		ClientId:      testClientId,
		KafkaUsername: testUsername,
		KafkaPassword: testPassword,
	}
	testChannel := NewChannel(config)

	// Verify Channel Configuration
	verifyChannel(t, testChannel, testBrokers, testTopic, testClientId, testUsername, testPassword, mockProducer)

	// Send TestMessage To The Channel & Wait A Short Bit For It To Be Processed
	err := testChannel.SendMessage(testCloudEvent)
	assert.Nil(t, err)

	// Close The Channel
	testChannel.Close()
}

func TestChannelReturnsError(t *testing.T) {

	produceChannel := make(chan *kafka.Message, 1)
	eventsChannel := make(chan kafka.Event, 1)

	expectedError := errors.New("mock kafka failure")

	// String pointer hell
	topicName := testTopic
	testResponseMessage := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topicName,
			Partition: 1,
			Offset:    -1,
			Error:     expectedError,
		},
	}

	var mockProducer kafkaproducer.ProducerInterface

	// Replace The NewAsyncProducerFromClient Wrapper To Provide Mock AsyncProducer & Defer Reset
	newProducerWrapperPlaceholder := kafkaproducer.NewProducerWrapper
	kafkaproducer.NewProducerWrapper = func(configMap *kafka.ConfigMap) (kafkaproducer.ProducerInterface, error) {
		verifyKafkaProducerConfig(t, configMap)
		mockProducer = MockProducer{configMap, produceChannel, eventsChannel, testResponseMessage}
		return mockProducer, nil
	}
	defer func() { kafkaproducer.NewProducerWrapper = newProducerWrapperPlaceholder }()

	// Create A New Channel
	config := ChannelConfig{
		Brokers:       testBrokers,
		Topic:         testTopic,
		ClientId:      testClientId,
		KafkaUsername: testUsername,
		KafkaPassword: testPassword,
	}
	testChannel := NewChannel(config)

	// Verify Channel Configuration
	verifyChannel(t, testChannel, testBrokers, testTopic, testClientId, testUsername, testPassword, mockProducer)

	// Send TestMessage To The Channel & Wait A Short Bit For It To Be Processed
	err := testChannel.SendMessage(testCloudEvent)
	assert.NotNil(t, err)

	// Close The Channel
	testChannel.Close()
}

// Verify The Channel Is Configured As Specified
func verifyChannel(t *testing.T,
	channel *Channel,
	expectedBrokers string,
	expectedTopic string,
	expectedClientId string,
	expectedUsername string,
	expectedPassword string,
	expectedProducer kafkaproducer.ProducerInterface) {

	assert.NotNil(t, channel)
	assert.Equal(t, expectedBrokers, channel.Brokers)
	assert.Equal(t, expectedTopic, channel.Topic)
	assert.Equal(t, expectedClientId, channel.ClientId)
	assert.Equal(t, expectedProducer, channel.producer)
	assert.Equal(t, expectedUsername, channel.KafkaUsername)
	assert.Equal(t, expectedPassword, channel.KafkaPassword)
}

// Verify The Specified Kafka Producer Config
func verifyKafkaProducerConfig(t *testing.T, configMap *kafka.ConfigMap) {

	assert.NotNil(t, configMap)

	verifyConfigMapValue(t, configMap, KafkaProducerConfigPropertyBootstrapServers, testBrokers)
	verifyConfigMapValue(t, configMap, KafkaProducerConfigPropertyPartitioner, KafkaProducerConfigPropertyPartitionerValue)

	idempotenceProperty, err := configMap.Get(KafkaProducerConfigPropertyIdempotence, nil)
	assert.Nil(t, err)
	assert.False(t, idempotenceProperty.(bool))

	verifyConfigMapValue(t, configMap, KafkaProducerConfigPropertyUsername, testUsername)
	verifyConfigMapValue(t, configMap, KafkaProducerConfigPropertyPassword, testPassword)
}

func verifyConfigMapValue(t *testing.T, configMap *kafka.ConfigMap, key string, expected string) {
	property, err := configMap.Get(key, nil)
	assert.Nil(t, err)
	assert.Equal(t, expected, property.(string))
}

//
// Mock Confluent Producer
//

var _ kafkaproducer.ProducerInterface = &MockProducer{}

type MockProducer struct {
	config              *kafka.ConfigMap
	produce             chan *kafka.Message
	events              chan kafka.Event
	testResponseMessage *kafka.Message
}

// String returns a human readable name for a Producer instance
func (p MockProducer) String() string {
	return ""
}

func (p MockProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	p.produce <- msg

	// Write back to the deliveryChan has to be done in a separate goroutine
	go func() {
		deliveryChan <- p.testResponseMessage
	}()

	return nil
}

func (p MockProducer) Events() chan kafka.Event {
	return p.events
}

func (p MockProducer) ProduceChannel() chan *kafka.Message {
	return p.produce
}

func (p MockProducer) Len() int {
	return 0
}

func (p MockProducer) Flush(timeoutMs int) int {
	return 0
}

func (p MockProducer) Close() {
	close(p.events)
	close(p.produce)
}

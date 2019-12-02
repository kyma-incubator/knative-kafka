package dispatcher

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaconsumer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/consumer"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/dispatcher/internal/client"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// Test Constants
const (
	testBrokers                 = "TestBrokers"
	testTopic                   = "TestTopic"
	testPartition               = 33
	testOffset                  = "latest"
	testConcurrency             = 4
	testGroupId                 = "TestGroupId"
	testKey                     = "TestKey"
	testValue                   = "TestValue"
	testMessagesToSend          = 3
	testPollTimeoutMillis       = 500 // Not Used By Mock Consumer ; )
	testOffsetCommitCount       = 2
	testOffsetCommitDuration    = 50 * time.Millisecond // Small Durations For Testing!
	testOffsetCommitDurationMin = 50 * time.Millisecond // Small Durations For Testing!
	testUsername                = "TestUsername"
	testPassword                = "TestPassword"
)

// Test Data (Non-Constants)
var (
	topic              = testTopic
	testError          = kafka.NewError(1, "sample error", false)
	testNotification   = kafka.RevokedPartitions{}
	numberOfCalls      = 0
	numberOfCallsMutex = &sync.Mutex{}
)

// Setup A Test Logger (Ignore Unused Warning - This Updates The log.Logger Reference!)
var logger = log.TestLogger()

// Test All The Dispatcher Functionality
func TestDispatcher(t *testing.T) {

	// Initialize The Test HTTP Server Instance & URL
	var httpServer = getHttpServer(t)
	testSubscriberUri := httpServer.URL
	defer httpServer.Close()

	// Replace The NewClient Wrapper To Provide Mock Consumer & Defer Reset
	newConsumerWrapperPlaceholder := kafkaconsumer.NewConsumerWrapper
	kafkaconsumer.NewConsumerWrapper = func(configMap *kafka.ConfigMap) (kafkaconsumer.ConsumerInterface, error) {
		verifyConfigMapValue(t, configMap, "bootstrap.servers", testBrokers)
		verifyConfigMapValue(t, configMap, "group.id", testGroupId)
		verifyConfigMapValue(t, configMap, "sasl.username", testUsername)
		verifyConfigMapValue(t, configMap, "sasl.password", testPassword)
		return NewMockConsumer(), nil
	}
	defer func() { kafkaconsumer.NewConsumerWrapper = newConsumerWrapperPlaceholder }()

	httpClient := client.NewHttpClient(testSubscriberUri, false, 500, 5000)

	// Create A New Dispatcher
	dispatcherConfig := DispatcherConfig{
		Brokers:                     testBrokers,
		Topic:                       testTopic,
		Offset:                      testOffset,
		Concurrency:                 testConcurrency,
		GroupId:                     testGroupId,
		SubscriberUri:               testSubscriberUri,
		PollTimeoutMillis:           testPollTimeoutMillis,
		OffsetCommitCount:           testOffsetCommitCount,
		OffsetCommitDuration:        testOffsetCommitDuration,
		OffsetCommitDurationMinimum: testOffsetCommitDurationMin,
		Username:                    testUsername,
		Password:                    testPassword,
		Client:                      httpClient,
	}
	testDispatcher := NewDispatcher(dispatcherConfig)

	// Verify Dispatcher Configuration
	verifyDispatcher(t,
		testDispatcher,
		testBrokers,
		testTopic,
		testOffset,
		testConcurrency,
		testGroupId,
		testOffsetCommitCount,
		testOffsetCommitDuration,
		testUsername,
		testPassword)

	// Start Consumers
	testDispatcher.StartConsumers()

	// Verify Consumers Started
	verifyConsumersCount(t, testDispatcher.consumers)

	// Send Errors, Notifications, Messages To Each Consumer
	sendErrorToConsumers(t, testDispatcher.consumers, testError)
	sendNotificationToConsumers(t, testDispatcher.consumers, testNotification)
	sendMessagesToConsumers(t, testDispatcher.consumers, testMessagesToSend)

	// Wait For Consumers To Process Messages
	waitForConsumersToProcessEvents(t, testDispatcher.consumers)

	// Stop Consumers
	testDispatcher.StopConsumers()

	// Verify The Consumer Offset Commit Messages
	verifyConsumerCommits(t, testDispatcher.consumers)

	// Verify Consumers Stopped
	verifyConsumersClosed(t, testDispatcher.consumers)

	// Wait A Bit To See The Channel Closing Logs (lame but only for visual verification of non-blocking channel shutdown)
	time.Sleep(2 * time.Second)

	// Verify Mock HTTP Server Is Called The Correct Number Of Times
	numberOfCallsMutex.Lock()
	if numberOfCalls != (testConcurrency * (testMessagesToSend)) {
		t.Errorf("expected %+v HTTP calls but got: %+v", testConcurrency, numberOfCalls)
	}
	numberOfCallsMutex.Unlock()
}

// Verify A Specific ConfigMap Value
func verifyConfigMapValue(t *testing.T, configMap *kafka.ConfigMap, key string, expected string) {
	property, err := configMap.Get(key, nil)
	assert.Nil(t, err)
	assert.Equal(t, expected, property.(string))
}

// Return Test HTTP Server - Always Responds With Success
func getHttpServer(t *testing.T) *httptest.Server {
	httpServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "POST":

			// Update The Number Of Calls (Async Safe)
			numberOfCallsMutex.Lock()
			numberOfCalls++
			numberOfCallsMutex.Unlock()

			// Verify Request Content-Type Header
			requestContentTypeHeader := request.Header.Get("Content-Type")
			if requestContentTypeHeader != "application/json" {
				t.Errorf("expected HTTP request ContentType Header: application/json got: %+v", requestContentTypeHeader)
			}

			// Verify The Request Body
			bodyBytes, err := ioutil.ReadAll(request.Body)
			if err != nil {
				t.Errorf("expected to be able to read HTTP request Body without error: %+v", err)
			} else {
				bodyString := string(bodyBytes)
				if bodyString != testValue {
					t.Errorf("expected HTTP request Body: %+v got: %+v", testValue, bodyString)
				}
			}

			// Send Success Response (This is the default, but just to be explicit)
			responseWriter.WriteHeader(http.StatusOK)

		default:
			t.Errorf("expected HTTP request method: POST got: %+v", request.Method)

		}
	}))
	return httpServer
}

// Verify The Dispatcher Is Configured As Specified
func verifyDispatcher(t *testing.T,
	dispatcher *Dispatcher,
	expectedBrokers string,
	expectedTopic string,
	expectedOffset string,
	expectedConcurrency int64,
	expectedGroupId string,
	expectedOffsetCommitCount int64,
	expectedOffsetCommitDuration time.Duration,
	expectedUsername string,
	expectedPassword string) {

	assert.NotNil(t, dispatcher)
	assert.Equal(t, expectedBrokers, dispatcher.Brokers)
	assert.Equal(t, expectedTopic, dispatcher.Topic)
	assert.Equal(t, expectedOffset, dispatcher.Offset)
	assert.Equal(t, expectedConcurrency, dispatcher.Concurrency)
	assert.Equal(t, expectedGroupId, dispatcher.GroupId)
	assert.NotNil(t, dispatcher.consumers)
	assert.Len(t, dispatcher.consumers, 0)
	assert.Equal(t, expectedOffsetCommitCount, dispatcher.OffsetCommitCount)
	assert.Equal(t, expectedOffsetCommitDuration, dispatcher.OffsetCommitDuration)
	assert.Equal(t, expectedUsername, dispatcher.Username)
	assert.Equal(t, expectedPassword, dispatcher.Password)
}

// Verify The Appropriate Consumers Were Created
func verifyConsumersCount(t *testing.T, consumers []kafkaconsumer.ConsumerInterface) {
	assert.NotNil(t, consumers)
	assert.Len(t, consumers, testConcurrency)
	for _, consumer := range consumers {
		mockConsumer, ok := consumer.(*MockConsumer)
		assert.False(t, mockConsumer.closed)
		assert.True(t, ok)
	}
}

// Verify The Consumers Have All Been Closed
func verifyConsumersClosed(t *testing.T, consumers []kafkaconsumer.ConsumerInterface) {
	assert.NotNil(t, consumers)
	assert.Len(t, consumers, testConcurrency)
	for _, consumer := range consumers {
		mockConsumer, ok := consumer.(*MockConsumer)
		assert.True(t, mockConsumer.closed)
		assert.True(t, ok)
	}
}

// Verify The Consumers Have The Correct Commit Messages / Offsets
func verifyConsumerCommits(t *testing.T, consumers []kafkaconsumer.ConsumerInterface) {
	assert.NotNil(t, consumers)
	assert.Len(t, consumers, testConcurrency)
	for _, consumer := range consumers {
		mockConsumer, ok := consumer.(*MockConsumer)
		assert.True(t, mockConsumer.closed)
		assert.True(t, ok)
		for _, offset := range mockConsumer.commits {
			assert.NotNil(t, offset)
			assert.Equal(t, kafka.Offset(testMessagesToSend+1), offset)
		}
	}
}

// Send An Error To The Specified Consumers
func sendErrorToConsumers(t *testing.T, consumers []kafkaconsumer.ConsumerInterface, err kafka.Error) {
	for _, consumer := range consumers {
		mockConsumer, ok := consumer.(*MockConsumer)
		assert.True(t, ok)
		mockConsumer.sendMessage(err)
	}
}

// Send A Notification To The Specified Consumers
func sendNotificationToConsumers(t *testing.T, consumers []kafkaconsumer.ConsumerInterface, notification kafka.Event) {
	for _, consumer := range consumers {
		mockConsumer, ok := consumer.(*MockConsumer)
		assert.True(t, ok)
		mockConsumer.sendMessage(notification)
	}
}

// Send Some Messages To The Specified Consumers
func sendMessagesToConsumers(t *testing.T, consumers []kafkaconsumer.ConsumerInterface, count int) {
	for i := 0; i < count; i++ {
		message := createKafkaMessage(kafka.Offset(i + 1))
		for _, consumer := range consumers {
			mockConsumer, ok := consumer.(*MockConsumer)
			assert.True(t, ok)
			mockConsumer.sendMessage(message)
		}
	}
}

// Create A New Kafka Message With Specified Offset
func createKafkaMessage(offset kafka.Offset) *kafka.Message {
	return &kafka.Message{
		Key:       []byte(testKey),
		Value:     []byte(testValue),
		Timestamp: time.Now(),
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: testPartition,
			Offset:    offset,
		},
	}
}

// Wait For The Consumers To Process All The Events
func waitForConsumersToProcessEvents(t *testing.T, consumers []kafkaconsumer.ConsumerInterface) {
	startTime := time.Now()
	for _, consumer := range consumers {
		mockConsumer, ok := consumer.(*MockConsumer)
		assert.True(t, ok)
		for {
			mockConsumer.offsetsMutex.Lock()
			if mockConsumer.getCommits()[testPartition] >= testMessagesToSend+1 {
				break
			} else if time.Now().Sub(startTime) > (4 * time.Second) {
				assert.FailNow(t, "Timed-out Waiting For Consumers To Process Events")
			} else {
				time.Sleep(100 * time.Millisecond)
			}
			mockConsumer.offsetsMutex.Unlock()
		}
	}
}

//
// Mock ConsumerInterface Implementation
//

var _ kafkaconsumer.ConsumerInterface = &MockConsumer{}

func NewMockConsumer() *MockConsumer {
	return &MockConsumer{
		events:       make(chan kafka.Event),
		commits:      make(map[int32]kafka.Offset),
		offsets:      make(map[int32]kafka.Offset),
		offsetsMutex: &sync.Mutex{},
		closed:       false,
	}
}

type MockConsumer struct {
	events             chan kafka.Event
	commits            map[int32]kafka.Offset
	offsets            map[int32]kafka.Offset
	offsetsMutex       *sync.Mutex
	eventsChanEnable   bool
	readerTermChan     chan bool
	rebalanceCb        kafka.RebalanceCb
	appReassigned      bool
	appRebalanceEnable bool // config setting
	closed             bool
}

func (mc *MockConsumer) OffsetsForTimes(times []kafka.TopicPartition, timeoutMs int) (offsets []kafka.TopicPartition, err error) {
	return nil, nil
}

func (mc *MockConsumer) GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error) {
	return nil, nil
}

func (mc *MockConsumer) CommitOffsets(offsets []kafka.TopicPartition) ([]kafka.TopicPartition, error) {
	return nil, nil
}

func (mc *MockConsumer) QueryWatermarkOffsets(topic string, partition int32, timeoutMs int) (low, high int64, err error) {
	return 0, 0, nil
}

func (mc *MockConsumer) Poll(timeout int) kafka.Event {
	// Non-Blocking Event Forwarding (Timeout Ignored)
	select {
	case event := <-mc.events:
		return event
	default:
		return nil
	}
}

func (mc *MockConsumer) CommitMessage(message *kafka.Message) ([]kafka.TopicPartition, error) {
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

func (mc *MockConsumer) getCommits() map[int32]kafka.Offset {
	return mc.commits
}

func (mc *MockConsumer) Assignment() (partitions []kafka.TopicPartition, err error) {
	return nil, nil
}

func (mc *MockConsumer) Commit() ([]kafka.TopicPartition, error) {
	mc.offsetsMutex.Lock()
	for partition, offset := range mc.offsets {
		mc.commits[partition] = offset
	}
	mc.offsetsMutex.Unlock()
	return nil, nil
}

func (mc *MockConsumer) StoreOffsets(offsets []kafka.TopicPartition) (storedOffsets []kafka.TopicPartition, err error) {
	mc.offsetsMutex.Lock()
	for _, partition := range offsets {
		mc.offsets[partition.Partition] = partition.Offset
	}
	mc.offsetsMutex.Unlock()
	return nil, nil
}

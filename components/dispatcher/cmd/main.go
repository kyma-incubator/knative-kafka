package main

import (
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/prometheus"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/util"
	"github.com/kyma-incubator/knative-kafka/components/dispatcher/internal/client"
	"github.com/kyma-incubator/knative-kafka/components/dispatcher/internal/dispatcher"
	"go.uber.org/zap"

	"os"
	"strconv"
	"syscall"
	"time"
)

// Constants
const (
	DefaultKafkaConsumerOffset                     = "latest"
	DefaultKafkaConsumerPollTimeoutMillis          = 500 // Timeout Millis When Polling For Events
	MinimumKafkaConsumerOffsetCommitDurationMillis = 250 // Azure EventHubs Restrict To 250ms Between Offset Commits
)

// Variables
var (
	logger *zap.Logger
	d      *dispatcher.Dispatcher
)

// The Main Function (Go Command)
func main() {

	// Initialize The Logger
	logger = log.Logger()

	// Load Environment Variables
	metricsPort := os.Getenv("METRICS_PORT")
	subscriberUri := os.Getenv("SUBSCRIBER_URI")
	rawExpBackoff, expBackoffPresent := os.LookupEnv("EXPONENTIAL_BACKOFF")
	exponentialBackoff, _ := strconv.ParseBool(rawExpBackoff)
	maxRetryTime, _ := strconv.ParseInt(os.Getenv("MAX_RETRY_TIME"), 10, 64)
	initialRetryInterval, _ := strconv.ParseInt(os.Getenv("INITIAL_RETRY_INTERVAL"), 10, 64)
	kafkaGroupId := os.Getenv("KAFKA_GROUP_ID")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	kafkaConsumers, _ := strconv.ParseInt(os.Getenv("KAFKA_CONSUMERS"), 10, 64)
	kafkaUsername := os.Getenv("KAFKA_USERNAME")
	kafkaPassword := os.Getenv("KAFKA_PASSWORD")
	kafkaPasswordLog := ""
	kafkaOffsetCommitMessageCount, _ := strconv.ParseInt(os.Getenv("KAFKA_OFFSET_COMMIT_MESSAGE_COUNT"), 10, 64)
	kafkaOffsetCommitDurationMillis, _ := strconv.ParseInt(os.Getenv("KAFKA_OFFSET_COMMIT_DURATION_MILLIS"), 10, 64)

	if len(kafkaPassword) > 0 {
		kafkaPasswordLog = "*************"
	}

	// Log Environment Variables
	logger.Info("Environment Variables",
		zap.String("METRICS_PORT", metricsPort),
		zap.String("SUBSCRIBER_URI", subscriberUri),
		zap.Bool("EXPONENTIAL_BACKOFF", exponentialBackoff),
		zap.Int64("INITIAL_RETRY_INTERVAL", initialRetryInterval),
		zap.Int64("MAX_RETRY_TIME", maxRetryTime),
		zap.String("KAFKA_GROUP_ID", kafkaGroupId),
		zap.String("KAFKA_BROKERS", kafkaBrokers),
		zap.String("KAFKA_TOPIC", kafkaTopic),
		zap.Int64("KAFKA_CONSUMERS", kafkaConsumers),
		zap.Int64("KAFKA_OFFSET_COMMIT_MESSAGE_COUNT", kafkaOffsetCommitMessageCount),
		zap.Int64("KAFKA_OFFSET_COMMIT_DURATION_MILLIS", kafkaOffsetCommitDurationMillis),
		zap.String("KAFKA_USERNAME", kafkaUsername),
		zap.String("KAFKA_PASSWORD", kafkaPasswordLog))

	// Validate Required Environment Variables
	if len(metricsPort) == 0 ||
		len(subscriberUri) == 0 ||
		maxRetryTime <= 0 ||
		initialRetryInterval <= 0 ||
		!expBackoffPresent ||
		len(kafkaGroupId) == 0 ||
		len(kafkaBrokers) == 0 ||
		len(kafkaTopic) == 0 ||
		kafkaConsumers <= 0 ||
		kafkaOffsetCommitMessageCount <= 0 ||
		kafkaOffsetCommitDurationMillis <= 0 {
		logger.Fatal("Invalid / Missing Environment Variables - Terminating")
	}

	// Start The Prometheus Metrics Server (Prometheus)
	metricsServer := prometheus.NewMetricsServer(metricsPort, "/metrics")
	metricsServer.Start()

	// Create HTTP Client With Retry Settings
	c := client.NewHttpClient(subscriberUri, exponentialBackoff, initialRetryInterval, maxRetryTime)

	// Create The Dispatcher With Specified Configuration
	dispatcherConfig := dispatcher.DispatcherConfig{
		SubscriberUri:               subscriberUri,
		Brokers:                     kafkaBrokers,
		Topic:                       kafkaTopic,
		Offset:                      DefaultKafkaConsumerOffset,
		Concurrency:                 kafkaConsumers,
		GroupId:                     kafkaGroupId,
		PollTimeoutMillis:           DefaultKafkaConsumerPollTimeoutMillis,
		OffsetCommitCount:           kafkaOffsetCommitMessageCount,
		OffsetCommitDuration:        time.Duration(kafkaOffsetCommitDurationMillis) * time.Millisecond,
		OffsetCommitDurationMinimum: MinimumKafkaConsumerOffsetCommitDurationMillis * time.Millisecond,
		Username:                    kafkaUsername,
		Password:                    kafkaPassword,
		Client:                      c,
	}
	d = dispatcher.NewDispatcher(dispatcherConfig)

	// Create Consumers & Connect To Kafka
	d.StartConsumers()

	// Block Waiting Termination Signal - Either SIGINT (manual ctrl-c) Or SIGTERM (sent by K8S)
	util.WaitForSignal(logger, syscall.SIGINT, syscall.SIGTERM)

	// Close Consumer Connections
	d.StopConsumers()

	// Shutdown The Prometheus Metrics Server
	metricsServer.Stop()
}

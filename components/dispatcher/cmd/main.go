package main

import (
	"flag"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/prometheus"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/informers/externalversions"
	"github.com/kyma-incubator/knative-kafka/components/dispatcher/internal/client"
	"github.com/kyma-incubator/knative-kafka/components/dispatcher/internal/controller"
	dispatch "github.com/kyma-incubator/knative-kafka/components/dispatcher/internal/dispatcher"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kncontroller "knative.dev/pkg/controller"
	"knative.dev/pkg/signals"
	"os"
	"strconv"
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
	logger     *zap.Logger
	dispatcher *dispatch.Dispatcher
	masterURL  = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
)

// The Main Function (Go Command)
func main() {

	// Initialize The Logger
	logger = log.Logger()

	// Load Environment Variables
	metricsPort := os.Getenv("METRICS_PORT")
	rawExpBackoff, expBackoffPresent := os.LookupEnv("EXPONENTIAL_BACKOFF")
	exponentialBackoff, _ := strconv.ParseBool(rawExpBackoff)
	maxRetryTime, _ := strconv.ParseInt(os.Getenv("MAX_RETRY_TIME"), 10, 64)
	initialRetryInterval, _ := strconv.ParseInt(os.Getenv("INITIAL_RETRY_INTERVAL"), 10, 64)
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	channelKey := os.Getenv("CHANNEL_KEY")
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
		zap.Bool("EXPONENTIAL_BACKOFF", exponentialBackoff),
		zap.Int64("INITIAL_RETRY_INTERVAL", initialRetryInterval),
		zap.Int64("MAX_RETRY_TIME", maxRetryTime),
		zap.String("KAFKA_BROKERS", kafkaBrokers),
		zap.String("KAFKA_TOPIC", kafkaTopic),
		zap.Int64("KAFKA_OFFSET_COMMIT_MESSAGE_COUNT", kafkaOffsetCommitMessageCount),
		zap.Int64("KAFKA_OFFSET_COMMIT_DURATION_MILLIS", kafkaOffsetCommitDurationMillis),
		zap.String("KAFKA_USERNAME", kafkaUsername),
		zap.String("KAFKA_PASSWORD", kafkaPasswordLog),
		zap.String("CHANNEL_KEY", channelKey))

	// Validate Required Environment Variables
	if len(metricsPort) == 0 ||
		maxRetryTime <= 0 ||
		initialRetryInterval <= 0 ||
		!expBackoffPresent ||
		len(kafkaBrokers) == 0 ||
		len(kafkaTopic) == 0 ||
		len(channelKey) == 0 ||
		kafkaOffsetCommitMessageCount <= 0 ||
		kafkaOffsetCommitDurationMillis <= 0 {
		logger.Fatal("Invalid / Missing Environment Variables - Terminating")
	}

	// Start The Prometheus Metrics Server (Prometheus)
	metricsServer := prometheus.NewMetricsServer(metricsPort, "/metrics")
	metricsServer.Start()

	// Create HTTP Client With Retry Settings
	ceClient := client.NewRetriableCloudEventClient(exponentialBackoff, initialRetryInterval, maxRetryTime)

	// Create The Dispatcher With Specified Configuration
	dispatcherConfig := dispatch.DispatcherConfig{
		Brokers:                     kafkaBrokers,
		Topic:                       kafkaTopic,
		Offset:                      DefaultKafkaConsumerOffset,
		PollTimeoutMillis:           DefaultKafkaConsumerPollTimeoutMillis,
		OffsetCommitCount:           kafkaOffsetCommitMessageCount,
		OffsetCommitDuration:        time.Duration(kafkaOffsetCommitDurationMillis) * time.Millisecond,
		OffsetCommitDurationMinimum: MinimumKafkaConsumerOffsetCommitDurationMillis * time.Millisecond,
		Username:                    kafkaUsername,
		Password:                    kafkaPassword,
		Client:                      ceClient,
		ChannelKey:                  channelKey,
	}
	dispatcher = dispatch.NewDispatcher(dispatcherConfig)

	config, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		logger.Error("Error building kubeconfig", zap.Error(err))
		return
	}

	stopCh := signals.SetupSignalHandler()

	const numControllers = 1
	config.QPS = numControllers * rest.DefaultQPS
	config.Burst = numControllers * rest.DefaultBurst
	kafkaClientSet := versioned.NewForConfigOrDie(config)
	kubeClient := kubernetes.NewForConfigOrDie(config)
	kafkaInformerFactory := externalversions.NewSharedInformerFactory(kafkaClientSet, kncontroller.DefaultResyncPeriod)

	// Create KafkaChannel Informer
	kafkaChannelInformer := kafkaInformerFactory.Knativekafka().V1alpha1().KafkaChannels()

	// Construct Array Of Controllers, In Our Case Just the One
	controllers := [...]*kncontroller.Impl{
		controller.NewController(
			logger,
			dispatcher,
			kafkaChannelInformer,
			kubeClient,
			kafkaClientSet,
			stopCh,
		),
	}

	// Start The Informers
	logger.Info("Starting informers.")
	if err := kncontroller.StartInformers(stopCh, kafkaChannelInformer.Informer()); err != nil {
		logger.Error("Failed to start informers", zap.Error(err))
		return
	}

	logger.Info("Starting controllers.")
	kncontroller.StartAll(stopCh, controllers[:]...)

	<-stopCh
	// Close Consumer Connections
	dispatcher.StopConsumers()

	// Shutdown The Prometheus Metrics Server
	metricsServer.Stop()
}

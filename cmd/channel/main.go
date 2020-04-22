package main

import (
	"context"
	"flag"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/kyma-incubator/knative-kafka/pkg/channel/channel"
	"github.com/kyma-incubator/knative-kafka/pkg/channel/constants"
	"github.com/kyma-incubator/knative-kafka/pkg/channel/env"
	channelhealth "github.com/kyma-incubator/knative-kafka/pkg/channel/health"
	"github.com/kyma-incubator/knative-kafka/pkg/channel/producer"
	commonk8s "github.com/kyma-incubator/knative-kafka/pkg/common/k8s"
	kafkautil "github.com/kyma-incubator/knative-kafka/pkg/common/kafka/util"
	"github.com/kyma-incubator/knative-kafka/pkg/common/prometheus"
	"go.uber.org/zap"
	eventingchannel "knative.dev/eventing/pkg/channel"
	"knative.dev/pkg/logging"
	nethttp "net/http"
)

// Variables
var (
	logger     *zap.Logger
	masterURL  = flag.String("masterurl", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
)

// The Main Function (Go Command)
func main() {

	// Parse The Flags For Local Development Usage
	flag.Parse()

	// Initialize A Knative Injection Lite Context (K8S Client & Logger)
	ctx := commonk8s.LoggingContext(context.Background(), constants.Component, *masterURL, *kubeconfig)

	// Get The Logger From The Context & Defer Flushing Any Buffered Log Entries On Exit
	logger = logging.FromContext(ctx).Desugar()
	defer func() { _ = logger.Sync() }()

	// Load Environment Variables
	environment, err := env.GetEnvironment(logger)
	if err != nil {
		logger.Fatal("Invalid / Missing Environment Variables - Terminating")
		return // Quiet The Compiler ; )
	}

	// Start The Liveness And Readiness Servers
	healthServer := channelhealth.NewChannelHealthServer(environment.HealthPort)
	healthServer.Start(logger)

	// Start The Prometheus Metrics Server (Prometheus)
	metricsServer := prometheus.NewMetricsServer(logger, environment.MetricsPort, "/metrics")
	metricsServer.Start()

	// Initialize The KafkaChannel Lister Used To Validate Events
	err = channel.InitializeKafkaChannelLister(ctx, *masterURL, *kubeconfig, healthServer)
	if err != nil {
		logger.Fatal("Failed To Initialize KafkaChannel Lister", zap.Error(err))
	}

	// Initialize The Kafka Producer In Order To Start Processing Status Events
	err = producer.InitializeProducer(logger, environment.KafkaBrokers, environment.KafkaUsername, environment.KafkaPassword, metricsServer, healthServer)
	if err != nil {
		logger.Fatal("Failed To Initialize Kafka Producer", zap.Error(err))
	}

	// Create A New Knative Eventing MessageReceiver (Parses The Channel From The Host Header)
	messageReceiver, err := eventingchannel.NewMessageReceiver(handleMessage, logger, eventingchannel.ResolveMessageChannelFromHostHeader(eventingchannel.ParseChannel))
	if err != nil {
		logger.Fatal("Failed To Create MessageReceiver", zap.Error(err))
	}

	// Set The Liveness Flag - Readiness Is Set By Individual Components
	healthServer.SetAlive(true)

	// Start The Message Receiver (Blocking)
	err = messageReceiver.Start(ctx)
	if err != nil {
		logger.Error("Failed To Start MessageReceiver", zap.Error(err))
	}

	// Reset The Liveness and Readiness Flags In Preparation For Shutdown
	healthServer.Shutdown()

	// Close The K8S KafkaChannel Lister & The Kafka Producer
	channel.Close()
	producer.Close()

	// Shutdown The Prometheus Metrics Server
	metricsServer.Stop()

	// Stop The Liveness And Readiness Servers
	healthServer.Stop(logger)
}

// CloudEvent Message Handler - Converts To KafkaMessage And Produces To Channel's Kafka Topic
func handleMessage(ctx context.Context, channelReference eventingchannel.ChannelReference, message binding.Message, transformers []binding.Transformer, _ nethttp.Header) error {

	// Note - The context provided here is a different context from the one created in main() and does not have our logger instance.
	logger.Debug("~~~~~~~~~~~~~~~~~~~~  Processing Request  ~~~~~~~~~~~~~~~~~~~~")
	logger.Debug("Received Message", zap.Any("Message", message), zap.Any("ChannelReference", channelReference))

	//
	// Convert The CloudEvents Binding Message To A CloudEvent
	//
	// TODO - It is potentially inefficient to take the CloudEvent binding/Message and convert it into a CloudEvent,
	//        just so that it can then be further transformed into a Confluent KafkaMessage.  The current implementation
	//        is based on CloudEvent Events, however, and until a "protocol" implementation for Confluent Kafka exists
	//        this is the simplest path forward.  Once such a protocol implementation exists, it would be more efficient
	//        to convert directly from the binding/Message to the protocol/Message.
	//
	cloudEvent, err := binding.ToEvent(ctx, message, transformers...)
	if err != nil {
		logger.Error("Failed To Convert Message To CloudEvent", zap.Error(err))
		return err
	}

	// Trim The "-kafkachannel" Suffix From The Service Name (Added Only To Workaround Kyma Naming Conflict)
	channelReference.Name = kafkautil.TrimKafkaChannelServiceNameSuffix(channelReference.Name)

	// Validate The KafkaChannel Prior To Producing Kafka Message
	err = channel.ValidateKafkaChannel(channelReference)
	if err != nil {
		logger.Warn("Unable To Validate ChannelReference", zap.Any("ChannelReference", channelReference), zap.Error(err))
		return err
	}

	// Send The Event To The Appropriate Channel/Topic
	err = producer.ProduceKafkaMessage(cloudEvent, channelReference)
	if err != nil {
		logger.Error("Failed To Produce Kafka Message", zap.Error(err))
		return err
	}

	// Return Success
	return nil
}

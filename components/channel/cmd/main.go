package main

import (
	"context"
	"flag"
	"github.com/cloudevents/sdk-go"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/channel"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/env"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/producer"
	kafkautil "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/util"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/prometheus"
	"go.uber.org/zap"
	eventingChannel "knative.dev/eventing/pkg/channel"
	"knative.dev/eventing/pkg/kncloudevents"
)

// Variables
var (
	logger     *zap.Logger
	masterURL  = flag.String("masterurl", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
)

// The Main Function (Go Command)
func main() {

	// Initialize The Logger
	logger = log.Logger()

	// Parse The Flags For Local Development Usage
	flag.Parse()

	// Load Environment Variables
	environment, err := env.GetEnvironment()
	if err != nil {
		logger.Fatal("Invalid / Missing Environment Variables - Terminating")
		return // Quiet The Compiler ; )
	}

	// Start The Prometheus Metrics Server (Prometheus)
	metricsServer := prometheus.NewMetricsServer(environment.MetricsPort, "/metrics")
	metricsServer.Start()

	// Initialize The KafkaChannel Lister Used To Validate Events
	err = channel.InitializeKafkaChannelLister(*masterURL, *kubeconfig)
	if err != nil {
		logger.Fatal("Failed To Initialize KafkaChannel Lister", zap.Error(err))
	}

	// Initialize The Kafka Producer In Order To Start Processing Status Events
	err = producer.InitializeProducer(environment.KafkaBrokers, environment.KafkaUsername, environment.KafkaPassword, metricsServer)
	if err != nil {
		logger.Fatal("Failed To Initialize Kafka Producer", zap.Error(err))
	}

	// The EventReceiver is responsible for processing the context (headers and binary/json content) of each request,
	// and then passing the context, channel details, and the constructed CloudEvent event to our handleEvent() function.
	eventReceiver, err := eventingChannel.NewEventReceiver(handleEvent, logger)
	if err != nil {
		logger.Fatal("Failed To Create Knative EventReceiver", zap.Error(err))
	}

	// The Knative CloudEvent Client handles the mux http server setup (middlewares and transport options) and invokes
	// the eventReceiver. Although the NewEventReceiver method above will also invoke kncloudevents.NewDefaultClient
	// internally, that client goes unused when using the ServeHTTP on the eventReceiver.
	//
	// IMPORTANT: Because the kncloudevents package does not allow injecting modified configuration,
	//            we can't override the default port being used (8080).
	knCloudEventClient, err := kncloudevents.NewDefaultClient()
	if err != nil {
		logger.Fatal("Failed To Create Knative CloudEvent Client", zap.Error(err))
	}

	// Start Receiving Events (Blocking Call :)
	err = knCloudEventClient.StartReceiver(context.Background(), eventReceiver.ServeHTTP)
	if err != nil {
		logger.Error("Failed To Start Event Receiver", zap.Error(err))
	}

	// Close The K8S KafkaChannel Lister & The Kafka Producer
	channel.Close()
	producer.Close()

	// Shutdown The Prometheus Metrics Server
	metricsServer.Stop()
}

// Handler For Receiving Cloud Events And Sending The Event To Kafka
func handleEvent(ctx context.Context, channelReference eventingChannel.ChannelReference, cloudEvent cloudevents.Event) error {

	logger.Debug("~~~~~~~~~~~~~~~~~~~~  Processing Request  ~~~~~~~~~~~~~~~~~~~~")
	logger.Debug("Received Cloud Event", zap.Any("CloudEvent", cloudEvent), zap.Any("ChannelReference", channelReference))

	// Trim The "-kafkachannel" Suffix From The Service Name (Added Only To Workaround Kyma Naming Conflict)
	channelReference.Name = kafkautil.TrimKafkaChannelServiceNameSuffix(channelReference.Name)

	// Validate The KafkaChannel Prior To Producing Kafka Message
	err := channel.ValidateKafkaChannel(channelReference)
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

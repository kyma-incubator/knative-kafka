package main

import (
	"context"
	"github.com/cloudevents/sdk-go"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/channel"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/prometheus"
	"go.uber.org/zap"
	eventingChannel "knative.dev/eventing/pkg/channel"
	"knative.dev/eventing/pkg/kncloudevents"
	"os"
)

// Variables
var (
	logger *zap.Logger
	c      *channel.Channel
)

// The Main Function (Go Command)
func main() {

	// Initialize The Logger
	logger = log.Logger()

	// Load Environment Variables
	metricsPort := os.Getenv("METRICS_PORT")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	clientId := os.Getenv("CLIENT_ID")
	kafkaUsername := os.Getenv("KAFKA_USERNAME")
	kafkaPass := os.Getenv("KAFKA_PASSWORD")
	kafkaPassLog := ""

	if len(kafkaPass) > 0 {
		kafkaPassLog = "*************"
	}

	// Log Environment Variables
	logger.Info("Environment Variables",
		zap.String("METRICS_PORT", metricsPort),
		zap.Any("KAFKA_BROKERS", kafkaBrokers),
		zap.String("KAFKA_TOPIC", kafkaTopic),
		zap.String("CLIENT_ID", clientId),
		zap.String("KAFKA_USERNAME", kafkaUsername),
		zap.String("KAFKA_PASSWORD", kafkaPassLog))

	// Validate Required Environment Variables
	if len(metricsPort) == 0 ||
		len(kafkaBrokers) == 0 ||
		len(kafkaTopic) == 0 ||
		len(clientId) == 0 {
		logger.Fatal("Invalid / Missing Environment Variables - Terminating")
	}

	ctx := context.Background()

	// The EventReceiver is responsible for processing the context
	// (headers and binary/json content) of each request and then
	// passing the context, channel details, and the constructed
	// CloudEvent event to our handleEvent function.
	eventReceiver, err := eventingChannel.NewEventReceiver(handleEvent, logger)
	if err != nil {
		logger.Error("failed to create the event_receiver", zap.Error(err))
		return
	}

	// The Knative CloudEvent Client handles the mux http server
	// setup (middlewares and transport options) and invokes
	// the eventReceiver. Althought the NewEventReceiver method
	// above will also invoke kncloudevents.NewDefaultClient
	// internally, that client goes unused when using the ServeHTTP
	// on the eventReceiver.
	// IMPORTANT: Because the kncloudevents package does not allow
	// injecting modified configuration, we can't override the
	// default port being used, 8080.
	knCloudEventClient, err := kncloudevents.NewDefaultClient()
	if err != nil {
		logger.Error("failed to create knative cloud event client", zap.Error(err))
		return
	}

	// Start The Prometheus Metrics Server (Prometheus)
	metricsServer := prometheus.NewMetricsServer(metricsPort, "/metrics")
	metricsServer.Start()

	// Create The Channel With Specified Configuration
	channelConfig := channel.ChannelConfig{
		Brokers:       kafkaBrokers,
		Topic:         kafkaTopic,
		ClientId:      clientId,
		KafkaUsername: kafkaUsername,
		KafkaPassword: kafkaPass,
	}
	c = channel.NewChannel(channelConfig)

	// Start Receiving Events
	knCloudEventClient.StartReceiver(ctx, eventReceiver.ServeHTTP)

	// Close The Channel
	c.Close()

	// Shutdown The Prometheus Metrics Server
	metricsServer.Stop()
}

// Handler For Receiving Cloud Events And Sending The Event To Kafka
func handleEvent(ctx context.Context, channelReference eventingChannel.ChannelReference, event cloudevents.Event) error {

	logger.Debug("~~~~~~~~~~~~~~~~~~~~  Processing Request  ~~~~~~~~~~~~~~~~~~~~")
	logger.Debug("Received Cloud Event", zap.Any("Event", event))

	err := c.SendMessage(event)
	if err != nil {
		logger.Error("Unable To Send Message To Kafka", zap.Error(err))
		return err
	}

	return nil
}

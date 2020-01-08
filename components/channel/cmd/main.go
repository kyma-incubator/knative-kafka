package main

import (
	"context"
	"github.com/cloudevents/sdk-go"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/channel"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/prometheus"
	"go.uber.org/zap"
	"os"
	"strconv"
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
	httpPort := os.Getenv("HTTP_PORT")
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
		zap.String("HTTP_PORT", httpPort),
		zap.Any("KAFKA_BROKERS", kafkaBrokers),
		zap.String("KAFKA_TOPIC", kafkaTopic),
		zap.String("CLIENT_ID", clientId),
		zap.String("KAFKA_USERNAME", kafkaUsername),
		zap.String("KAFKA_PASSWORD", kafkaPassLog))

	// Validate Required Environment Variables
	if len(httpPort) == 0 ||
		len(metricsPort) == 0 ||
		len(kafkaBrokers) == 0 ||
		len(kafkaTopic) == 0 ||
		len(clientId) == 0 {
		logger.Fatal("Invalid / Missing Environment Variables - Terminating")
	}

	ctx := context.Background()
	port, _ := strconv.Atoi(httpPort)

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithPort(port),
		cloudevents.WithPath("/"),
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV03),
	)
	if err != nil {
		logger.Error("failed to create transport", zap.Error(err))
		return
	}

	cloudEventClient, err := cloudevents.NewClient(t)
	if err != nil {
		logger.Error("failed to create client", zap.Error(err))
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
	cloudEventClient.StartReceiver(ctx, handleEvent)

	// Close The Channel
	c.Close()

	// Shutdown The Prometheus Metrics Server
	metricsServer.Stop()
}

// Handler For Receiving Cloud Events And Sending The Event To Kafka
func handleEvent(ctx context.Context, event cloudevents.Event) error {

	logger.Debug("~~~~~~~~~~~~~~~~~~~~  Processing Request  ~~~~~~~~~~~~~~~~~~~~")
	logger.Debug("Received Cloud Event", zap.Any("Event", event))

	err := c.SendMessage(event)
	if err != nil {
		logger.Error("Unable To Send Message To Kafka", zap.Error(err))
		return err
	}

	return nil
}

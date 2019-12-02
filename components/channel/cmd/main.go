package main

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/channel/internal/channel"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/prometheus"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/util"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"syscall"
)

// Channel / Producer Constants
const (
	PartitionKeyHttpHeader = "ce-comSAPObjectID"
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

	// Create The Server For Configured HTTP Port & Register Request Handler
	server := &http.Server{Addr: ":" + httpPort, Handler: nil}
	http.HandleFunc("/", handleRequests)

	// Start The HTTP Server Listening For Requests In A Go Routine
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Info("HTTP ListenAndServe Returned Error", zap.Error(err)) // Could Just Be The Normal Shutdown
		}
	}()

	// Block Waiting Termination Signal - Either SIGINT (manual ctrl-c) Or SIGTERM (sent by K8S)
	util.WaitForSignal(logger, syscall.SIGINT, syscall.SIGTERM)

	// Stop The HTTP Server
	err := server.Shutdown(context.TODO())
	if err != nil {
		logger.Error("Failed To Shutdown HTTP Server", zap.Error(err))
	}

	// Close The Channel
	c.Close()

	// Shutdown The Prometheus Metrics Server
	metricsServer.Stop()
}

// HTTP Request Handler
func handleRequests(responseWriter http.ResponseWriter, request *http.Request) {

	// Log Request Separator
	logger.Debug("~~~~~~~~~~~~~~~~~~~~  Processing Request  ~~~~~~~~~~~~~~~~~~~~", zap.String("Request.Method", request.Method))

	// Handle Request According To Type
	switch request.Method {
	case "POST":

		// Get The Partition Key From The Request Headers
		partitionKey := request.Header.Get(PartitionKeyHttpHeader)

		// Read The HTTP Request Body Into A Byte Array
		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			logger.Error("Failed To Read Request Body", zap.Error(err))
			_, err = fmt.Fprint(responseWriter, "Failed To Read Request Body!\n")
			if err != nil {
				logger.Error("Failed To Write Error Message To ResponseWriter", zap.Error(err))
			}
		} else {
			logger.Debug("Read Request Body", zap.String("PartitionKey", partitionKey), zap.ByteString("Body", bodyBytes))
		}

		// Send The PartitionKey & Message To Kafka Topic
		sendMessageError := c.SendMessage(partitionKey, bodyBytes)
		if sendMessageError != nil {
			responseWriter.WriteHeader(http.StatusInternalServerError)
			logger.Error("Message failed to persist", zap.Error(sendMessageError))
		} else {
			// Write The Response To The HTTP ResponseWriter Stream (Kyma requires a 202 response or it will return 500 error!)
			responseWriter.WriteHeader(http.StatusAccepted)
			_, err = responseWriter.Write([]byte("Message Processed!"))
			if err != nil {
				logger.Error("Failed To Write Processed Message To ResponseWriter", zap.Error(err))
			}
		}

	default:
		logger.Error("Only POST Requests Are Supported!")
		_, err := fmt.Fprint(responseWriter, "Only POST Requests Are Supported!")
		if err != nil {
			logger.Error("Failed To Write Post Requests Message To ResponseWriter", zap.Error(err))
		}
	}
}

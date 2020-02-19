package prometheus

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

// The MetricsServer Struct
type MetricsServer struct {
	logger                 *zap.Logger
	server                 *http.Server
	httpPort               string
	path                   string
	producedMsgCountGauges *prometheus.GaugeVec
	receivedMsgCountGauges *prometheus.GaugeVec
}

// MetricsServer Constructor
func NewMetricsServer(httpPort string, path string) *MetricsServer {

	// Create The MetricsServer Instance
	metricsServer := &MetricsServer{
		httpPort: httpPort,
		path:     path,
		producedMsgCountGauges: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "knative_kafka_produced_msg_count",
		}, []string{"producer", "topic", "partition"}),
		receivedMsgCountGauges: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "knative_kafka_consumed_msg_count",
		}, []string{"consumer", "topic", "partition"}),
	}

	prometheus.Register(metricsServer.producedMsgCountGauges)
	prometheus.Register(metricsServer.receivedMsgCountGauges)

	// Initialize The Logger
	metricsServer.initializeLogger()

	// Initialize The HTTP Server
	metricsServer.initializeServer()

	// Return The MetricsServer
	return metricsServer
}

// Start The Metrics HTTP Server Listening For Requests
func (m *MetricsServer) Start() {
	m.logger.Info("Starting Prometheus Metrics HTTP Server")
	go func() {
		err := m.server.ListenAndServe()
		if err != nil {
			m.logger.Info("Prometheus Metrics HTTP ListenAndServe Returned Error", zap.Error(err)) // Info log since it could just be normal shutdown
		}
	}()
}

// Stop The Metrics HTTP Server Listening For Requests
func (m *MetricsServer) Stop() {
	m.logger.Info("Stopping Prometheus Metrics HTTP Server")
	err := m.server.Shutdown(context.TODO())
	if err != nil {
		m.logger.Error("Failed To Shutdown Prometheus Metrics HTTP Server", zap.Error(err))
	}
}

// Initialize The Prometheus Metrics HTTP Server
func (m *MetricsServer) initializeServer() {

	// Create The ServeMux
	serveMux := http.NewServeMux()

	// Configure The Metrics Path To Use The Default Prometheus Handler
	serveMux.Handle(m.path, promhttp.Handler())

	// Create The Server For Configured HTTP Port
	server := &http.Server{Addr: ":" + m.httpPort, Handler: serveMux}

	// Set The Initialized HTTP Server
	m.server = server
}

// Initialize The MetricsServer's Logger
func (m *MetricsServer) initializeLogger() {

	// Create A New Zap Production Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Logger().Fatal("Failed To Create Zap Production Logger: %v", zap.Error(err))
	}
	if logger == nil {
		log.Logger().Fatal("Created Nil Zap Production Logger")
	}

	// Set The Initialized Logger
	m.logger = logger
}

// Observe librdkafka Metrics
func (m *MetricsServer) Observe(stats string) {

	log.Logger().Debug("New Producer Metrics Observed")
	var statsMap map[string]interface{}
	err := json.Unmarshal([]byte(stats), &statsMap)

	// If Unable To Parse Log And Move On
	if err != nil {
		log.Logger().Error("Unable To Parse Kafka Metrics", zap.Error(err))
		return
	}

	// Name Of Producer Or Consumer
	name := statsMap["name"].(string)

	// Loop Over Topics And Partitions Updating Metrics
	for _, t := range statsMap["topics"].(map[string]interface{}) {
		topic := t.(map[string]interface{})
		topicName := topic["topic"].(string)

		for _, partition := range topic["partitions"].(map[string]interface{}) {
			updatePartitionGauge(name, topicName, partition.(map[string]interface{}), "txmsgs", m.producedMsgCountGauges)
			updatePartitionGauge(name, topicName, partition.(map[string]interface{}), "rxmsgs", m.receivedMsgCountGauges)
		}
	}
}

// Update A Partition Level Gauge
func updatePartitionGauge(name string, topicName string, partition map[string]interface{}, field string, gauge *prometheus.GaugeVec) {
	partitionNum := partition["partition"].(float64)
	partitionNumStr := strconv.FormatFloat(partitionNum, 'f', -1, 64)

	if partitionNumStr != "-1" {
		msgCount := partition[field].(float64)
		gauge.WithLabelValues(name, topicName, partitionNumStr).Set(msgCount)
	}
}

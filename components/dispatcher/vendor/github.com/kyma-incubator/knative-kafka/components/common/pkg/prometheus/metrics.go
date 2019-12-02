package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"log"
	"net/http"
)

// The MetricsServer Struct
type metricsServer struct {
	logger   *zap.Logger
	server   *http.Server
	httpPort string
	path     string
}

// MetricsServer Constructor
func NewMetricsServer(httpPort string, path string) *metricsServer {

	// Create The MetricsServer Instance
	metricsServer := &metricsServer{
		httpPort: httpPort,
		path:     path,
	}

	// Initialize The Logger
	metricsServer.initializeLogger()

	// Initialize The HTTP Server
	metricsServer.initializeServer()

	// Return The MetricsServer
	return metricsServer
}

// Start The Metrics HTTP Server Listening For Requests
func (m *metricsServer) Start() {
	m.logger.Info("Starting Prometheus Metrics HTTP Server")
	go func() {
		err := m.server.ListenAndServe()
		if err != nil {
			m.logger.Info("Prometheus Metrics HTTP ListenAndServe Returned Error", zap.Error(err)) // Info log since it could just be normal shutdown
		}
	}()
}

// Stop The Metrics HTTP Server Listening For Requests
func (m *metricsServer) Stop() {
	m.logger.Info("Stopping Prometheus Metrics HTTP Server")
	err := m.server.Shutdown(context.TODO())
	if err != nil {
		m.logger.Error("Failed To Shutdown Prometheus Metrics HTTP Server", zap.Error(err))
	}
}

// Initialize The Prometheus Metrics HTTP Server
func (m *metricsServer) initializeServer() {

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
func (m *metricsServer) initializeLogger() {

	// Create A New Zap Production Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed To Create Zap Production Logger: %v", err)
	}
	if logger == nil {
		log.Fatalf("Created Nil Zap Production Logger")
	}

	// Set The Initialized Logger
	m.logger = logger
}

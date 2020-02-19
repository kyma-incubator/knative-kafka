package health

import (
	"context"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"go.uber.org/zap"
	"net/http"
)

// Start The HTTP Server Listening For Requests

type Health struct {
	server      *http.Server // The Golang HTTP Server Instance
	httpPort    string       // The HTTP Port The Health Server Listens On
}

// Creates A New EventProxy With Specified Configuration
func NewHealthServer(httpPort string) *Health {
	health := &Health{
		httpPort:    httpPort,
	}

	// Initialize The HTTP Server
	health.initializeServer(httpPort)

	// Return The Health Server
	return health
}

// Initialize The HTTP Server
func (ep *Health) initializeServer(httpPort string) {

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/healthz", ep.handleLiveness)
	serveMux.HandleFunc("/healthy", ep.handleReadiness)

	// Create The Server For Configured HTTP Port
	server := &http.Server{Addr: ":" + httpPort, Handler: serveMux}

	// Set The Initialized HTTP Server
	ep.server = server
}

func (ep *Health) Start() {
	log.Logger().Info("Starting Health HTTP Server on port " + ep.httpPort)
	go func() {
		err := ep.server.ListenAndServe()
		if err != nil {
			log.Logger().Info("Health HTTP ListenAndServe Returned Error", zap.Error(err)) // Info log since it could just be normal shutdown
		}
	}()
}

// Stop The HTTP Server Listening For Requests
func (ep *Health) Stop() {
	log.Logger().Info("Stopping Health HTTP Server")
	err := ep.server.Shutdown(context.TODO())
	if err != nil {
		log.Logger().Error("Health Failed To Shutdown HTTP Server", zap.Error(err))
	}
}

// HTTP Request Handler For Liveness Requests (/healthz)
func (ep *Health) handleLiveness(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		responseWriter.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	responseWriter.WriteHeader(http.StatusOK)
}

// HTTP Request Handler For Readiness Requests (/healthy)
func (ep *Health) handleReadiness(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		responseWriter.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	responseWriter.WriteHeader(http.StatusOK)
}

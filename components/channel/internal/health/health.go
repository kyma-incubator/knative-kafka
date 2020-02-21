package health

import (
	"context"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"go.uber.org/zap"
	"net/http"
)

// Start The HTTP Server Listening For Requests

type HealthServer struct {
	server         *http.Server // The Golang HTTP Server Instance
	httpPort       string       // The HTTP Port The HealthServer Server Listens On
	Alive          bool         // A flag that controls the response to liveness requests
	ProducerReady  bool         // A flag that the producer sets when it is ready
	ChannelReady   bool         // A flag that the channel sets when it is ready
}

// Creates A New EventProxy With Specified Configuration
func NewHealthServer(httpPort string) *HealthServer {
	health := &HealthServer{
		httpPort:        httpPort,
		Alive:           false,
		ProducerReady:   false,
		ChannelReady:    false,
	}

	// Initialize The HTTP Server
	health.initializeServer(httpPort)

	// Return The HealthServer Server
	return health
}

// Set All Liveness And Readiness Flags To False
func (ep *HealthServer) ShuttingDown() {
	ep.Alive = false
	ep.ProducerReady = false
	ep.ChannelReady = false
}

// Initialize The HTTP Server
func (ep *HealthServer) initializeServer(httpPort string) {

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/healthz", ep.handleLiveness)
	serveMux.HandleFunc("/healthy", ep.handleReadiness)

	// Create The Server For Configured HTTP Port
	server := &http.Server{Addr: ":" + httpPort, Handler: serveMux}

	// Set The Initialized HTTP Server
	ep.server = server
}

func (ep *HealthServer) Start() {
	log.Logger().Info("Starting HealthServer HTTP Server on port " + ep.httpPort)
	go func() {
		err := ep.server.ListenAndServe()
		if err != nil {
			log.Logger().Info("HealthServer HTTP ListenAndServe Returned Error", zap.Error(err)) // Info log since it could just be normal shutdown
		}
	}()
}

// Stop The HTTP Server Listening For Requests
func (ep *HealthServer) Stop() {
	log.Logger().Info("Stopping HealthServer HTTP Server")
	err := ep.server.Shutdown(context.TODO())
	if err != nil {
		log.Logger().Error("HealthServer Failed To Shutdown HTTP Server", zap.Error(err))
	}
}

// HTTP Request Handler For Liveness Requests (/healthz)
func (ep *HealthServer) handleLiveness(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		responseWriter.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if ep.Alive {
		responseWriter.WriteHeader(http.StatusOK)
	} else {
		responseWriter.WriteHeader(http.StatusInternalServerError)
	}
}

// HTTP Request Handler For Readiness Requests (/healthy)
func (ep *HealthServer) handleReadiness(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		responseWriter.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if ep.ProducerReady && ep.ChannelReady {
		responseWriter.WriteHeader(http.StatusOK)
	} else {
		responseWriter.WriteHeader(http.StatusInternalServerError)
	}
}

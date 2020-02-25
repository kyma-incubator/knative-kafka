package dispatcherhealth

import (
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/health"
	"sync"
)

// Start The HTTP Server Listening For Requests

type DispatcherHealthServer struct {
	health.Server

	// Additional Synchronization Mutexes
	dispatcherMutex sync.Mutex // Synchronizes access to the dispatcherReady flag

	// Additional Internal Flags
	dispatcherReady bool // A flag that the producer sets when it is ready
}

// Creates A New DispatcherHealthServer With Specified Configuration
func NewDispatcherHealthServer(httpPort string) *DispatcherHealthServer {
	channelHealth := &DispatcherHealthServer{}
	health := health.NewHealthServer(httpPort, channelHealth)
	channelHealth.Server = *health

	// Return The DispatcherHealthServer
	return channelHealth
}

// Synchronized Function To Set Producer Ready Flag
func (chs *DispatcherHealthServer) SetDispatcherReady(isReady bool) {
	chs.dispatcherMutex.Lock()
	chs.dispatcherReady = isReady
	chs.dispatcherMutex.Unlock()
}

// Set All Liveness And Readiness Flags To False
func (chs *DispatcherHealthServer) ShuttingDown() {
	chs.Server.ShuttingDown()
	chs.SetDispatcherReady(false)
}

// Access Function For DispatcherReady Flag
func (chs *DispatcherHealthServer) IsDispatcherReady() bool {
	return chs.dispatcherReady
}

// Functions That Implement The HealthInterface

// HTTP Request Handler For Readiness Requests (/healthy)
func (chs *DispatcherHealthServer) IsReady() bool {
	return chs.dispatcherReady
}

// HTTP Request Handler For Liveness Requests (/healthz)
func (chs *DispatcherHealthServer) IsAlive() bool {
	return chs.Server.IsAlive()
}

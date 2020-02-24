package health

import (
	"sync"
)

// Start The HTTP Server Listening For Requests

type ChannelHealthServer struct {
	Server

	// Additional Synchronization Mutexes
	producerMutex  sync.Mutex	// Synchronizes access to the producerReady flag
	channelMutex   sync.Mutex	// Synchronizes access to the channelReady flag

	// Additional Internal Flags
	producerReady  bool         // A flag that the producer sets when it is ready
	channelReady   bool         // A flag that the channel sets when it is ready
}

// Creates A New ChannelHealthServer With Specified Configuration
func NewChannelHealthServer(httpPort string) *ChannelHealthServer {
	channelHealth := &ChannelHealthServer{}
	health := NewHealthServer(httpPort, channelHealth)
	channelHealth.Server = *health

	// Return The ChannelHealthServer
	return channelHealth
}

// Synchronized Function To Set Producer Ready Flag
func (chs *ChannelHealthServer) SetProducerReady(isReady bool) {
	chs.producerMutex.Lock()
	chs.producerReady = isReady
	chs.producerMutex.Unlock()
}

// Synchronized Function To Set Channel Ready Flag
func (chs *ChannelHealthServer) SetChannelReady(isReady bool) {
	chs.channelMutex.Lock()
	chs.channelReady = isReady
	chs.channelMutex.Unlock()
}

// Set All Liveness And Readiness Flags To False
func (chs *ChannelHealthServer) ShuttingDown() {
	chs.Server.ShuttingDown()
	chs.SetProducerReady(false)
	chs.SetChannelReady(false)
}

// Access Function For ProducerReady Flag
func (chs *ChannelHealthServer) IsProducerReady() bool {
	return chs.producerReady
}

// Access Function For ChannelReady Flag
func (chs *ChannelHealthServer) IsChannelReady() bool {
	return chs.channelReady
}

// Functions That Implement The HealthInterface

// HTTP Request Handler For Readiness Requests (/healthy)
func (chs *ChannelHealthServer) IsReady() bool {
	return chs.producerReady && chs.channelReady
}

// HTTP Request Handler For Readiness Requests (/healthy)
func (chs *ChannelHealthServer) IsAlive() bool {
	return chs.Server.IsAlive()
}

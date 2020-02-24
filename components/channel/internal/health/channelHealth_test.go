package health

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

// Set Up Test Logger
// var logger = log.TestLogger()

// Test The NewChannelHealthServer() Functionality
func TestNewChannelHealthServer(t *testing.T) {

	// Create A Health Server
	health := NewChannelHealthServer(testHttpPort)

	// Validate The EventProxy
	assert.NotNil(t, health)
	assert.Equal(t, false, health.IsAlive())
	assert.Equal(t, false, health.channelReady)
	assert.Equal(t, false, health.producerReady)
}

// Test Flag Set And Reset Functions
func TestReadinessFlagWrites(t *testing.T) {

	// Create A New Health Server
	chs := NewChannelHealthServer(testHttpPort)

	// Test Readiness Flags
	chs.SetProducerReady(false)
	assert.Equal(t, false, chs.IsProducerReady())
	chs.SetProducerReady(true)
	assert.Equal(t, true, chs.IsProducerReady())
	chs.SetChannelReady(false)
	assert.Equal(t, false, chs.IsChannelReady())
	chs.SetChannelReady(true)
	assert.Equal(t, true, chs.IsChannelReady())

}

// Test The Health Server Via The HTTP Handlers
func TestChannelHealthHandler(t *testing.T) {

	// Create A New Channel Health Server
	chs := NewChannelHealthServer(testHttpPort)

	// Verify that initially the readiness status is false
	getEventToHandler(t, chs.handleReadiness, readinessPath, http.StatusInternalServerError)

	// Verify that the readiness status required setting all of the readiness flags
	chs.SetChannelReady(true)
	getEventToHandler(t, chs.handleReadiness, readinessPath, http.StatusInternalServerError)
	chs.SetProducerReady(true)
	getEventToHandler(t, chs.handleReadiness, readinessPath, http.StatusOK)
	chs.SetChannelReady(false)
	getEventToHandler(t, chs.handleReadiness, readinessPath, http.StatusInternalServerError)

	// Verify that the shutdown process sets all statuses to not live / not ready
	chs.SetProducerReady(true)
	chs.SetChannelReady(true)
	getEventToHandler(t, chs.handleReadiness, readinessPath, http.StatusOK)

	chs.ShuttingDown()
	getEventToHandler(t, chs.handleReadiness, readinessPath, http.StatusInternalServerError)
}

// Test The Health Server Via Live HTTP Calls
func TestChannelHealthServer(t *testing.T) {
	chs := NewChannelHealthServer(testHttpPort)
	chs.Start()

	readinessUri, err := url.Parse(fmt.Sprintf("http://%s:%s%s", testHttpHost , testHttpPort, readinessPath))
	assert.Nil(t, err)

	// Test basic functionality - advanced logical tests are in TestChannelHealthHandler
	getEventToServer(t, readinessUri, http.StatusInternalServerError)
	chs.SetChannelReady(true)
	chs.SetProducerReady(true)
	getEventToServer(t, readinessUri, http.StatusOK)

	chs.Stop()
}

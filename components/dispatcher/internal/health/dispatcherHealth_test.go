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

// Test The NewDispatcherHealthServer() Functionality
func TestNewDispatcherHealthServer(t *testing.T) {

	// Create A Health Server
	health := NewDispatcherHealthServer(testHttpPort)

	// Validate The EventProxy
	assert.NotNil(t, health)
	assert.Equal(t, false, health.IsAlive())
	assert.Equal(t, false, health.dispatcherReady)
}

// Test Flag Set And Reset Functions
func TestReadinessFlagWrites(t *testing.T) {

	// Create A New Health Server
	chs := NewDispatcherHealthServer(testHttpPort)

	// Test Readiness Flags
	chs.SetDispatcherReady(false)
	assert.Equal(t, false, chs.IsDispatcherReady())
	chs.SetDispatcherReady(true)
	assert.Equal(t, true, chs.IsDispatcherReady())
}

// Test The Health Server Via The HTTP Handlers
func TestDispatcherHealthHandler(t *testing.T) {

	// Create A New Dispatcher Server
	chs := NewDispatcherHealthServer(testHttpPort)

	// Verify that initially the readiness status is false
	getEventToHandler(t, chs.handleReadiness, readinessPath, http.StatusInternalServerError)

	// Verify that the readiness status required setting all of the readiness flags
	chs.SetDispatcherReady(true)
	getEventToHandler(t, chs.handleReadiness, readinessPath, http.StatusOK)

	// Verify that the shutdown process sets all statuses to not live / not ready
	chs.SetDispatcherReady(true)
	getEventToHandler(t, chs.handleReadiness, readinessPath, http.StatusOK)

	chs.ShuttingDown()
	getEventToHandler(t, chs.handleReadiness, readinessPath, http.StatusInternalServerError)
}

// Test The Health Server Via Live HTTP Calls
func TestDispatcherHealthServer(t *testing.T) {
	chs := NewDispatcherHealthServer(testHttpPort)
	chs.Start()

	readinessUri, err := url.Parse(fmt.Sprintf("http://%s:%s%s", testHttpHost , testHttpPort, readinessPath))
	assert.Nil(t, err)

	// Test basic functionality - advanced logical tests are in TestDispatcherHealthHandler
	getEventToServer(t, readinessUri, http.StatusInternalServerError)
	chs.SetDispatcherReady(true)
	getEventToServer(t, readinessUri, http.StatusOK)

	chs.Stop()
}

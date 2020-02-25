package dispatcherhealth

import (
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

const (
	testHttpPort    = "8089"
	testHttpHost    = "localhost"
	readinessPath   = "/healthy"
)

// Set Up Test Logger
var logger = log.TestLogger()

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

// Test The Dispatcher Health Server Via Live HTTP Calls
func TestDispatcherHealthServer(t *testing.T) {
	chs := NewDispatcherHealthServer(testHttpPort)
	chs.Start(logger)

	readinessUri, err := url.Parse(fmt.Sprintf("http://%s:%s%s", testHttpHost , testHttpPort, readinessPath))
	assert.Nil(t, err)

	// Verify that initially the readiness status is false
	getEventToServer(t, readinessUri, http.StatusInternalServerError)

	// Verify that the readiness status required setting all of the readiness flags
	chs.SetDispatcherReady(true)
	getEventToServer(t, readinessUri, http.StatusOK)

	// Verify that the shutdown process sets all statuses to not live / not ready
	chs.SetDispatcherReady(true)
	getEventToServer(t, readinessUri, http.StatusOK)

	chs.ShuttingDown()
	getEventToServer(t, readinessUri, http.StatusInternalServerError)

	chs.Stop(logger)
}

//
// Private Utility Functions
//

// Sends A Simple GET Event To A URL Expecting A Specific Response Code
func getEventToServer(t *testing.T, uri *url.URL, expectedStatus int) {

	// Create An HTTP Client And Send The Request
	client := http.DefaultClient
	resp, err := client.Get(uri.String())

	// Verify The Client Response Is As Expected
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.Equal(t, expectedStatus, resp.StatusCode)
}

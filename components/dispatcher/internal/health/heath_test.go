package health

import (
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Test Constants
const (
	testHttpPort    = "8089"
	testHttpHost    = "localhost"
	livenessPath    = "/healthz"
	readinessPath   = "/healthy"
)

// Set Up Test Logger
var logger = log.TestLogger()

// Test The NewEventProxy() Functionality
func TestNewEventProxy(t *testing.T) {

	// Create An EventProxy
	health := NewHealthServer(testHttpPort)

	// Validate The EventProxy
	assert.NotNil(t, health)
	assert.NotNil(t, health.server)
	assert.Equal(t, ":"+testHttpPort, health.server.Addr)
	assert.Equal(t, false, health.Alive)
	assert.Equal(t, false, health.DispatcherReady)
}

// Test Unsupported Events Requests
func TestUnsupportedEventsRequests(t *testing.T) {

	// Create A New EventProxy & Get The Events Handler Function
	health := NewHealthServer(testHttpPort)

	// Test All Unsupported Events Requests
	performUnsupportedMethodRequestTest(t, http.MethodConnect, livenessPath, health.handleLiveness)
	performUnsupportedMethodRequestTest(t, http.MethodDelete, livenessPath, health.handleLiveness)
	performUnsupportedMethodRequestTest(t, http.MethodPost, livenessPath, health.handleLiveness)
	performUnsupportedMethodRequestTest(t, http.MethodOptions, livenessPath, health.handleLiveness)
	performUnsupportedMethodRequestTest(t, http.MethodPatch, livenessPath, health.handleLiveness)
	performUnsupportedMethodRequestTest(t, http.MethodPut, livenessPath, health.handleLiveness)
	performUnsupportedMethodRequestTest(t, http.MethodTrace, livenessPath, health.handleLiveness)

	performUnsupportedMethodRequestTest(t, http.MethodConnect, readinessPath, health.handleReadiness)
	performUnsupportedMethodRequestTest(t, http.MethodDelete, readinessPath, health.handleReadiness)
	performUnsupportedMethodRequestTest(t, http.MethodPost, readinessPath, health.handleReadiness)
	performUnsupportedMethodRequestTest(t, http.MethodOptions, readinessPath, health.handleReadiness)
	performUnsupportedMethodRequestTest(t, http.MethodPatch, readinessPath, health.handleReadiness)
	performUnsupportedMethodRequestTest(t, http.MethodPut, readinessPath, health.handleReadiness)
	performUnsupportedMethodRequestTest(t, http.MethodTrace, readinessPath, health.handleReadiness)
}

// Test The Health Server Via The HTTP Handlers
func TestHealthHandler(t *testing.T) {

	// Create A New HealthServer Server
	health := NewHealthServer(testHttpPort)

	// Verify that initially the statuses are not live / not ready
	getEventToHandler(t, health.handleLiveness, livenessPath, http.StatusInternalServerError)
	getEventToHandler(t, health.handleReadiness, readinessPath, http.StatusInternalServerError)

	// Verify that the liveness status follows the health.Alive flag
	health.Alive = true
	getEventToHandler(t, health.handleLiveness, livenessPath, http.StatusOK)

	// Verify that the readiness status required setting all of the readiness flags
	health.DispatcherReady = true
	getEventToHandler(t, health.handleReadiness, readinessPath, http.StatusOK)

	// Verify that the shutdown process sets all statuses to not live / not ready
	health.DispatcherReady = true
	getEventToHandler(t, health.handleReadiness, readinessPath, http.StatusOK)

	health.ShuttingDown()
	getEventToHandler(t, health.handleReadiness, readinessPath, http.StatusInternalServerError)
	getEventToHandler(t, health.handleLiveness, livenessPath, http.StatusInternalServerError)

}

// Test The Health Server Via Live HTTP Calls
func TestHealthServer(t *testing.T) {
	health := NewHealthServer(testHttpPort)
	health.Start()

	livenessUri, err := url.Parse(fmt.Sprintf("http://%s:%s%s", testHttpHost , testHttpPort, livenessPath))
	assert.Nil(t, err)
	readinessUri, err := url.Parse(fmt.Sprintf("http://%s:%s%s", testHttpHost , testHttpPort, readinessPath))
	assert.Nil(t, err)

	// Test basic functionality - advanced logical tests are in TestHealthGets
	getEventToServer(t, livenessUri, http.StatusInternalServerError)
	getEventToServer(t, readinessUri, http.StatusInternalServerError)
	health.Alive = true
	getEventToServer(t, livenessUri, http.StatusOK)
	health.DispatcherReady = true
	getEventToServer(t, readinessUri, http.StatusOK)

	health.Stop()
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

// Sends A Request To An HTTP Response Recorder Directly Expecting A Specific Response Code
func getEventToHandler(t *testing.T, handlerFunc http.HandlerFunc, path string, expectedStatus int) {
	// Create A Test HTTP GET Request For requested path
	request := createNewRequest(t, "GET", path, nil)

	// Create An HTTP ResponseRecorder & Handler For Request
	responseRecorder := httptest.NewRecorder()
	handler := handlerFunc

	// Call The HTTP Request Handler Function For Path
	handler.ServeHTTP(responseRecorder, request)

	// Verify The StatusMethodNotAllowed Response Code Is Returned
	statusCode := responseRecorder.Code
	assert.Equal(t, expectedStatus, statusCode)

}

// Create A Test HTTP Request For The Specified Method / Path
func createNewRequest(t *testing.T, method string, path string, body io.Reader) *http.Request {
	request, err := http.NewRequest(method, path, body)
	assert.Nil(t, err)
	return request
}

// Perform An Unsupported Method Request Test To Specified Path/Handler
func performUnsupportedMethodRequestTest(t *testing.T, method string, path string, handlerFunc http.HandlerFunc) {

	// Create A Test HTTP POST Request For Metrics Path
	request := createNewRequest(t, method, path, nil)

	// Create An HTTP ResponseRecorder & Handler For Request
	responseRecorder := httptest.NewRecorder()
	handler := handlerFunc

	// Call The HTTP Request Handler Function For Path
	handler.ServeHTTP(responseRecorder, request)

	// Verify The StatusMethodNotAllowed Response Code Is Returned
	statusCode := responseRecorder.Code
	assert.Equal(t, http.StatusMethodNotAllowed, statusCode)
}

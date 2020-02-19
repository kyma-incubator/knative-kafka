package health

import (
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test Constants
const (
	testHost        = "localhost"
	testHttpPort    = "8088"
	testUrl         = testHost + ":" + testHttpPort
	livenessPath    = "/healthz"
	readinessPath   = "/healthy"
	testContentType = "application/json"
	testUsername    = "TestUsername"
	testPassword    = "TestPassword"
	testMessage     = "TestMessage"
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
func TestHealthGets(t *testing.T) {

	// Create A New Health Server
	health := NewHealthServer(testHttpPort)

	getEventToHandler(t, health.handleLiveness, livenessPath, http.StatusOK)
	getEventToHandler(t, health.handleReadiness, readinessPath, http.StatusOK)
}

func getEventToHandler(t *testing.T, handlerFunc http.HandlerFunc, path string, expectedStatus int) {
	// Create A Test HTTP GET Request For requested path
	request := createNewRequest(t, "GET", path, nil)

	// Create An HTTP ResponseRecorder & Handler For Request
	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(handlerFunc)

	// Call The HTTP Request Handler Function For Path
	handler.ServeHTTP(responseRecorder, request)

	// Verify The StatusMethodNotAllowed Response Code Is Returned
	statusCode := responseRecorder.Code
	assert.Equal(t, expectedStatus, statusCode)

}

//
// Private Utility Functions
//

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
	handler := http.HandlerFunc(handlerFunc)

	// Call The HTTP Request Handler Function For Path
	handler.ServeHTTP(responseRecorder, request)

	// Verify The StatusMethodNotAllowed Response Code Is Returned
	statusCode := responseRecorder.Code
	assert.Equal(t, http.StatusMethodNotAllowed, statusCode)
}

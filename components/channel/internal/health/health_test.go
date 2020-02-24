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

// Test struct that implements the HealthInterface functions
type testStatus struct {
	server *Server
}

func (ts *testStatus) IsAlive() bool {
	return ts.server.IsAlive()
}

func (ts *testStatus) IsReady() bool {
	return true
}

var mockStatus testStatus

func getTestHealthServer() *Server {
	health := NewHealthServer(testHttpPort, &mockStatus)
	mockStatus.server = health
	return health
}

// Test The NewHealthServer() Functionality
func TestNewHealthServer(t *testing.T) {

	// Create A Health Server
	health := getTestHealthServer()

	// Validate The EventProxy
	assert.NotNil(t, health)
	assert.NotNil(t, health.server)
	assert.Equal(t, ":"+testHttpPort, health.server.Addr)
	assert.Equal(t, false, health.alive)
}

// Test Flag Set And Reset Functions
func TestFlagWrites(t *testing.T) {

	// Create A New Health Server
	health := getTestHealthServer()

	// Test Liveness Flag
	health.SetAlive(false)
	assert.Equal(t, false, health.IsAlive())
	health.SetAlive(true)
	assert.Equal(t, true, health.IsAlive())

	// Verify that Readiness is true by default
	assert.Equal(t, true, health.status.IsReady())
}

// Test Unsupported Events Requests
func TestUnsupportedEventsRequests(t *testing.T) {

	// Create A New Health Server
	health := getTestHealthServer()

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

	// Create A New Health Server
	health := getTestHealthServer()

	// Verify that initially the statuses are not live / ready
	getEventToHandler(t, health.handleLiveness, livenessPath, http.StatusInternalServerError)
	getEventToHandler(t, health.handleReadiness, readinessPath, http.StatusOK)

	// Verify that the liveness status follows the health.Alive flag
	health.SetAlive(true)
	getEventToHandler(t, health.handleLiveness, livenessPath, http.StatusOK)

	// Verify that the shutdown process sets liveness to false
	health.ShuttingDown()
	getEventToHandler(t, health.handleLiveness, livenessPath, http.StatusInternalServerError)
}

// Test The Health Server Via Live HTTP Calls
func TestHealthServer(t *testing.T) {
	health := getTestHealthServer()
	health.Start()

	livenessUri, err := url.Parse(fmt.Sprintf("http://%s:%s%s", testHttpHost , testHttpPort, livenessPath))
	assert.Nil(t, err)
	readinessUri, err := url.Parse(fmt.Sprintf("http://%s:%s%s", testHttpHost , testHttpPort, readinessPath))
	assert.Nil(t, err)

	// Test basic functionality - advanced logical tests are in TestHealthHandler
	getEventToServer(t, livenessUri, http.StatusInternalServerError)
	getEventToServer(t, readinessUri, http.StatusOK)
	health.SetAlive(true)
	getEventToServer(t, livenessUri, http.StatusOK)

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

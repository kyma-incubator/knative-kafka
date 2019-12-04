package util

import (
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// Test Constants
const (
	subscriptionName = "test-subscription-name"
)

// Test The newControllerRef() Functionality
func TestNewControllerRef(t *testing.T) {

	// Test Data
	subscription := &eventingv1alpha1.Subscription{ObjectMeta: metav1.ObjectMeta{Name: subscriptionName}}

	// Perform The Test
	actualControllerRef := NewSubscriptionControllerRef(subscription)

	// Verify The Results
	assert.NotNil(t, actualControllerRef)
	assert.Equal(t, eventingv1alpha1.SchemeGroupVersion.Group+"/"+eventingv1alpha1.SchemeGroupVersion.Version, actualControllerRef.APIVersion)
	assert.Equal(t, "Subscription", actualControllerRef.Kind)
	assert.Equal(t, subscriptionName, actualControllerRef.Name)
}

// Test The dispatcherServiceName() Functionality
func TestDispatcherServiceName(t *testing.T) {

	// Test Data
	subscription := &eventingv1alpha1.Subscription{ObjectMeta: metav1.ObjectMeta{Name: subscriptionName}}

	// Perform The Test
	expectedDispatcherServiceName := subscriptionName + "-dispatcher"

	// Verify The Results
	actualDispatcherServiceName := DispatcherServiceName(subscription)
	assert.Equal(t, expectedDispatcherServiceName, actualDispatcherServiceName)
}

// Test The dispatcherDeploymentName() Functionality
func TestDispatcherDeploymentName(t *testing.T) {

	// Test Data
	subscription := &eventingv1alpha1.Subscription{ObjectMeta: metav1.ObjectMeta{Name: subscriptionName}}

	// Perform The Test
	expectedDispatcherDeploymentName := subscriptionName + "-dispatcher"

	// Verify The Results
	actualDispatcherDeploymentName := DispatcherDeploymentName(subscription)
	assert.Equal(t, expectedDispatcherDeploymentName, actualDispatcherDeploymentName)
}

// Test The generateValidDispatcherDnsName() Functionality
func TestGenerateValidDispatcherDnsName(t *testing.T) {

	// Define The TestCase Struct
	type TestCase struct {
		testValue      string;
		expectedResult string
	}

	// Define The Test Cases
	testCases := []TestCase{
		{"", "kk--dispatcher"},
		{"1234567890", "kk-1234567890-dispatcher"},
		{"lambda-foo-product.created-v1--stage", "lambda-foo-productcreated-v1--stage-dispatcher"},
		{"LAMBDA-FOO-PRODUCT.CREATED-V1--STAGE", "lambda-foo-productcreated-v1--stage-dispatcher"},
		{"somelongstringwith54charactershmmmmmmmmmmmmmmmmmmmmmm7", "somelongstringwith54charactershmmmmmmmmmmmmmmmmmmmmmm"},
	}

	// Test All The TestCases
	for _, testCase := range testCases {
		actualResult := generateValidDispatcherDnsName(testCase.testValue)
		assert.Equal(t, testCase.expectedResult, actualResult)
	}
}

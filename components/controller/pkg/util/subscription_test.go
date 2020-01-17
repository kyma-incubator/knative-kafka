package util

import (
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"testing"
)

// Test The SubscriptionLogger Functionality
func TestSubscriptionLogger(t *testing.T) {

	// Test Logger
	logger := log.TestLogger()

	// Test Data
	subscription := &messagingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{Name: "TestSubscriptionName", Namespace: "TestSubscriptionNamespace"},
	}

	// Perform The Test
	subscriptionLogger := SubscriptionLogger(logger, subscription)
	assert.NotNil(t, subscriptionLogger)
	assert.NotEqual(t, logger, subscriptionLogger)
	subscriptionLogger.Info("Testing Subscription Logger")
}

// Test The NewSubscriptionControllerRef Functionality
func TestNewSubscriptionControllerRef(t *testing.T) {

	// Test Data
	subscription := &messagingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{Name: "TestName"},
	}

	// Perform The Test
	controllerRef := NewSubscriptionControllerRef(subscription)

	// Validate Results
	assert.NotNil(t, controllerRef)
	assert.Equal(t, messagingv1alpha1.SchemeGroupVersion.String(), controllerRef.APIVersion)
	assert.Equal(t, constants.KnativeSubscriptionKind, controllerRef.Kind)
	assert.Equal(t, subscription.ObjectMeta.Name, controllerRef.Name)
	assert.True(t, *controllerRef.Controller)
}

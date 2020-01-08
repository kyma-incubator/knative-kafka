package util

import (
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"strconv"
	"testing"
)

// Test Constants
const (
	eventStartTime                         = "TestEventStartTime"
	eventRetryInitialIntervalMillis        = 111
	defaultEventRetryInitialIntervalMillis = int64(222)
	eventRetryTimeMillisMax                = 333
	defaultEventRetryTimeMillisMax         = int64(444)
	exponentialBackoff                     = false
	defaultExponentialBackoff              = true
	kafkaConsumers                         = 555
	defaultKafkaConsumers                  = 666
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

// Test The EventStartTime Accessor
func TestEventStartTime(t *testing.T) {

	// Test Logger
	logger := log.TestLogger()

	// Test The Default Failover Use Case
	subscription := &messagingv1alpha1.Subscription{}
	actualEventStartTime := EventStartTime(subscription, logger)
	assert.Equal(t, "", actualEventStartTime)

	// Test The Valid NumPartitions Use Case
	subscription = &messagingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{EventStartTimeAnnotation: eventStartTime},
		},
	}
	actualEventStartTime = EventStartTime(subscription, logger)
	assert.Equal(t, eventStartTime, actualEventStartTime)
}

// Test The EventRetryInitialIntervalMillis Accessor
func TestEventRetryInitialIntervalMillis(t *testing.T) {

	// Test Logger
	logger := log.TestLogger()

	// Test Data
	environment := &env.Environment{DefaultEventRetryInitialIntervalMillis: defaultEventRetryInitialIntervalMillis}

	// Test The Default Failover Use Case
	subscription := &messagingv1alpha1.Subscription{}
	actualEventRetryInitialIntervalMillis := EventRetryInitialIntervalMillis(subscription, environment, logger)
	assert.Equal(t, defaultEventRetryInitialIntervalMillis, actualEventRetryInitialIntervalMillis)

	// Test The Valid EventRetryIntervalMillis Use Case
	subscription = &messagingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{EventRetryInitialIntervalMillisAnnotation: strconv.Itoa(eventRetryInitialIntervalMillis)},
		},
	}
	actualEventRetryInitialIntervalMillis = EventRetryInitialIntervalMillis(subscription, environment, logger)
	assert.Equal(t, int64(eventRetryInitialIntervalMillis), actualEventRetryInitialIntervalMillis)
}

// Test The EventRetryTimeMillisMax Accessor
func TestEventRetryTimeMillisMax(t *testing.T) {

	// Test Logger
	logger := log.TestLogger()

	// Test Data
	environment := &env.Environment{DefaultEventRetryTimeMillisMax: defaultEventRetryTimeMillisMax}

	// Test The Default Failover Use Case
	subscription := &messagingv1alpha1.Subscription{}
	actualEventRetryTimeMillisMax := EventRetryTimeMillisMax(subscription, environment, logger)
	assert.Equal(t, defaultEventRetryTimeMillisMax, actualEventRetryTimeMillisMax)

	// Test The Valid EventRetryTimeMillisMax Use Case
	subscription = &messagingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{EventRetryTimeMillisMaxAnnotation: strconv.Itoa(eventRetryTimeMillisMax)},
		},
	}
	actualEventRetryTimeMillisMax = EventRetryTimeMillisMax(subscription, environment, logger)
	assert.Equal(t, int64(eventRetryTimeMillisMax), actualEventRetryTimeMillisMax)
}

// Test The ExponentialBackoff Accessor
func TestExponentialBackoff(t *testing.T) {

	// Test Logger
	logger := log.TestLogger()

	// Test Data
	environment := &env.Environment{DefaultExponentialBackoff: defaultExponentialBackoff}

	// Test The Default Failover Use Case
	subscription := &messagingv1alpha1.Subscription{}
	actualDefaultExponentialBackoff := ExponentialBackoff(subscription, environment, logger)
	assert.Equal(t, defaultExponentialBackoff, actualDefaultExponentialBackoff)

	// Test The Valid ExponentialBackoff Use Case
	subscription = &messagingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{ExponentialBackoffAnnotation: strconv.FormatBool(exponentialBackoff)},
		},
	}
	actualDefaultExponentialBackoff = ExponentialBackoff(subscription, environment, logger)
	assert.Equal(t, exponentialBackoff, actualDefaultExponentialBackoff)
}

// Test The KafkaConsumers Accessor
func TestKafkaConsumers(t *testing.T) {

	// Test Logger
	logger := log.TestLogger()

	// Test Data
	environment := &env.Environment{DefaultKafkaConsumers: defaultKafkaConsumers}

	// Test The Default Failover Use Case
	subscription := &messagingv1alpha1.Subscription{}
	actualKafkaConsumers := KafkaConsumers(subscription, environment, logger)
	assert.Equal(t, defaultKafkaConsumers, actualKafkaConsumers)

	// Test The Valid KafkaConsumers Use Case
	subscription = &messagingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{KafkaConsumersAnnotation: strconv.Itoa(kafkaConsumers)},
		},
	}
	actualKafkaConsumers = KafkaConsumers(subscription, environment, logger)
	assert.Equal(t, kafkaConsumers, actualKafkaConsumers)
}

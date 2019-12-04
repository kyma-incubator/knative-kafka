package util

import (
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strconv"
)

// Constants
const (
	// Subscription Annotations
	KnativeKafkaAnnotationPrefix              = "knativekafka.kyma-project.io"
	EventStartTimeAnnotation                  = KnativeKafkaAnnotationPrefix + "/EventStartTime"
	ExponentialBackoffAnnotation              = KnativeKafkaAnnotationPrefix + "/ExponentialBackoff"
	EventRetryInitialIntervalMillisAnnotation = KnativeKafkaAnnotationPrefix + "/EventRetryInitialIntervalMillis"
	EventRetryTimeMillisMaxAnnotation         = KnativeKafkaAnnotationPrefix + "/EventRetryTimeMillisMax"
	KafkaConsumersAnnotation                  = KnativeKafkaAnnotationPrefix + "/KafkaConsumers"
)

// Get A Logger With Subscription Info
func SubscriptionLogger(logger *zap.Logger, subscription *eventingv1alpha1.Subscription) *zap.Logger {
	return logger.With(zap.String("Namespace", subscription.Namespace), zap.String("Name", subscription.Name))
}

// Create A New ControllerReference Model For The Specified Subscription
func NewSubscriptionControllerRef(subscription *eventingv1alpha1.Subscription) metav1.OwnerReference {
	return *metav1.NewControllerRef(subscription, schema.GroupVersionKind{
		Group:   eventingv1alpha1.SchemeGroupVersion.Group,
		Version: eventingv1alpha1.SchemeGroupVersion.Version,
		Kind:    constants.KnativeSubscriptionKind,
	})
}

// Utility Function To Get The EventStartTime - From Subscription Annotations Only
func EventStartTime(subscription *eventingv1alpha1.Subscription, logger *zap.Logger) string {
	valueString := subscription.Annotations[EventStartTimeAnnotation]
	if len(valueString) > 0 {
		logger.Debug("Found 'EventStartTime' In Subscription", zap.String("Value", valueString))
	} else {
		logger.Debug("Found No 'EventStartTime' In Subscription")
	}
	return valueString
}

// Utility Function To Get The EventRetryIntervalMillis - First From Subscription Annotations And Then From Environment
func EventRetryInitialIntervalMillis(subscription *eventingv1alpha1.Subscription, environment *env.Environment, logger *zap.Logger) int64 {
	valueString := subscription.Annotations[EventRetryInitialIntervalMillisAnnotation]
	if len(valueString) > 0 {
		value, err := strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			logger.Error("Failed To Convert 'EventRetryInitialIntervalMillis' To Int64 - Using Default", zap.Error(err), zap.String("Value", valueString))
			return environment.DefaultEventRetryInitialIntervalMillis
		} else {
			logger.Debug("Found Valid EventRetryInitialIntervalMillis In Subscription", zap.Int64("Value", value))
			return value
		}
	} else {
		logger.Warn("Subscription Annotation 'EventRetryInitialIntervalMillis' Not Specified - Using Default", zap.Int64("Value", environment.DefaultEventRetryInitialIntervalMillis))
		return environment.DefaultEventRetryInitialIntervalMillis
	}
}

// Utility Function To Get The EventRetryTimeMillisMax - First From Subscription Annotations And Then From Environment
func EventRetryTimeMillisMax(subscription *eventingv1alpha1.Subscription, environment *env.Environment, logger *zap.Logger) int64 {
	valueString := subscription.Annotations[EventRetryTimeMillisMaxAnnotation]
	if len(valueString) > 0 {
		value, err := strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			logger.Error("Failed To Convert 'EventRetryTimeMillisMax' To Int64 - Using Default", zap.Error(err), zap.String("Value", valueString))
			return environment.DefaultEventRetryTimeMillisMax
		} else {
			logger.Debug("Found Valid EventRetryTimeMillisMax In Subscription", zap.Int64("Value", value))
			return value
		}
	} else {
		logger.Warn("Subscription Annotation 'EventRetryTimeMillisMax' Not Specified - Using Default", zap.Int64("Value", environment.DefaultEventRetryTimeMillisMax))
		return environment.DefaultEventRetryTimeMillisMax
	}
}

// Utility Function To Get The ExponentialBackoff - First From Subscription Annotations And Then From Environment
func ExponentialBackoff(subscription *eventingv1alpha1.Subscription, environment *env.Environment, logger *zap.Logger) bool {
	valueString := subscription.Annotations[ExponentialBackoffAnnotation]
	if len(valueString) > 0 {
		value, err := strconv.ParseBool(valueString)
		if err != nil {
			logger.Error("Failed To Convert 'ExponentialBackoff' To Bool - Using Default", zap.Error(err), zap.String("Value", valueString))
			return environment.DefaultExponentialBackoff
		} else {
			logger.Debug("Found Valid ExponentialBackoff In Subscription", zap.Bool("Value", value))
			return value
		}
	} else {
		logger.Warn("Subscription Annotation 'ExponentialBackoff' Not Specified - Using Default", zap.Bool("Value", environment.DefaultExponentialBackoff))
		return environment.DefaultExponentialBackoff
	}
}

// Utility Function To Get The KafkaConsumers - First From Subscription Annotations And Then From Environment
func KafkaConsumers(subscription *eventingv1alpha1.Subscription, environment *env.Environment, logger *zap.Logger) int {
	valueString := subscription.Annotations[KafkaConsumersAnnotation]
	if len(valueString) > 0 {
		value, err := strconv.Atoi(valueString)
		if err != nil {
			logger.Error("Failed To Convert 'KafkaConsumers' To Int - Using Default", zap.Error(err), zap.String("Value", valueString))
			return environment.DefaultKafkaConsumers
		} else {
			logger.Debug("Found Valid KafkaConsumers In Subscription", zap.Int("Value", value))
			return value
		}
	} else {
		logger.Warn("Subscription Annotation 'KafkaConsumers' Not Specified - Using Default", zap.Int("Value", environment.DefaultKafkaConsumers))
		return environment.DefaultKafkaConsumers
	}
}

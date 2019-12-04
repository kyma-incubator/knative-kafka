package util

import (
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// FinalizerResult Indicates Whether Finalizer Was Added Or Not
type AddFinalizerResult bool
type RemoveFinalizerResult bool

const (
	FinalizerAlreadyPresent AddFinalizerResult    = false
	FinalizerAdded          AddFinalizerResult    = true
	FinalizerNotPresent     RemoveFinalizerResult = false
	FinalizerRemoved        RemoveFinalizerResult = true
)

// Add Finalizer To The Specified Channel
func AddFinalizerToChannel(channel *kafkav1alpha1.KafkaChannel, finalizerName string) AddFinalizerResult {
	updatedFinalizers, addFinalizerResult := addFinalizer(channel.Finalizers, finalizerName)
	channel.Finalizers = updatedFinalizers
	return addFinalizerResult
}

// Remove Finalizer From The Specified Channel
func RemoveFinalizerFromChannel(channel *kafkav1alpha1.KafkaChannel, finalizerName string) RemoveFinalizerResult {
	updatedFinalizers, removeFinalizerResult := removeFinalizer(channel.Finalizers, finalizerName)
	channel.Finalizers = updatedFinalizers
	return removeFinalizerResult
}

// Add Finalizer To The Specified Subscription
func AddFinalizerToSubscription(subscription *eventingv1alpha1.Subscription, finalizerName string) AddFinalizerResult {
	updatedFinalizers, addFinalizerResult := addFinalizer(subscription.Finalizers, finalizerName)
	subscription.Finalizers = updatedFinalizers
	return addFinalizerResult
}

// Remove Finalizer From The Specified Subscription
func RemoveFinalizerFromSubscription(subscription *eventingv1alpha1.Subscription, finalizerName string) RemoveFinalizerResult {
	updatedFinalizers, removeFinalizerResult := removeFinalizer(subscription.Finalizers, finalizerName)
	subscription.Finalizers = updatedFinalizers
	return removeFinalizerResult
}

// Add The Specified Finalizer Name To The Specified Array Of Finalizers
func addFinalizer(finalizers []string, finalizerName string) ([]string, AddFinalizerResult) {
	finalizersSetString := sets.NewString(finalizers...)
	if finalizersSetString.Has(finalizerName) {
		return finalizersSetString.List(), FinalizerAlreadyPresent
	} else {
		finalizersSetString.Insert(finalizerName)
		return finalizersSetString.List(), FinalizerAdded
	}
}

// Remove The Specified Finalizer Name From The Specified Array Of Finalizers
func removeFinalizer(finalizers []string, finalizerName string) ([]string, RemoveFinalizerResult) {
	finalizersSetString := sets.NewString(finalizers...)
	if finalizersSetString.Has(finalizerName) {
		finalizersSetString.Delete(finalizerName)
		return finalizersSetString.List(), FinalizerRemoved
	} else {
		return finalizersSetString.List(), FinalizerNotPresent
	}
}

package util

import (
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"testing"
)

// Test Constants
const (
	finalizerName1 = "TestFinalizerName1"
	finalizerName2 = "TestFinalizerName2"
	finalizerName3 = "TestFinalizerName3"
)

// Test All Permutations Of AddFinalizerToChannel()
func TestAddFinalizerToChannel(t *testing.T) {

	// Define The TestCase
	type TestCase struct {
		initialFinalizers []string
		addFinalizer      string
	}

	// Create The TestCases
	testCases := []TestCase{
		{initialFinalizers: []string{}, addFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName2}, addFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName2, finalizerName3}, addFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName1, finalizerName2, finalizerName3}, addFinalizer: finalizerName1},
	}

	// Iterate Over All The TestCases
	for _, testCase := range testCases {

		// Determine The Expected AddFinalizerResult Value
		expectedAddFinalizerResult := FinalizerAdded
		if contains(testCase.initialFinalizers, testCase.addFinalizer) {
			expectedAddFinalizerResult = FinalizerAlreadyPresent
		}

		// Create A Test Channel With Initial Finalizers
		channel := &kafkav1alpha1.KafkaChannel{ObjectMeta: v1.ObjectMeta{Finalizers: testCase.initialFinalizers}}

		// Perform The Test - Add The Finalizer To The Channel
		actualAddFinalizerResult := AddFinalizerToChannel(channel, testCase.addFinalizer)

		// Verify Results
		assert.Equal(t, expectedAddFinalizerResult, actualAddFinalizerResult)
		assert.NotNil(t, channel.Finalizers)
		assert.Contains(t, channel.Finalizers, testCase.addFinalizer)
		for _, existingFinalizer := range testCase.initialFinalizers {
			assert.Contains(t, channel.Finalizers, existingFinalizer)
		}
	}
}

// Test All Permutations Of RemoveFinalizerFromChannel()
func TestRemoveFinalizerFromChannel(t *testing.T) {

	// Define The TestCase
	type TestCase struct {
		initialFinalizers []string
		removeFinalizer   string
	}

	// Create The TestCases
	testCases := []TestCase{
		{initialFinalizers: []string{}, removeFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName2}, removeFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName2, finalizerName3}, removeFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName1, finalizerName2, finalizerName3}, removeFinalizer: finalizerName1},
	}

	// Iterate Over All The TestCases
	for _, testCase := range testCases {

		// Determine The Expected AddFinalizerResult Value
		expectedRemoveFinalizerResult := FinalizerNotPresent
		if contains(testCase.initialFinalizers, testCase.removeFinalizer) {
			expectedRemoveFinalizerResult = FinalizerRemoved
		}

		// Create A Test Channel With Initial Finalizers
		channel := &kafkav1alpha1.KafkaChannel{ObjectMeta: v1.ObjectMeta{Finalizers: testCase.initialFinalizers}}

		// Perform The Test - Remove The Finalizer To The Channel
		actualRemoveFinalizerResult := RemoveFinalizerFromChannel(channel, testCase.removeFinalizer)

		// Verify Results
		assert.Equal(t, expectedRemoveFinalizerResult, actualRemoveFinalizerResult)
		assert.NotNil(t, channel.Finalizers)
		assert.NotContains(t, channel.Finalizers, testCase.removeFinalizer)
		for _, existingFinalizer := range testCase.initialFinalizers {
			if existingFinalizer != testCase.removeFinalizer {
				assert.Contains(t, channel.Finalizers, existingFinalizer)
			}
		}
	}
}

// Test All Permutations Of AddFinalizerToSubscription()
func TestAddFinalizerToSubscription(t *testing.T) {

	// Define The TestCase
	type TestCase struct {
		initialFinalizers []string
		addFinalizer      string
	}

	// Create The TestCases
	testCases := []TestCase{
		{initialFinalizers: []string{}, addFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName2}, addFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName2, finalizerName3}, addFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName1, finalizerName2, finalizerName3}, addFinalizer: finalizerName1},
	}

	// Iterate Over All The TestCases
	for _, testCase := range testCases {

		// Determine The Expected AddFinalizerResult Value
		expectedAddFinalizerResult := FinalizerAdded
		if contains(testCase.initialFinalizers, testCase.addFinalizer) {
			expectedAddFinalizerResult = FinalizerAlreadyPresent
		}

		// Create A Test Subscription With Initial Finalizers
		subscription := &eventingv1alpha1.Subscription{ObjectMeta: v1.ObjectMeta{Finalizers: testCase.initialFinalizers}}

		// Perform The Test - Add The Finalizer To The Subscription
		actualAddFinalizerResult := AddFinalizerToSubscription(subscription, testCase.addFinalizer)

		// Verify Results
		assert.Equal(t, expectedAddFinalizerResult, actualAddFinalizerResult)
		assert.NotNil(t, subscription.Finalizers)
		assert.Contains(t, subscription.Finalizers, testCase.addFinalizer)
		for _, existingFinalizer := range testCase.initialFinalizers {
			assert.Contains(t, subscription.Finalizers, existingFinalizer)
		}
	}
}

// Test All Permutations Of RemoveFinalizerFromSubscription()
func TestRemoveFinalizerFromSubscription(t *testing.T) {

	// Define The TestCase
	type TestCase struct {
		initialFinalizers []string
		removeFinalizer   string
	}

	// Create The TestCases
	testCases := []TestCase{
		{initialFinalizers: []string{}, removeFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName2}, removeFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName2, finalizerName3}, removeFinalizer: finalizerName1},
		{initialFinalizers: []string{finalizerName1, finalizerName2, finalizerName3}, removeFinalizer: finalizerName1},
	}

	// Iterate Over All The TestCases
	for _, testCase := range testCases {

		// Determine The Expected AddFinalizerResult Value
		expectedRemoveFinalizerResult := FinalizerNotPresent
		if contains(testCase.initialFinalizers, testCase.removeFinalizer) {
			expectedRemoveFinalizerResult = FinalizerRemoved
		}

		// Create A Test Subscription With Initial Finalizers
		subscription := &eventingv1alpha1.Subscription{ObjectMeta: v1.ObjectMeta{Finalizers: testCase.initialFinalizers}}

		// Perform The Test - Remove The Finalizer To The Subscription
		actualRemoveFinalizerResult := RemoveFinalizerFromSubscription(subscription, testCase.removeFinalizer)

		// Verify Results
		assert.Equal(t, expectedRemoveFinalizerResult, actualRemoveFinalizerResult)
		assert.NotNil(t, subscription.Finalizers)
		assert.NotContains(t, subscription.Finalizers, testCase.removeFinalizer)
		for _, existingFinalizer := range testCase.initialFinalizers {
			if existingFinalizer != testCase.removeFinalizer {
				assert.Contains(t, subscription.Finalizers, existingFinalizer)
			}
		}
	}
}

// Utility Function For Determining Whether Array Contains Specified
func contains(names []string, name string) bool {
	result := false
	for _, value := range names {
		if value == name {
			result = true
			break
		}
	}
	return result
}

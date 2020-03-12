package util

import (
	commonconstants "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/constants"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logtesting "knative.dev/pkg/logging/testing"
	"testing"
)

// Test The SecretLogger() Functionality
func TestSecretLogger(t *testing.T) {

	// Test Logger
	logger := logtesting.TestLogger(t).Desugar()

	// Test Data
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "TestSecretName", Namespace: "TestSecretNamespace"},
	}

	// Perform The Test
	secretLogger := SecretLogger(logger, secret)
	assert.NotNil(t, secretLogger)
	assert.NotEqual(t, logger, secretLogger)
	secretLogger.Info("Testing Secret Logger")
}

// Test The NewSecretOwnerReference() Functionality
func TestNewSecretOwnerReference(t *testing.T) {

	// Test Data
	const secretName = "TestSecretName"
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName},
	}

	// Perform The Test
	controllerRef := NewSecretOwnerReference(secret)

	// Validate Results
	assert.NotNil(t, controllerRef)
	assert.Equal(t, corev1.SchemeGroupVersion.String(), controllerRef.APIVersion)
	assert.Equal(t, constants.SecretKind, controllerRef.Kind)
	assert.Equal(t, secret.ObjectMeta.Name, controllerRef.Name)
	assert.True(t, *controllerRef.BlockOwnerDeletion)
	assert.True(t, *controllerRef.Controller)
}

// Test The FilterKafkaSecrets() Functionality
func TestFilterKafkaSecrets(t *testing.T) {
	performFilterKafkaSecretsTest(t, constants.KnativeEventingNamespace, commonconstants.KafkaSecretLabel, "true", true)
	performFilterKafkaSecretsTest(t, constants.KnativeEventingNamespace, commonconstants.KafkaSecretLabel, "TRUE", true)
	performFilterKafkaSecretsTest(t, "InvalidNamespace", commonconstants.KafkaSecretLabel, "true", false)
	performFilterKafkaSecretsTest(t, constants.KnativeEventingNamespace, "InvalidLabel", "true", false)
	performFilterKafkaSecretsTest(t, constants.KnativeEventingNamespace, commonconstants.KafkaSecretLabel, "false", false)
	performFilterKafkaSecretsTest(t, constants.KnativeEventingNamespace, commonconstants.KafkaSecretLabel, "InvalidValue", false)
}

// Perform A Single Instance Of The FilterKafkaSecrets() Test
func performFilterKafkaSecretsTest(t *testing.T, namespace string, labelKey string, labelValue string, expectedFilterResult bool) {

	// Test Data
	const secretName = "TestSecretName"

	// The Secret To Test
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       constants.SecretKind,
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"TestLabelA": "TestValueA",
				labelKey:     labelValue,
				"TestLabelB": "TestValueB",
			},
		},
	}

	// Get The Filter Function
	result := FilterKafkaSecrets()
	assert.NotNil(t, result)

	// Test The Filter Function
	actualFilterResult := result(secret.GetObjectMeta())
	assert.Equal(t, expectedFilterResult, actualFilterResult)
}

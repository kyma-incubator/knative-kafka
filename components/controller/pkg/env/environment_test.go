package env

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"log"
	"os"
	"strconv"
	"testing"
)

// Test Constants
const (
	serviceAccount                  = "TestServiceAccount"
	metricsPort                     = "9999"
	kafkaProvider                   = "confluent"
	channelImage                    = "TestChannelImage"
	dispatcherImage                 = "TestDispatcherImage"
	kafkaOffsetCommitMessageCount   = "500"
	kafkaOffsetCommitDurationMillis = "2000"

	defaultTenantId                        = "TestDefaultTenantId"
	defaultNumPartitions                   = "7"
	defaultReplicationFactor               = "2"
	defaultRetentionMillis                 = "13579"
	defaultEventRetryInitialIntervalMillis = "246810"
	defaultEventRetryTimeMillisMax         = "1234567890"
	defaultExponentialBackoff              = "true"
	defaultKafkaConsumers                  = "5"

	dispatcherReplicas      = "1"
	dispatcherMemoryRequest = "20Mi"
	dispatcherCpuRequest    = "100m"
	dispatcherMemoryLimit   = "50Mi"
	dispatcherCpuLimit      = "300m"

	channelMemoryRequest = "10Mi"
	channelCpuRquest     = "10m"
	channelMemoryLimit   = "20Mi"
	channelCpuLimit      = "100m"
)

// Define The TestCase Struct
type TestCase struct {
	name                                 string
	serviceAccount                       string
	metricsPort                          string
	kafkaProvider                        string
	channelImage                         string
	dispatcherImage                      string
	kafkaOffsetCommitMessageCount        string
	kafkaOffsetCommitDurationMillis      string
	defaultTenantId                      string
	defaultNumPartitions                 string
	defaultReplicationFactor             string
	defaultRetentionMillis               string
	dispatcherRetryInitialIntervalMillis string
	dispatcherRetryTimeMillisMax         string
	dispatcherRetryExponentialBackoff    string
	defaultKafkaConsumers                string
	dispatcherReplicas                   string
	dispatcherMemoryRequest              string
	dispatcherMemoryLimit                string
	dispatcherCpuRequest                 string
	dispatcherCpuLimit                   string
	channelMemoryRequest                 string
	channelMemoryLimit                   string
	channelCpuRequest                    string
	channelCpuLimit                      string
	expectedError                        error
}

// Test All Permutations Of The GetEnvironment() Functionality
func TestGetEnvironment(t *testing.T) {

	// Get A Logger Reference For Testing
	logger := getLogger()

	// Define The TestCases
	testCases := make([]TestCase, 0, 30)
	testCase := getValidTestCase("Valid Complete Config")
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - ServiceAccount")
	testCase.serviceAccount = ""
	testCase.expectedError = getMissingRequiredEnvironmentVariableError(ServiceAccountEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - MetricsPort")
	testCase.metricsPort = ""
	testCase.expectedError = getMissingRequiredEnvironmentVariableError(MetricsPortEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Invalid Config - MetricsPort")
	testCase.metricsPort = "NAN"
	testCase.expectedError = getInvalidIntegerEnvironmentVariableError(testCase.metricsPort, MetricsPortEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - KafkaProvider")
	testCase.kafkaProvider = ""
	testCase.expectedError = getMissingRequiredEnvironmentVariableError(KafkaProviderEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Invalid Config - KafkaProvider")
	testCase.kafkaProvider = "foo"
	testCase.expectedError = fmt.Errorf("invalid (unknown) value 'foo' for environment variable '%s'", KafkaProviderEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - ChannelImage")
	testCase.channelImage = ""
	testCase.expectedError = getMissingRequiredEnvironmentVariableError(ChannelImageEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - DispatcherImage")
	testCase.dispatcherImage = ""
	testCase.expectedError = getMissingRequiredEnvironmentVariableError(DispatcherImageEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Optional Config - KafkaOffsetCommitMessageCount")
	testCase.kafkaOffsetCommitMessageCount = ""
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Optional Config - KafkaOffsetCommitDurationMillis")
	testCase.kafkaOffsetCommitDurationMillis = ""
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Optional Config - DefaultTenantId")
	testCase.defaultTenantId = ""
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - DefaultNumPartitions")
	testCase.defaultNumPartitions = ""
	testCase.expectedError = getMissingRequiredEnvironmentVariableError(DefaultNumPartitionsEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Invalid Config - DefaultNumPartitions")
	testCase.defaultNumPartitions = "NAN"
	testCase.expectedError = getInvalidIntegerEnvironmentVariableError(testCase.defaultNumPartitions, DefaultNumPartitionsEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - DefaultReplicationFactor")
	testCase.defaultReplicationFactor = ""
	testCase.expectedError = getMissingRequiredEnvironmentVariableError(DefaultReplicationFactorEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Invalid Config - DefaultReplicationFactor")
	testCase.defaultReplicationFactor = "NAN"
	testCase.expectedError = getInvalidIntegerEnvironmentVariableError(testCase.defaultReplicationFactor, DefaultReplicationFactorEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Optional Config - DefaultRetentionMillis")
	testCase.defaultRetentionMillis = ""
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Invalid Config - DefaultRetentionMillis")
	testCase.defaultRetentionMillis = "NAN"
	testCase.expectedError = getInvalidIntegerEnvironmentVariableError(testCase.defaultRetentionMillis, DefaultRetentionMillisEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Optional Config - DispatcherRetryInitialIntervalMillis")
	testCase.dispatcherRetryInitialIntervalMillis = ""
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Invalid Config - DispatcherRetryInitialIntervalMillis")
	testCase.dispatcherRetryInitialIntervalMillis = "NAN"
	testCase.expectedError = getInvalidIntegerEnvironmentVariableError(testCase.dispatcherRetryInitialIntervalMillis, DispatcherRetryInitialIntervalMillisEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Optional Config - DispatcherRetryTimeMillisMax")
	testCase.dispatcherRetryTimeMillisMax = ""
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Invalid Config - DispatcherRetryTimeMillisMax")
	testCase.dispatcherRetryTimeMillisMax = "NAN"
	testCase.expectedError = getInvalidIntegerEnvironmentVariableError(testCase.dispatcherRetryTimeMillisMax, DispatcherRetryTimeMillisMaxEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Optional Config - DispatcherRetryExponentialBackoff")
	testCase.dispatcherRetryExponentialBackoff = ""
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Invalid Config - DispatcherRetryExponentialBackoff")
	testCase.dispatcherRetryExponentialBackoff = "NAB"
	testCase.expectedError = getInvalidBooleanEnvironmentVariableError(testCase.dispatcherRetryExponentialBackoff, DispatcherRetryExponentialBackoffEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - DispatcherReplicas")
	testCase.dispatcherReplicas = ""
	testCase.expectedError = getMissingRequiredEnvironmentVariableError(DispatcherReplicasEnvVarKey)
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Invalid Config - DispatcherReplicas")
	testCase.dispatcherReplicas = "NAN"
	testCase.expectedError = getInvalidIntegerEnvironmentVariableError(testCase.dispatcherReplicas, DispatcherReplicasEnvVarKey)
	testCases = append(testCases, testCase)

	// Loop Over All The TestCases
	for _, testCase := range testCases {

		// (Re)Setup The Environment Variables From TestCase
		os.Clearenv()
		assert.Nil(t, os.Setenv(ServiceAccountEnvVarKey, testCase.serviceAccount))
		if len(testCase.metricsPort) > 0 {
			assert.Nil(t, os.Setenv(MetricsPortEnvVarKey, testCase.metricsPort))
		}
		assert.Nil(t, os.Setenv(KafkaProviderEnvVarKey, testCase.kafkaProvider))
		assert.Nil(t, os.Setenv(ChannelImageEnvVarKey, testCase.channelImage))
		assert.Nil(t, os.Setenv(DispatcherImageEnvVarKey, testCase.dispatcherImage))
		assert.Nil(t, os.Setenv(KafkaOffsetCommitMessageCountEnvVarKey, testCase.kafkaOffsetCommitMessageCount))
		assert.Nil(t, os.Setenv(KafkaOffsetCommitDurationMillisEnvVarKey, testCase.kafkaOffsetCommitDurationMillis))
		assert.Nil(t, os.Setenv(DefaultTenantIdEnvVarKey, testCase.defaultTenantId))
		if len(testCase.defaultNumPartitions) > 0 {
			assert.Nil(t, os.Setenv(DefaultNumPartitionsEnvVarKey, testCase.defaultNumPartitions))
		}
		assert.Nil(t, os.Setenv(DefaultReplicationFactorEnvVarKey, testCase.defaultReplicationFactor))
		assert.Nil(t, os.Setenv(DefaultRetentionMillisEnvVarKey, testCase.defaultRetentionMillis))
		assert.Nil(t, os.Setenv(DispatcherRetryInitialIntervalMillisEnvVarKey, testCase.dispatcherRetryInitialIntervalMillis))
		assert.Nil(t, os.Setenv(DispatcherRetryTimeMillisMaxEnvVarKey, testCase.dispatcherRetryTimeMillisMax))
		assert.Nil(t, os.Setenv(DispatcherRetryExponentialBackoffEnvVarKey, testCase.dispatcherRetryExponentialBackoff))
		if len(testCase.dispatcherReplicas) > 0 {
			assert.Nil(t, os.Setenv(DispatcherReplicasEnvVarKey, testCase.dispatcherReplicas))
		}
		assert.Nil(t, os.Setenv(DispatcherCpuLimitEnvVarKey, testCase.dispatcherCpuLimit))
		assert.Nil(t, os.Setenv(DispatcherCpuRequestEnvVarKey, testCase.dispatcherCpuRequest))
		assert.Nil(t, os.Setenv(DispatcherMemoryLimitEnvVarKey, testCase.dispatcherMemoryLimit))
		assert.Nil(t, os.Setenv(DispatcherMemoryRequestEnvVarKey, testCase.dispatcherMemoryRequest))
		assert.Nil(t, os.Setenv(ChannelMemoryRequestEnvVarKey, testCase.channelMemoryRequest))
		assert.Nil(t, os.Setenv(ChannelCpuRequestEnvVarKey, testCase.channelCpuRequest))
		assert.Nil(t, os.Setenv(ChannelMemoryLimitEnvVarKey, testCase.channelMemoryLimit))
		assert.Nil(t, os.Setenv(ChannelCpuLimitEnvVarKey, testCase.channelCpuLimit))

		// Perform The Test
		environment, err := GetEnvironment(logger)

		// Verify The Results
		if testCase.expectedError == nil {

			assert.Nil(t, err)
			assert.NotNil(t, environment)
			assert.Equal(t, testCase.serviceAccount, environment.ServiceAccount)
			assert.Equal(t, testCase.metricsPort, strconv.Itoa(environment.MetricsPort))
			assert.Equal(t, testCase.channelImage, environment.ChannelImage)
			assert.Equal(t, testCase.dispatcherImage, environment.DispatcherImage)

			if len(testCase.kafkaOffsetCommitMessageCount) > 0 {
				assert.Equal(t, testCase.kafkaOffsetCommitMessageCount, strconv.FormatInt(environment.KafkaOffsetCommitMessageCount, 10))
			} else {
				assert.Equal(t, DefaultKafkaOffsetCommitMessageCount, strconv.FormatInt(environment.KafkaOffsetCommitMessageCount, 10))
			}

			if len(testCase.kafkaOffsetCommitDurationMillis) > 0 {
				assert.Equal(t, testCase.kafkaOffsetCommitDurationMillis, strconv.FormatInt(environment.KafkaOffsetCommitDurationMillis, 10))
			} else {
				assert.Equal(t, DefaultKafkaOffsetCommitDurationMillis, strconv.FormatInt(environment.KafkaOffsetCommitDurationMillis, 10))
			}

			if len(testCase.defaultTenantId) > 0 {
				assert.Equal(t, testCase.defaultTenantId, environment.DefaultTenantId)
			} else {
				assert.Equal(t, DefaultTenantId, environment.DefaultTenantId)
			}

			assert.Equal(t, testCase.defaultNumPartitions, strconv.Itoa(environment.DefaultNumPartitions))
			assert.Equal(t, testCase.defaultReplicationFactor, strconv.Itoa(environment.DefaultReplicationFactor))

			if len(testCase.defaultRetentionMillis) > 0 {
				assert.Equal(t, testCase.defaultRetentionMillis, strconv.FormatInt(environment.DefaultRetentionMillis, 10))
			} else {
				assert.Equal(t, DefaultRetentionMillis, strconv.FormatInt(environment.DefaultRetentionMillis, 10))
			}

			if len(testCase.dispatcherRetryInitialIntervalMillis) > 0 {
				assert.Equal(t, testCase.dispatcherRetryInitialIntervalMillis, strconv.FormatInt(environment.DispatcherRetryInitialIntervalMillis, 10))
			} else {
				assert.Equal(t, DefaultEventRetryInitialIntervalMillis, strconv.FormatInt(environment.DispatcherRetryInitialIntervalMillis, 10))
			}

			if len(testCase.dispatcherRetryTimeMillisMax) > 0 {
				assert.Equal(t, testCase.dispatcherRetryTimeMillisMax, strconv.FormatInt(environment.DispatcherRetryTimeMillisMax, 10))
			} else {
				assert.Equal(t, DefaultEventRetryTimeMillisMax, strconv.FormatInt(environment.DispatcherRetryTimeMillisMax, 10))
			}

			if len(testCase.dispatcherRetryExponentialBackoff) > 0 {
				assert.Equal(t, testCase.dispatcherRetryExponentialBackoff, strconv.FormatBool(environment.DispatcherRetryExponentialBackoff))
			} else {
				assert.Equal(t, DefaultExponentialBackoff, strconv.FormatBool(environment.DispatcherRetryExponentialBackoff))
			}

			assert.Equal(t, testCase.dispatcherReplicas, strconv.Itoa(environment.DispatcherReplicas))

		} else {
			assert.Equal(t, testCase.expectedError, err)
			assert.Nil(t, environment)
		}

	}
}

// Get The Base / Valid Test Case - All Config Specified / No Errors
func getValidTestCase(name string) TestCase {
	return TestCase{
		name:                                 name,
		serviceAccount:                       serviceAccount,
		metricsPort:                          metricsPort,
		kafkaProvider:                        kafkaProvider,
		channelImage:                         channelImage,
		dispatcherImage:                      dispatcherImage,
		kafkaOffsetCommitMessageCount:        kafkaOffsetCommitMessageCount,
		kafkaOffsetCommitDurationMillis:      kafkaOffsetCommitDurationMillis,
		defaultTenantId:                      defaultTenantId,
		defaultNumPartitions:                 defaultNumPartitions,
		defaultReplicationFactor:             defaultReplicationFactor,
		defaultRetentionMillis:               defaultRetentionMillis,
		dispatcherRetryInitialIntervalMillis: defaultEventRetryInitialIntervalMillis,
		dispatcherRetryTimeMillisMax:         defaultEventRetryTimeMillisMax,
		dispatcherRetryExponentialBackoff:    defaultExponentialBackoff,
		defaultKafkaConsumers:                defaultKafkaConsumers,
		dispatcherReplicas:                   dispatcherReplicas,
		dispatcherCpuRequest:                 dispatcherCpuRequest,
		dispatcherCpuLimit:                   dispatcherCpuLimit,
		dispatcherMemoryLimit:                dispatcherMemoryLimit,
		dispatcherMemoryRequest:              dispatcherMemoryRequest,
		channelMemoryRequest:                 channelMemoryRequest,
		channelCpuRequest:                    channelCpuRquest,
		channelMemoryLimit:                   channelMemoryLimit,
		channelCpuLimit:                      channelCpuLimit,
		expectedError:                        nil,
	}
}

// Get The Expected Error Message For A Missing Required Environment Variable
func getMissingRequiredEnvironmentVariableError(envVarKey string) error {
	return fmt.Errorf("missing required environment variable '%s'", envVarKey)
}

// Get The Expected Error Message For An Invalid Integer Environment Variable
func getInvalidIntegerEnvironmentVariableError(value string, envVarKey string) error {
	return fmt.Errorf("invalid (non-integer) value '%s' for environment variable '%s'", value, envVarKey)
}

// Get The Expected Error Message For An Invalid Boolean Environment Variable
func getInvalidBooleanEnvironmentVariableError(value string, envVarKey string) error {
	return fmt.Errorf("invalid (non-boolean) value '%s' for environment variable '%s'", value, envVarKey)
}

// Initialize The Logger - Fatal Exit Upon Error
func getLogger() *zap.Logger {
	logger, err := zap.NewProduction() // For Now Just Use The Default Zap Production Logger
	if err != nil {
		log.Fatalf("Failed To Create New Zap Production Logger: %+v", err)
	}
	return logger
}

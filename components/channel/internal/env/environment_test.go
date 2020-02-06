package env

import (
	"errors"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// Test Constants
const (
	metricsPort   = "TestMetricsPort"
	kafkaBrokers  = "TestKafkaBrokers"
	kafkaUsername = "TestKafkaUsername"
	kafkaPassword = "TestKafkaPassword"
)

// Package Variables
var _ = log.TestLogger() // Force The Use Of The TestLogger!

// Define The TestCase Struct
type TestCase struct {
	name          string
	metricsPort   string
	kafkaBrokers  string
	kafkaUsername string
	kafkaPassword string
	expectError   bool
}

// Test All Permutations Of The GetEnvironment() Functionality
func TestGetEnvironment(t *testing.T) {

	// Define The TestCases
	testCases := make([]TestCase, 0, 30)
	testCase := getValidTestCase("Valid Complete Config")
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - MetricsPort")
	testCase.metricsPort = ""
	testCase.expectError = true
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - KafkaBrokers")
	testCase.kafkaBrokers = ""
	testCase.expectError = true
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - KafkaUsername")
	testCase.kafkaUsername = ""
	testCase.expectError = true
	testCases = append(testCases, testCase)

	testCase = getValidTestCase("Missing Required Config - KafkaPassword")
	testCase.kafkaPassword = ""
	testCase.expectError = true
	testCases = append(testCases, testCase)

	// Loop Over All The TestCases
	for _, testCase := range testCases {

		// (Re)Setup The Environment Variables From TestCase
		os.Clearenv()
		assert.Nil(t, os.Setenv(MetricsPortEnvVarKey, testCase.metricsPort))
		assert.Nil(t, os.Setenv(KafkaBrokersEnvVarKey, testCase.kafkaBrokers))
		assert.Nil(t, os.Setenv(KafkaUsernameEnvVarKey, testCase.kafkaUsername))
		assert.Nil(t, os.Setenv(KafkaPasswordEnvVarKey, testCase.kafkaPassword))

		// Perform The Test
		environment, err := GetEnvironment()

		// Verify The Results
		if testCase.expectError {
			assert.Equal(t, errors.New("invalid / incomplete environment variables"), err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.metricsPort, environment.MetricsPort)
			assert.Equal(t, testCase.kafkaBrokers, environment.KafkaBrokers)
			assert.Equal(t, testCase.kafkaUsername, environment.KafkaUsername)
			assert.Equal(t, testCase.kafkaPassword, environment.KafkaPassword)
		}
	}
}

// Get The Base / Valid Test Case - All Config Specified / No Errors
func getValidTestCase(name string) TestCase {
	return TestCase{
		name:          name,
		metricsPort:   metricsPort,
		kafkaBrokers:  kafkaBrokers,
		kafkaUsername: kafkaUsername,
		kafkaPassword: kafkaPassword,
		expectError:   false,
	}
}

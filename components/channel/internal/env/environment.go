package env

import (
	"errors"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"go.uber.org/zap"
	"os"
)

// Environment Constants
const (

	// Environment Variable Keys
	MetricsPortEnvVarKey   = "METRICS_PORT"
	KafkaBrokersEnvVarKey  = "KAFKA_BROKERS"
	KafkaUsernameEnvVarKey = "KAFKA_USERNAME"
	KafkaPasswordEnvVarKey = "KAFKA_PASSWORD"
)

// The Environment Struct
type Environment struct {
	MetricsPort   string
	KafkaBrokers  string
	KafkaUsername string
	KafkaPassword string
}

// Load & Return The Environment Variables
func GetEnvironment() (Environment, error) {

	// Create The Environment With Current Values
	environment := Environment{
		MetricsPort:   os.Getenv(MetricsPortEnvVarKey),
		KafkaBrokers:  os.Getenv(KafkaBrokersEnvVarKey),
		KafkaUsername: os.Getenv(KafkaUsernameEnvVarKey),
		KafkaPassword: os.Getenv(KafkaPasswordEnvVarKey),
	}

	// Safely Log The ControllerConfig Loaded From Environment Variables
	safeEnvironment := environment
	safeEnvironment.KafkaPassword = ""
	if len(environment.KafkaPassword) > 0 {
		safeEnvironment.KafkaPassword = "********"
	}
	log.Logger().Info("Environment Variables", zap.Any("Environment", safeEnvironment))

	// Validate The Environment Variables
	err := validateEnvironment(environment)

	// Return Results
	return environment, err
}

// Validate The Specified Environment Variables
func validateEnvironment(environment Environment) error {

	valid := validateRequiredEnvironmentVariable(MetricsPortEnvVarKey, environment.MetricsPort) &&
		validateRequiredEnvironmentVariable(KafkaBrokersEnvVarKey, environment.KafkaBrokers) &&
		validateRequiredEnvironmentVariable(KafkaUsernameEnvVarKey, environment.KafkaUsername) &&
		validateRequiredEnvironmentVariable(KafkaPasswordEnvVarKey, environment.KafkaPassword)

	if valid {
		return nil
	} else {
		return errors.New("invalid / incomplete environment variables")
	}
}

// Log The Missing Required Environment Variable With Specified Key
func validateRequiredEnvironmentVariable(key string, value string) bool {
	if len(value) <= 0 {
		log.Logger().Error("Missing Required Environment Variable", zap.String("Key", key))
		return false
	} else {
		return true
	}
}

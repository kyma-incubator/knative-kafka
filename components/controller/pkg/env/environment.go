package env

import (
	"fmt"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/resource"
	"os"
	"strconv"
	"strings"
)

// Package Constants
const (
	// Knative-Kafka Configuration
	HttpPortEnvVarKey        = "HTTP_PORT"
	MetricsPortEnvVarKey     = "METRICS_PORT"
	ChannelImageEnvVarKey    = "CHANNEL_IMAGE"
	DispatcherImageEnvVarKey = "DISPATCHER_IMAGE"

	// The Event Runtime Namespace
	RuntimeNamespaceEnvVarKey = "RUNTIME_NAMESPACE"

	// Kafka Authorization
	KafkaBrokerEnvVarKey   = "KAFKA_BROKERS"
	KafkaUsernameEnvVarKey = "KAFKA_USERNAME"
	KafkaPasswordEnvVarKey = "KAFKA_PASSWORD"

	// Kafka Configuration
	KafkaProviderEnvVarKey                   = "KAFKA_PROVIDER"
	KafkaOffsetCommitMessageCountEnvVarKey   = "KAFKA_OFFSET_COMMIT_MESSAGE_COUNT"
	KafkaOffsetCommitDurationMillisEnvVarKey = "KAFKA_OFFSET_COMMIT_DURATION_MILLIS"
	KafkaTopicEnvVarKey                      = "KAFKA_TOPIC"
	KafkaGroupIdEnvVarKey                    = "KAFKA_GROUP_ID"
	KafkaConsumersEnvVarKey                  = "KAFKA_CONSUMERS"
	KafkaClientIdEnvVarKey                   = "CLIENT_ID"

	// Dispatcher Configuration
	SubscriberUriEnvVarKey        = "SUBSCRIBER_URI"
	ExponentialBackoffEnvVarKey   = "EXPONENTIAL_BACKOFF"
	InitialRetryIntervalEnvVarKey = "INITIAL_RETRY_INTERVAL"
	MaxRetryTimeEnvVarKey         = "MAX_RETRY_TIME"

	// Default Values To Use If Not Available In Env Variables
	DefaultKafkaOffsetCommitMessageCount   = "100"
	DefaultKafkaOffsetCommitDurationMillis = "5000"

	// Default Values To Use If Not Available In Knative Channels Argument
	DefaultTenantIdEnvVarKey          = "DEFAULT_TENANT_ID"
	DefaultNumPartitionsEnvVarKey     = "DEFAULT_NUM_PARTITIONS"
	DefaultReplicationFactorEnvVarKey = "DEFAULT_REPLICATION_FACTOR"
	DefaultRetentionMillisEnvVarKey   = "DEFAULT_RETENTION_MILLIS"

	// Default Values To Use If Not Available In Knative Subscription Annotations
	DefaultEventRetryInitialIntervalMillisEnvVarKey = "DEFAULT_EVENT_RETRY_INITIAL_INTERVAL_MILLIS"
	DefaultEventRetryTimeMillisMaxEnvVarKey         = "DEFAULT_EVENT_RETRY_TIME_MILLIS"
	DefaultExponentialBackoffEnvVarKey              = "DEFAULT_EXPONENTIAL_BACKOFF"
	DefaultKafkaConsumersEnvVarKey                  = "DEFAULT_KAFKA_CONSUMERS"

	// Default Values If Optional Environment Variable Defaults Not Specified
	DefaultTenantId                        = "default-tenant"
	DefaultRetentionMillis                 = "604800000" // 1 Week
	DefaultEventRetryInitialIntervalMillis = "500"       // 0.5 seconds
	DefaultEventRetryTimeMillisMax         = "300000"    // 5 minutes
	DefaultExponentialBackoff              = "true"

	// Kafka Provider Types
	KafkaProviderValueLocal     = "local"
	KafkaProviderValueConfluent = "confluent"
	KafkaProviderValueAzure     = "azure"

	// Dispatcher Resources
	DispatcherCpuRequestEnvVarKey    = "DISPATCHER_CPU_REQUEST"
	DispatcherCpuLimitEnvVarKey      = "DISPATCHER_CPU_LIMIT"
	DispatcherMemoryRequestEnvVarKey = "DISPATCHER_MEMORY_REQUEST"
	DispatcherMemoryLimitEnvVarKey   = "DISPATCHER_MEMORY_LIMIT"

	// Channel Resources
	ChannelMemoryRequestEnvVarKey = "CHANNEL_MEMORY_REQUEST"
	ChannelMemoryLimitEnvVarKey   = "CHANNEL_MEMORY_LIMIT"
	ChannelCpuRequestEnvVarKey    = "CHANNEL_CPU_REQUEST"
	ChannelCpuLimitEnvVarKey      = "CHANNEL_CPU_LIMIT"
)

// Environment Structure
type Environment struct {
	// Knative-Kafka Configuration
	MetricsPort     int    // Required
	ChannelImage    string // Required
	DispatcherImage string // Required

	// Eventing Runtime Namespace
	RuntimeNamespace string // Required

	// Kafka Configuration / Authorization
	KafkaProvider                   string // Required
	KafkaOffsetCommitMessageCount   int64  // Optional
	KafkaOffsetCommitDurationMillis int64  // Optional

	// Default Values To Use If Not Available In Knative Channels Argument
	DefaultTenantId          string // Optional
	DefaultNumPartitions     int    // Required
	DefaultReplicationFactor int    // Required
	DefaultRetentionMillis   int64  // Optional

	// Default Values To Use If Not Available In Knative Subscription Annotations
	DefaultEventRetryInitialIntervalMillis int64 // Optional
	DefaultEventRetryTimeMillisMax         int64 // Optional
	DefaultExponentialBackoff              bool  // Optional
	DefaultKafkaConsumers                  int   // Required

	// Resource configuration
	DispatcherMemoryRequest resource.Quantity // Required
	DispatcherMemoryLimit   resource.Quantity // Required
	DispatcherCpuRequest    resource.Quantity // Required
	DispatcherCpuLimit      resource.Quantity // Required

	// Resource Limits for each Channel Deployment
	ChannelMemoryRequest resource.Quantity // Required
	ChannelMemoryLimit   resource.Quantity // Required
	ChannelCpuRequest    resource.Quantity // Required
	ChannelCpuLimit      resource.Quantity // Required
}

// Get The Environment
func GetEnvironment(logger *zap.Logger) (*Environment, error) {

	// Error Reference
	var err error

	// The ControllerConfig Reference
	environment := &Environment{}

	// Get The Required Metrics Port Config Value & Convert To Int
	metricsPortString, err := getRequiredConfigValue(logger, MetricsPortEnvVarKey)
	if err != nil {
		return nil, err
	} else {
		environment.MetricsPort, err = strconv.Atoi(metricsPortString)
		if err != nil {
			logger.Error("Invalid MetricsPort (Non Integer)", zap.String("Value", metricsPortString), zap.Error(err))
			return nil, fmt.Errorf("invalid (non-integer) value '%s' for environment variable '%s'", metricsPortString, MetricsPortEnvVarKey)
		}
	}

	// Get The Required ChannelImage Config Value
	environment.ChannelImage, err = getRequiredConfigValue(logger, ChannelImageEnvVarKey)
	if err != nil {
		return nil, err
	}

	// Get The Required DispatcherImage Config Value
	environment.DispatcherImage, err = getRequiredConfigValue(logger, DispatcherImageEnvVarKey)
	if err != nil {
		return nil, err
	}

	// Get The Required RuntimeNamespace Config Value
	environment.RuntimeNamespace, err = getRequiredConfigValue(logger, RuntimeNamespaceEnvVarKey)
	if err != nil {
		return nil, err
	}

	// Get The Required Kafka Provider Config Value
	kafkaProviderString, err := getRequiredConfigValue(logger, KafkaProviderEnvVarKey)
	if err != nil {
		return nil, err
	} else {
		switch strings.ToLower(kafkaProviderString) {
		case KafkaProviderValueLocal:
			environment.KafkaProvider = KafkaProviderValueLocal
		case KafkaProviderValueConfluent:
			environment.KafkaProvider = KafkaProviderValueConfluent
		case KafkaProviderValueAzure:
			environment.KafkaProvider = KafkaProviderValueAzure
		default:
			logger.Error("Invalid / Unknown KafkaProvider", zap.String("Value", kafkaProviderString), zap.Error(err))
			return nil, fmt.Errorf("invalid (unknown) value '%s' for environment variable '%s'", kafkaProviderString, KafkaProviderEnvVarKey)
		}
	}

	// Get The Optional KafkaOffsetCommitMessageCount Config Value
	kafkaOffsetCommitMessageCountString := getOptionalConfigValue(logger, KafkaOffsetCommitMessageCountEnvVarKey, DefaultKafkaOffsetCommitMessageCount)
	environment.KafkaOffsetCommitMessageCount, err = strconv.ParseInt(kafkaOffsetCommitMessageCountString, 10, 64)
	if err != nil {
		logger.Error("Invalid KafkaOffsetCommitMessageCount (Non Integer)", zap.String("Value", kafkaOffsetCommitMessageCountString), zap.Error(err))
		return nil, fmt.Errorf("invalid (non-integer) value '%s' for environment variable '%s'", kafkaOffsetCommitMessageCountString, DefaultKafkaOffsetCommitMessageCount)
	}

	// Get The Optional KafkaOffsetCommitDurationMillis Config Value
	kafkaOffsetCommitDurationMillisString := getOptionalConfigValue(logger, KafkaOffsetCommitDurationMillisEnvVarKey, DefaultKafkaOffsetCommitDurationMillis)
	environment.KafkaOffsetCommitDurationMillis, err = strconv.ParseInt(kafkaOffsetCommitDurationMillisString, 10, 64)
	if err != nil {
		logger.Error("Invalid KafkaOffsetCommitDurationMillis (Non Integer)", zap.String("Value", kafkaOffsetCommitDurationMillisString), zap.Error(err))
		return nil, fmt.Errorf("invalid (non-integer) value '%s' for environment variable '%s'", kafkaOffsetCommitDurationMillisString, DefaultKafkaOffsetCommitDurationMillis)
	}

	// Get The Optional DefaultTenantId Config Value
	environment.DefaultTenantId = getOptionalConfigValue(logger, DefaultTenantIdEnvVarKey, DefaultTenantId)

	// Get The Required DefaultNumPartitions Config Value & Convert To Int
	defaultNumPartitionsString, err := getRequiredConfigValue(logger, DefaultNumPartitionsEnvVarKey)
	if err != nil {
		return nil, err
	} else {
		environment.DefaultNumPartitions, err = strconv.Atoi(defaultNumPartitionsString)
		if err != nil {
			logger.Error("Invalid DefaultNumPartitions (Non Integer)", zap.String("Value", defaultNumPartitionsString), zap.Error(err))
			return nil, fmt.Errorf("invalid (non-integer) value '%s' for environment variable '%s'", defaultNumPartitionsString, DefaultNumPartitionsEnvVarKey)
		}
	}

	// Get The Required DefaultReplicationFactor Config Value & Convert To Int
	defaultReplicationFactorString, err := getRequiredConfigValue(logger, DefaultReplicationFactorEnvVarKey)
	if err != nil {
		return nil, err
	} else {
		environment.DefaultReplicationFactor, err = strconv.Atoi(defaultReplicationFactorString)
		if err != nil {
			logger.Error("Invalid DefaultReplicationFactor (Non Integer)", zap.String("Value", defaultReplicationFactorString), zap.Error(err))
			return nil, fmt.Errorf("invalid (non-integer) value '%s' for environment variable '%s'", defaultReplicationFactorString, DefaultReplicationFactorEnvVarKey)
		}
	}

	// Get The Optional DefaultRetentionMillis Config Value & Convert To Int
	defaultRetentionMillisString := getOptionalConfigValue(logger, DefaultRetentionMillisEnvVarKey, DefaultRetentionMillis)
	environment.DefaultRetentionMillis, err = strconv.ParseInt(defaultRetentionMillisString, 10, 64)
	if err != nil {
		logger.Error("Invalid DefaultRetentionMillis (Non Integer)", zap.String("Value", defaultRetentionMillisString), zap.Error(err))
		return nil, fmt.Errorf("invalid (non-integer) value '%s' for environment variable '%s'", defaultRetentionMillisString, DefaultRetentionMillisEnvVarKey)
	}

	// Get The Optional DefaultEventRetryInitialIntervalMillis Config Value & Convert To Int
	defaultEventRetryInitialIntervalMillisString := getOptionalConfigValue(logger, DefaultEventRetryInitialIntervalMillisEnvVarKey, DefaultEventRetryInitialIntervalMillis)
	environment.DefaultEventRetryInitialIntervalMillis, err = strconv.ParseInt(defaultEventRetryInitialIntervalMillisString, 10, 64)
	if err != nil {
		logger.Error("Invalid DefaultEventRetryInitialIntervalMillis (Non Integer)", zap.String("Value", defaultEventRetryInitialIntervalMillisString), zap.Error(err))
		return nil, fmt.Errorf("invalid (non-integer) value '%s' for environment variable '%s'", defaultEventRetryInitialIntervalMillisString, DefaultEventRetryInitialIntervalMillisEnvVarKey)
	}

	// Get The Optional DefaultEventRetryTimeMillisMax Config Value & Convert To Int
	defaultEventRetryTimeMillisMaxString := getOptionalConfigValue(logger, DefaultEventRetryTimeMillisMaxEnvVarKey, DefaultEventRetryTimeMillisMax)
	environment.DefaultEventRetryTimeMillisMax, err = strconv.ParseInt(defaultEventRetryTimeMillisMaxString, 10, 64)
	if err != nil {
		logger.Error("Invalid DefaultEventRetryTimeMillisMax (Non Integer)", zap.String("Value", defaultEventRetryTimeMillisMaxString), zap.Error(err))
		return nil, fmt.Errorf("invalid (non-integer) value '%s' for environment variable '%s'", defaultEventRetryTimeMillisMaxString, DefaultEventRetryTimeMillisMaxEnvVarKey)
	}

	// Get The Optional DefaultExponentialBackoff Config Value & Convert To Bool
	defaultExponentialBackoffString := getOptionalConfigValue(logger, DefaultExponentialBackoffEnvVarKey, DefaultExponentialBackoff)
	environment.DefaultExponentialBackoff, err = strconv.ParseBool(defaultExponentialBackoffString)
	if err != nil {
		logger.Error("Invalid DefaultExponentialBackoff (Non Boolean)", zap.String("Value", defaultExponentialBackoffString), zap.Error(err))
		return nil, fmt.Errorf("invalid (non-boolean) value '%s' for environment variable '%s'", defaultExponentialBackoffString, DefaultExponentialBackoffEnvVarKey)
	}

	// Get The Required DefaultKafkaConsumers Config Value & Convert To Int
	defaultKafkaConsumersString, err := getRequiredConfigValue(logger, DefaultKafkaConsumersEnvVarKey)
	if err != nil {
		return nil, err
	} else {
		environment.DefaultKafkaConsumers, err = strconv.Atoi(defaultKafkaConsumersString)
		if err != nil {
			logger.Error("Invalid DefaultKafkaConsumers (Non Integer)", zap.String("Value", defaultKafkaConsumersString), zap.Error(err))
			return nil, fmt.Errorf("invalid (non-integer) value '%s' for environment variable '%s'", defaultKafkaConsumersString, DefaultKafkaConsumersEnvVarKey)
		}
	}

	// Get The Values For Dispatcher Requests And Limits
	dispatcherMemRequest, err := getRequiredConfigValue(logger, DispatcherMemoryRequestEnvVarKey)
	if err != nil {
		return nil, err
	}
	environment.DispatcherMemoryRequest = resource.MustParse(dispatcherMemRequest)

	dispatcherMemLimit, err := getRequiredConfigValue(logger, DispatcherMemoryLimitEnvVarKey)
	if err != nil {
		return nil, err
	}
	environment.DispatcherMemoryLimit = resource.MustParse(dispatcherMemLimit)

	dispatcherCpuRequest, err := getRequiredConfigValue(logger, DispatcherCpuRequestEnvVarKey)
	if err != nil {
		return nil, err
	}
	environment.DispatcherCpuRequest = resource.MustParse(dispatcherCpuRequest)

	dispatcherCpuLimit, err := getRequiredConfigValue(logger, DispatcherCpuLimitEnvVarKey)
	if err != nil {
		return nil, err
	}
	environment.DispatcherCpuLimit = resource.MustParse(dispatcherCpuLimit)

	// Get The Values For Channel Requests And Limits
	memoryRequest, err := getRequiredConfigValue(logger, ChannelMemoryRequestEnvVarKey)
	if err != nil {
		return nil, err
	} else {
		quantity, err := resource.ParseQuantity(memoryRequest)
		if err != nil {
			message := fmt.Sprintf("Invalid value %s for environment varaible %s, failed to parse as resource.Quantity", memoryRequest, ChannelMemoryRequestEnvVarKey)
			logger.Error(message, zap.Error(err))
			return nil, fmt.Errorf(message)
		}
		environment.ChannelMemoryRequest = quantity
	}

	memoryLimit, err := getRequiredConfigValue(logger, ChannelMemoryLimitEnvVarKey)
	if err != nil {
		return nil, err
	} else {
		quantity, err := resource.ParseQuantity(memoryLimit)
		if err != nil {
			message := fmt.Sprintf("Invalid value %s for environment varaible %s, failed to parse as resource.Quantity", memoryLimit, ChannelMemoryLimitEnvVarKey)
			logger.Error(message, zap.Error(err))
			return nil, fmt.Errorf(message)
		}
		environment.ChannelMemoryLimit = quantity
	}

	cpuRequest, err := getRequiredConfigValue(logger, ChannelCpuRequestEnvVarKey)
	if err != nil {
		return nil, err
	} else {
		quantity, err := resource.ParseQuantity(cpuRequest)
		if err != nil {
			message := fmt.Sprintf("Invalid value %s for environment varaible %s, failed to parse as resource.Quantity", cpuRequest, ChannelCpuRequestEnvVarKey)
			logger.Error(message, zap.Error(err))
			return nil, fmt.Errorf(message)
		}
		environment.ChannelCpuRequest = quantity
	}

	cpuLimit, err := getRequiredConfigValue(logger, ChannelCpuLimitEnvVarKey)
	if err != nil {
		return nil, err
	} else {
		quantity, err := resource.ParseQuantity(cpuLimit)
		if err != nil {
			message := fmt.Sprintf("Invalid value %s for environment varaible %s, failed to parse as resource.Quantity", cpuLimit, ChannelCpuLimitEnvVarKey)
			logger.Error(message, zap.Error(err))
			return nil, fmt.Errorf(message)
		}
		environment.ChannelCpuLimit = quantity
	}

	// Log The ControllerConfig Loaded From Environment Variables
	logger.Info("Environment Variables", zap.Any("Environment", environment))

	// Return The Populated ControllerConfig
	return environment, nil
}

// Get The Specified Required Config Value From OS & Log Errors If Not Present
func getRequiredConfigValue(logger *zap.Logger, key string) (string, error) {
	value := os.Getenv(key)
	if len(value) > 0 {
		return value, nil
	} else {
		logger.Error("Missing Required Environment Variable", zap.String("key", key))
		return "", fmt.Errorf("missing required environment variable '%s'", key)
	}
}

// Get The Specified Optional Config Value From OS
func getOptionalConfigValue(logger *zap.Logger, key string, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) <= 0 {
		logger.Info("Optional Environment Variable Not Specified - Using Default", zap.String("key", key), zap.String("value", defaultValue))
		value = defaultValue
	}
	return value
}

package constants

// Constants
const (
	// Duration Convenience
	MillisPerDay = 24 * 60 * 60 * 1000 // 86400000

	// Kafka Secret Label
	KafkaSecretLabel = "knativekafka.kyma-project.io/kafka-secret"

	// Kafka Secret Keys
	KafkaSecretKeyBrokers   = "brokers"
	KafkaSecretKeyNamespace = "namespace"
	KafkaSecretKeyUsername  = "username"
	KafkaSecretKeyPassword  = "password"

	// Common Kafka Configuration Properties
	ConfigPropertyDebug                   = "debug"
	ConfigPropertyBootstrapServers        = "bootstrap.servers"
	ConfigPropertyRequestTimeoutMs        = "request.timeout.ms"
	ConfigPropertyRequestTimeoutMsValue   = 60000
	ConfigPropertyStatisticsInterval      = "statistics.interval.ms"
	ConfigPropertyStatisticsIntervalValue = 5000

	// Kafka Security/Auth Configuration Properties
	ConfigPropertySecurityProtocol      = "security.protocol"
	ConfigPropertySecurityProtocolValue = "SASL_SSL"
	ConfigPropertySaslMechanisms        = "sasl.mechanisms"
	ConfigPropertySaslMechanismsPlain   = "PLAIN"
	ConfigPropertySaslUsername          = "sasl.username"
	ConfigPropertySaslPassword          = "sasl.password"

	// Kafka Producer Configuration Properties
	ConfigPropertyPartitioner      = "partitioner"
	ConfigPropertyPartitionerValue = "murmur2_random"
	ConfigPropertyIdempotence      = "enable.idempotence"
	ConfigPropertyIdempotenceValue = false // Desirable but not available in Azure EventHubs yet, so disabled for now.

	// Kafka Consumer Configuration Properties
	ConfigPropertyBrokerAddressFamily          = "broker.address.family"
	ConfigPropertyBrokerAddressFamilyValue     = "v4"
	ConfigPropertyGroupId                      = "group.id"
	ConfigPropertyEnableAutoOffset             = "enable.auto.commit"
	ConfigPropertyEnableAutoOffsetValue        = false // Event loss is possible with auto-commit enabled so we will manually commit offsets!
	ConfigPropertyAutoOffsetReset              = "auto.offset.reset"
	ConfigPropertyQueuedMaxMessagesKbytes      = "queued.max.messages.kbytes" // Controls the amount of pre-fetched messages the consumer will pull down per partition
	ConfigPropertyQueuedMaxMessagesKbytesValue = "7000"

	// KafkaTopic Config Keys
	TopicSpecificationConfigRetentionMs = "retention.ms"

	// EventHub Error Codes
	EventHubErrorCodeUnknown      = -2
	EventHubErrorCodeParseFailure = -1
	EventHubErrorCodeConflict     = 409

	// EventHub Constraints
	MaxEventHubNamespaces    = 100
	MaxEventHubsPerNamespace = 10

	// KafkaChannel Constants
	KafkaChannelServiceNameSuffix = "kafkachannel"
)

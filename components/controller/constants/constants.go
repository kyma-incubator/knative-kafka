package constants

const (
	// Knative Eventing Namespace
	KnativeEventingNamespace = "knative-eventing"

	// Knative Controller Naming
	KafkaChannelControllerAgentName      = "kafka-channel-controller"
	KafkaSubscriptionControllerAgentName = "kafka-subscription-controller"

	// CRD Kinds
	KnativeChannelKind      = "Channel"
	KafkaChannelKind        = "KafkaChannel"
	KnativeSubscriptionKind = "Subscription"

	// HTTP Port
	HttpPortName   = "http"
	HttpPortNumber = 80

	// Kafka Secret Label
	KafkaSecretLabel = "knativekafka.kyma-project.io/kafka-secret"

	// Kafka Secret Data Keys
	KafkaSecretDataKeyBrokers  = "brokers"
	KafkaSecretDataKeyUsername = "username"
	KafkaSecretDataKeyPassword = "password"

	// Prometheus MetricsPort
	MetricsPortName = "metrics"

	// Logging Configuration
	LoggingConfigVolumeName = "logging-config"
	LoggingConfigMountPath  = "/etc/knative-kafka"
	LoggingConfigMapName    = "knative-kafka-logging"
)

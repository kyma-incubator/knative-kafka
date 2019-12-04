package kafkasubscription

// Package Level Constants
const (
	// Labels
	DispatcherLabel = "dispatcher" // Dispatcher Label - Used To Mark Deployment As Dispatcher
	ChannelLabel    = "channel"    // Channel Label - Used To Mark Deployment As Related To Channel

	// Prometheus MetricsPort
	MetricsPortName = "metrics"

	// Prometheus ServiceMonitor Selector Labels / Values
	K8sAppDispatcherSelectorLabel = "k8s-app"
	K8sAppDispatcherSelectorValue = "knative-kafka-dispatchers"

	KafkaApiTimeout          = 5000
	EventStartTimeAnnotation = "knativekafka.kyma-project.io/EventStartTime"
)

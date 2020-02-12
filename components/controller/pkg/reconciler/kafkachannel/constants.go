package kafkachannel

// Package Level Constants
const (
	// Labels
	AppLabel                    = "app"
	KafkaChannelLabel           = "kafkachannel"
	KafkaChannelDispatcherLabel = "kafkachannel-dispatcher" // Dispatcher Label - Used To Mark Deployment As Dispatcher
	KafkaChannelChannelLabel    = "kafkachannel-channel"    // Channel Label - Used To Mark Deployment As Related To Channel

	// Prometheus MetricsPort
	MetricsPortName = "metrics"

	// Prometheus ServiceMonitor Selector Labels / Values
	K8sAppChannelSelectorLabel    = "k8s-app"
	K8sAppChannelSelectorValue    = "knative-kafka-channels"
	K8sAppDispatcherSelectorLabel = "k8s-app"
	K8sAppDispatcherSelectorValue = "knative-kafka-dispatchers"
)

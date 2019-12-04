package event

// CoreV1 EventType "Enum" Type
type CoreV1EventType int

// CoreV1 EventType "Enum" Values
const (
	// ClusterChannelProvisioner Reconciliation
	ClusterChannelProvisionerReconciliationFailed CoreV1EventType = iota
	ClusterChannelProvisionerUpdateStatusFailed

	// Channel Updates (Finalizers, Status)
	ChannelUpdateFailed

	// Channel (Kafka Producer) Reconciliation
	ChannelServiceReconciliationFailed
	ChannelDeploymentReconciliationFailed

	// Kafka Topic Reconciliation
	KafkaTopicReconciliationFailed

	// Dispatcher (Kafka Consumer) Reconciliation
	DispatcherK8sServiceReconciliationFailed
	DispatcherDeploymentReconciliationFailed
)

// CoreV1 EventType String Value
func (et CoreV1EventType) String() string {

	// Default The EventType String Value
	eventTypeString := "Unknown Event Type"

	// Map EventTypes To Their String Values
	switch et {
	case ClusterChannelProvisionerReconciliationFailed:
		eventTypeString = "ClusterChannelProvisionerReconciliationFailed"
	case ClusterChannelProvisionerUpdateStatusFailed:
		eventTypeString = "ClusterChannelProvisionerUpdateStatusFailed"
	case ChannelUpdateFailed:
		eventTypeString = "ChannelUpdateFailed"
	case ChannelServiceReconciliationFailed:
		eventTypeString = "ChannelK8sServiceReconciliationFailed"
	case ChannelDeploymentReconciliationFailed:
		eventTypeString = "ChannelDeploymentReconciliationFailed"
	case KafkaTopicReconciliationFailed:
		eventTypeString = "KafkaTopicReconciliationFailed"
	case DispatcherK8sServiceReconciliationFailed:
		eventTypeString = "DispatcherK8sServiceReconciliationFailed"
	case DispatcherDeploymentReconciliationFailed:
		eventTypeString = "DispatcherDeploymentReconciliationFailed"
	}

	// Return The EventType String Value
	return eventTypeString
}

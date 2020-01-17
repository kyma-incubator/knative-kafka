package event

import (
	"testing"
)

// Test The CoreV1 EventType "Enum" String Values
func TestNewChannelLogger(t *testing.T) {
	performEventTypeStringTest(t, ClusterChannelProvisionerReconciliationFailed, "ClusterChannelProvisionerReconciliationFailed")
	performEventTypeStringTest(t, ClusterChannelProvisionerUpdateStatusFailed, "ClusterChannelProvisionerUpdateStatusFailed")
	performEventTypeStringTest(t, ChannelUpdateFailed, "ChannelUpdateFailed")
	performEventTypeStringTest(t, ChannelServiceReconciliationFailed, "ChannelK8sServiceReconciliationFailed")
	performEventTypeStringTest(t, ChannelDeploymentReconciliationFailed, "ChannelDeploymentReconciliationFailed")
	performEventTypeStringTest(t, KafkaTopicReconciliationFailed, "KafkaTopicReconciliationFailed")
	performEventTypeStringTest(t, DispatcherServiceReconciliationFailed, "DispatcherK8sServiceReconciliationFailed")
	performEventTypeStringTest(t, DispatcherDeploymentReconciliationFailed, "DispatcherDeploymentReconciliationFailed")
}

// Perform A Single Instance Of The CoreV1 EventType String Test
func performEventTypeStringTest(t *testing.T, eventType CoreV1EventType, expectedString string) {
	actualString := eventType.String()
	if actualString != expectedString {
		t.Errorf("Expected '%s' but got '%s'", expectedString, actualString)
	}
}

package v1alpha1

import (
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck/v1alpha1"
)

// The Set Of KafkaChannel Status Conditions Required To Be Happy/Ready
var kafkaCondSet = apis.NewLivingConditionSet(
	ConditionTopicReady,
	ConditionChannelDeploymentServiceReady,
	ConditionChannelDeploymentReady,
	ConditionAddressable,
	ConditionChannelServiceReady,
	ConditionDispatcherDeploymentReady)

const (
	// KafkaChannel Condition Addressable has status True when the KafkaChannel meets the Addressable contract and has a non-empty hostname.
	ConditionAddressable apis.ConditionType = "Addressable"

	// KafkaChannel Condition TopicReady has status True when the Kafka Topic associated with the KafkaChannel exists.
	ConditionTopicReady apis.ConditionType = "TopicReady"

	// KafkaChannel Condition ChannelServiceReady has status True when the Service representing the KafkaChannel exists.
	ConditionChannelServiceReady apis.ConditionType = "ChannelServiceReady"

	// KafkaChannel Condition DeploymentServiceReady has status True when the associated Deployment's Service exists.
	ConditionChannelDeploymentServiceReady apis.ConditionType = "ChannelDeploymentServiceReady"

	// KafkaChannel Condition DeploymentReady has status True when the associated Deployment exists.
	ConditionChannelDeploymentReady apis.ConditionType = "ChannelDeploymentReady"

	// KafkaChannel Condition DispatcherDeploymentReady has status True when the associated Deployment exists.
	ConditionDispatcherDeploymentReady apis.ConditionType = "DispatcherDeploymentReady"
)

// GetCondition returns the condition currently associated with the given type, or nil.
func (in *KafkaChannelStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return kafkaCondSet.Manage(in).GetCondition(t)
}

// IsReady returns true if the resource is ready overall.
func (in *KafkaChannelStatus) IsReady() bool {
	return kafkaCondSet.Manage(in).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (in *KafkaChannelStatus) InitializeConditions() {
	kafkaCondSet.Manage(in).InitializeConditions()
}

// SetAddress sets the address (as part of Addressable contract) and marks the correct condition.
func (in *KafkaChannelStatus) SetAddress(url *apis.URL) {
	if in.Address == nil {
		in.Address = &v1alpha1.Addressable{}
	}
	if url != nil {
		in.Address.Hostname = url.Host
		in.Address.URL = url
		kafkaCondSet.Manage(in).MarkTrue(ConditionAddressable)
	} else {
		in.Address.Hostname = ""
		in.Address.URL = nil
		kafkaCondSet.Manage(in).MarkFalse(ConditionAddressable, "emptyHostname", "hostname is the empty string")
	}
}

func (in *KafkaChannelStatus) MarkTopicTrue() {
	kafkaCondSet.Manage(in).MarkTrue(ConditionTopicReady)
}

func (in *KafkaChannelStatus) MarkTopicFailed(reason, messageFormat string, messageA ...interface{}) {
	kafkaCondSet.Manage(in).MarkFalse(ConditionTopicReady, reason, messageFormat, messageA...)
}

func (in *KafkaChannelStatus) MarkChannelServiceTrue() {
	kafkaCondSet.Manage(in).MarkTrue(ConditionChannelServiceReady)
}

func (in *KafkaChannelStatus) MarkChannelServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	kafkaCondSet.Manage(in).MarkFalse(ConditionChannelServiceReady, reason, messageFormat, messageA...)
}

func (in *KafkaChannelStatus) MarkChannelDeploymentServiceTrue() {
	kafkaCondSet.Manage(in).MarkTrue(ConditionChannelDeploymentServiceReady)
}

func (in *KafkaChannelStatus) MarkChannelDeploymentServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	kafkaCondSet.Manage(in).MarkFalse(ConditionChannelDeploymentServiceReady, reason, messageFormat, messageA...)
}

func (in *KafkaChannelStatus) MarkChannelDeploymentTrue() {
	kafkaCondSet.Manage(in).MarkTrue(ConditionChannelDeploymentReady)
}

func (in *KafkaChannelStatus) MarkChannelDeploymentFailed(reason, messageFormat string, messageA ...interface{}) {
	kafkaCondSet.Manage(in).MarkFalse(ConditionChannelDeploymentReady, reason, messageFormat, messageA...)
}

func (in *KafkaChannelStatus) MarkDispatcherDeploymentTrue() {
	kafkaCondSet.Manage(in).MarkTrue(ConditionDispatcherDeploymentReady)
}

func (in *KafkaChannelStatus) MarkDispatcherDeploymentFailed(reason, messageFormat string, messageA ...interface{}) {
	kafkaCondSet.Manage(in).MarkFalse(ConditionDispatcherDeploymentReady, reason, messageFormat, messageA...)
}

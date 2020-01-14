package v1alpha1

import (
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck/v1alpha1"
)

var kafkaCondSet = apis.NewLivingConditionSet(KafkaChannelConditionAddressable, KafkaChannelConditionChannelServiceReady)

const (
	// KafkaChannelConditionAddressable has status true when this KafkaChannel meets
	// the Addressable contract and has a non-empty hostname.
	KafkaChannelConditionAddressable apis.ConditionType = "Addressable"

	// KafkaChannelConditionServiceReady has status True when a k8s Service representing the channel is ready.
	// Because this uses ExternalName, there are no endpoints to check.
	KafkaChannelConditionChannelServiceReady apis.ConditionType = "ChannelServiceReady"
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
		kafkaCondSet.Manage(in).MarkTrue(KafkaChannelConditionAddressable)
	} else {
		in.Address.Hostname = ""
		in.Address.URL = nil
		kafkaCondSet.Manage(in).MarkFalse(KafkaChannelConditionAddressable, "emptyHostname", "hostname is the empty string")
	}
}

func (in *KafkaChannelStatus) MarkChannelServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	kafkaCondSet.Manage(in).MarkFalse(KafkaChannelConditionChannelServiceReady, reason, messageFormat, messageA...)
}

func (in *KafkaChannelStatus) MarkChannelServiceTrue() {
	kafkaCondSet.Manage(in).MarkTrue(KafkaChannelConditionChannelServiceReady)
}

/*
.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingduck "knative.dev/eventing/pkg/apis/duck/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KafkaChannelSpec defines the desired state of KafkaChannel
type KafkaChannelSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	NumPartitions int `json:"numPartitions,omitempty"`

	ReplicationFactor int `json:"replicationFactor,omitempty"`

	RetentionMillis int64 `json:"retentionMillis,omitempty"`

	// Channel conforms to Duck type Subscribable.
	Subscribable *eventingduck.Subscribable `json:"subscribable,omitempty"`

	// TODO - Convert To Channelable Duck Type When Knative-Eventing Moves To v1beta1 (eventingduck & Subscription Reconciler Specifically)
	// Channel conforms to Duck type Channelable
	//eventingduck.ChannelableSpec `json:",inline"`
}

// KafkaChannelStatus defines the observed state of KafkaChannel - Lifted From Knative Samples
type KafkaChannelStatus struct {

	// inherits duck/v1beta1 Status, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last processed by the controller.
	// * Conditions - the latest available observations of a resource's current state.
	duckv1.Status `json:",inline"`

	// KafkaChannel is Addressable. It currently exposes the endpoint as a
	// fully-qualified DNS name which will distribute traffic over the
	// provided targets from inside the cluster.
	//
	// It generally has the form {channel}.{namespace}.svc.{cluster domain name}
	duckv1alpha1.AddressStatus `json:",inline"`

	// Subscribers is populated with the statuses of each of the Channelable's subscribers.
	eventingduck.SubscribableTypeStatus `json:",inline"`

	// TODO - Convert To Channelable Duck Type When Knative-Eventing Moves To v1beta1 (eventingduck & Subscription Reconciler Specifically)
	// Channel conforms to Duck type Channelable.
	//eventingduck.ChannelableStatus `json:",inline"`
}

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KafkaChannel is the Schema for the kafkachannels API
type KafkaChannel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaChannelSpec   `json:"spec,omitempty"`
	Status KafkaChannelStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KafkaChannelList contains a list of KafkaChannel
type KafkaChannelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaChannel `json:"items"`
}

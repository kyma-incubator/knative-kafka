/*
.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingduck "knative.dev/eventing/pkg/apis/duck/v1beta1"
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

	// Channel conforms to Duck type Channelable
	eventingduck.ChannelableSpec `json:",inline"`
}

// KafkaChannelStatus defines the observed state of KafkaChannel - Lifted From Knative Samples
type KafkaChannelStatus struct {
	// Channel conforms to Duck type Channelable.
	eventingduck.ChannelableStatus `json:",inline"`
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

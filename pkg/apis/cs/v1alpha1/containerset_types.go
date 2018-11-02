package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ContainersetSpec defines the desired state of Containerset
type ContainersetSpec struct {
	Replicas int    `json:"replicas"`
	Version  string `json:"version"`
	Env      string `json:"env"`
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// ContainersetStatus defines the observed state of Containerset
type ContainersetStatus struct {
	AvailableReplicas int      `json:"availableReplicas"`
	PodNames          []string `json:"podNames"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Containerset is the Schema for the containersets API
// +k8s:openapi-gen=true
type Containerset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContainersetSpec   `json:"spec,omitempty"`
	Status ContainersetStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ContainersetList contains a list of Containerset
type ContainersetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Containerset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Containerset{}, &ContainersetList{})
}

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProductStatusError represents that the combination of configuration in the BackendSpec
// is not supported by this cluster. This is not a transient error, but
// indicates a state that must be fixed before progress can be made.
// Example: the BackendSpec references non existing internal Metric refenrece
type BackendStatusError string

// BackendSpec defines the desired state of Backend
type BackendSpec struct {
	FriendlyName string `json:"friendlyName,omitempty"`
	// +kubebuilder:validation:Pattern=^https?:\/\/.*$
	PrivateBaseURL string `json:"privateBaseURL,omitempty"`
	// +optional
	Description *string `json:"description,omitempty"`
}

// BackendStatus defines the observed state of Backend
type BackendStatus struct {
	ID         int64  `json:"backendId,omitempty"`
	SystemName string `json:"systemName,omitempty"`
	State      string `json:"state,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
	UpdatedAt  string `json:"updatedAt,omitempty"`

	// In the event that there is a terminal problem reconciling the
	// replicas, both ErrorReason and ErrorMessage will be set. ErrorReason
	// will be populated with a succinct value suitable for machine
	// interpretation, while ErrorMessage will contain a more verbose
	// string suitable for logging and human consumption.
	//
	// These fields should not be set for transitive errors that a
	// controller faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Backend's spec or the configuration of
	// the backend controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the backend controller, or the
	// responsible backend controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Backends
	// can be added as events to the Backend's object and/or logged in the
	// controller's output.
	// +optional
	ErrorReason *BackendStatusError `json:"errorReason,omitempty"`
	// +optional
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Backend is the Schema for the backends API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=backends,scope=Namespaced
type Backend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackendSpec   `json:"spec,omitempty"`
	Status BackendStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BackendList contains a list of Backend
type BackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backend `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Backend{}, &BackendList{})
}

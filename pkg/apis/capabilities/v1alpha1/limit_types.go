package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LimitSpec defines SPEC of the limit object, contains LimitBase and the LimitObjectRef
type LimitSpec struct {
	LimitBase      `json:",inline"`
	LimitObjectRef `json:",inline"`
}

// LimitBase contains the limit period and the max value for said period.
type LimitBase struct {
	Period   string `json:"period"`
	MaxValue int64  `json:"maxValue"`
}

// LimitObjectRef contains he Metric ObjectReference
type LimitObjectRef struct {
	Metric v1.ObjectReference `json:"metricRef"`
}

// LimitStatus defines the observed state of Limit
type LimitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Limit is the Schema for the limits API object
// +k8s:openapi-gen=true
type Limit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LimitSpec   `json:"spec,omitempty"`
	Status LimitStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LimitList contains a list of Limit
type LimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Limit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Limit{}, &LimitList{})
}

package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LimitSpec defines the desired state of Limit
type LimitSpec struct {
	Metric      v1.ObjectReference `json:"metric"`
	Description string `json:"description"`
	Period      string `json:"period"`
	MaxValue    int64  `json:"maxValue"`
}

// LimitStatus defines the observed state of Limit
type LimitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Limit is the Schema for the limits API
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

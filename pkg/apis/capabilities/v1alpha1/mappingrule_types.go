package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MappingRuleSpec defines the desired state of MappingRule
type MappingRuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	MappingRuleBase      `json:",inline"`
	MappingRuleMetricRef `json:",inline"`
}

type MappingRuleBase struct {
	Path      string `json:"path"`
	Method    string `json:"method"`
	Increment int64  `json:"increment"`
}

type MappingRuleMetricRef struct {
	MetricRef v1.ObjectReference `json:"metricRef"`
}

// MappingRuleStatus defines the observed state of MappingRule
type MappingRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MappingRule is the Schema for the mappingrules API
// +k8s:openapi-gen=true
type MappingRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MappingRuleSpec   `json:"spec,omitempty"`
	Status MappingRuleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MappingRuleList contains a list of MappingRule
type MappingRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MappingRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MappingRule{}, &MappingRuleList{})
}

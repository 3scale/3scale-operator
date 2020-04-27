package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ProductStatusError string

const (
	// Represents that the combination of configuration in the ProductSpec
	// is not supported by this cluster. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	//
	// Example: the ProductSpec references non existing internal Metric refenrece
	InvalidInternalMetricReferenceError ProductStatusError = "InvalidInternalMetricReferenceError"
)

// ProductSpec defines the desired state of Product
type ProductSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// ProductStatus defines the observed state of Product
type ProductStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Product is the Schema for the products API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=products,scope=Namespaced
type Product struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProductSpec   `json:"spec,omitempty"`
	Status ProductStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProductList contains a list of Product
type ProductList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Product `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Product{}, &ProductList{})
}

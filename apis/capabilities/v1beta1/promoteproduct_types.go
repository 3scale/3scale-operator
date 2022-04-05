/*
Copyright 2020 Red Hat.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PromoteProductSpec defines the desired state of PromoteProduct
type PromoteProductSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// product id you wish to promote
	ProductId string `json:"productId,omitempty"`
	// promotion configuration version you wish to promote too
	PromoteVersion string `json:"promoteVersion,omitempty"`

	// Environment you wish to promote to staging or production
	PromoteEnvironment string `json:"promoteEnvironment,omitempty"`

	// Promote when true
	Promote bool `json:"promote,omitempty"`

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`
}

// PromoteProductStatus defines the observed state of PromoteProduct
type PromoteProductStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	ProductId *string `json:"productId,omitempty"`
	// +optional
	State *string `json:"state,omitempty"`

	// Environment you have promoted to i.e. staging or production
	//+optional
	PromoteEnvironment string `json:"promoteEnvironment,omitempty"`

	// promotion configuration version you wish to promote too
	//+optional
	PromoteVersion string `json:"promoteVersion,omitempty"`

	// 3scale control plane host
	// +optional
	ProviderAccountHost string `json:"providerAccountHost,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed PromoteProduct Spec.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PromoteProduct is the Schema for the promoteproducts API
type PromoteProduct struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PromoteProductSpec   `json:"spec,omitempty"`
	Status PromoteProductStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PromoteProductList contains a list of PromoteProduct
type PromoteProductList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PromoteProduct `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PromoteProduct{}, &PromoteProductList{})
}

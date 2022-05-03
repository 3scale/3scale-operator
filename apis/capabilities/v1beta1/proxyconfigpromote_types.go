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

// ProxyConfigPromoteSpec defines the desired state of ProxyConfigPromote
type ProxyConfigPromoteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// product CR metadata.name
	ProductCRName string `json:"productCRName,omitempty"`

	// Environment you wish to promote to, if not present defaults to staging and if set to true promotes to production
	// +optional
	Production bool `json:"production,omitempty"`

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`
}

// ProxyConfigPromoteStatus defines the observed state of ProxyConfigPromote
type ProxyConfigPromoteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The id of the product that has been promoted
	//+optional
	ProductId string `json:"productId,omitempty"`

	// The most recent Environment you have promoted to i.e. staging or production
	//+optional
	PromoteEnvironment string `json:"promoteEnvironment,omitempty"`

	// State of promotion i.e. failed or completed
	//+optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ProxyConfigPromote is the Schema for the proxyconfigpromotes API
type ProxyConfigPromote struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxyConfigPromoteSpec   `json:"spec,omitempty"`
	Status ProxyConfigPromoteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProxyConfigPromoteList contains a list of ProxyConfigPromote
type ProxyConfigPromoteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxyConfigPromote `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProxyConfigPromote{}, &ProxyConfigPromoteList{})
}

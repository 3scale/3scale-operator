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
	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	ApplicationAuthReadyConditionType  common.ConditionType = "Ready"
	ApplicationAuthFailedConditionType common.ConditionType = "Failed"
)

// ApplicationAuthSpec defines the desired state of ApplicationAuth
type ApplicationAuthSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// application CR metadata.name
	ApplicationCRName string `json:"applicationCRName"`

	// GenerateSecret Secret is generated if true and empty
	// +optional
	GenerateSecret *bool `json:"generateSecret,omitempty"`

	// AuthSecretRef references account provider credentials
	AuthSecretRef *corev1.LocalObjectReference `json:"authSecretRef"`

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`
}

// ApplicationAuthStatus defines the observed state of ApplicationAuth
type ApplicationAuthStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Current state of the ApplicationAuth resource.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ApplicationAuth is the Schema for the applicationauths API
type ApplicationAuth struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationAuthSpec   `json:"spec,omitempty"`
	Status ApplicationAuthStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationAuthList contains a list of ApplicationAuth
type ApplicationAuthList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationAuth `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApplicationAuth{}, &ApplicationAuthList{})
}

func (a *ApplicationAuthStatus) Equals(other *ApplicationAuthStatus, logger logr.Logger) bool {

	// Marshalling sorts by condition type
	currentMarshaledJSON, _ := a.Conditions.MarshalJSON()
	otherMarshaledJSON, _ := other.Conditions.MarshalJSON()
	if string(currentMarshaledJSON) != string(otherMarshaledJSON) {
		diff := cmp.Diff(string(currentMarshaledJSON), string(otherMarshaledJSON))
		logger.V(1).Info("Conditions not equal", "difference", diff)
		return false
	}

	return true
}

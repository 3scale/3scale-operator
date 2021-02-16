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
	"reflect"

	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	PolicyRegistryKind = "PolicyRegistry"

	// PolicyRegistryInvalidConditionType represents that the combination of configuration
	// in the PolicyRegistrySpec is not supported. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	PolicyRegistryInvalidConditionType common.ConditionType = "Invalid"

	// PolicyRegistryReadyConditionType indicates the policy registry has been successfully synchronized.
	// Steady state
	PolicyRegistryReadyConditionType common.ConditionType = "Ready"

	// PolicyRegistryFailedConditionType indicates that an error occurred during synchronization.
	// The operator will retry.
	PolicyRegistryFailedConditionType common.ConditionType = "Failed"
)

// PolicyRegistrySchemaSpec defines the desired Policy Registry schema
type PolicyRegistrySchemaSpec struct {
	// Name is the name of the policy registry schema
	Name string `json:"name"`

	// Version is the version of the policy registry schema
	Version string `json:"version"`

	// Summary is the summary of the policy registry schema
	Summary string `json:"summary"`

	// Description is an array of description messages for the policy schema registry
	Description *[]string `json:"description,omitempty"`

	// Schema the $schema keyword is used to declare that this is a JSON Schema
	Schema string `json:"$schema"`

	// Configuration defines the structural schema for the policy registry
	// +kubebuilder:pruning:PreserveUnknownFields
	Configuration runtime.RawExtension `json:"configuration"`
}

// PolicyRegistrySpec defines the desired state of PolicyRegistry
type PolicyRegistrySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`

	// Name is the name of the policy registry
	Name string `json:"name"`

	// Version is the version of the policy registry
	Version string `json:"version"`

	// Schema is the schema of the policy registry
	Schema PolicyRegistrySchemaSpec `json:"schema"`
}

// PolicyRegistryStatus defines the observed state of PolicyRegistry
type PolicyRegistryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	ID *int64 `json:"policyRegistryID,omitempty"`

	// ProviderAccountHost contains the 3scale account's provider URL
	// +optional
	ProviderAccountHost string `json:"providerAccountHost,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Backend Spec.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Current state of the policy registry resource.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

func (p *PolicyRegistryStatus) Equals(other *PolicyRegistryStatus, logger logr.Logger) bool {
	if !reflect.DeepEqual(p.ID, other.ID) {
		diff := cmp.Diff(p.ID, other.ID)
		logger.V(1).Info("ID not equal", "difference", diff)
		return false
	}

	if p.ProviderAccountHost != other.ProviderAccountHost {
		diff := cmp.Diff(p.ProviderAccountHost, other.ProviderAccountHost)
		logger.V(1).Info("ProviderAccountHost not equal", "difference", diff)
		return false
	}

	if p.ObservedGeneration != other.ObservedGeneration {
		diff := cmp.Diff(p.ObservedGeneration, other.ObservedGeneration)
		logger.V(1).Info("ObservedGeneration not equal", "difference", diff)
		return false
	}

	// Marshalling sorts by condition type
	currentMarshaledJSON, _ := p.Conditions.MarshalJSON()
	otherMarshaledJSON, _ := other.Conditions.MarshalJSON()
	if string(currentMarshaledJSON) != string(otherMarshaledJSON) {
		diff := cmp.Diff(string(currentMarshaledJSON), string(otherMarshaledJSON))
		logger.V(1).Info("Conditions not equal", "difference", diff)
		return false
	}

	return true
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.providerAccountHost",name="Provider Account",type=string
// +kubebuilder:printcolumn:JSONPath=".status.conditions[?(@.type=='Ready')].status",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".status.policyRegistryID",name="3scale ID",type=integer

// PolicyRegistry is the Schema for the policyregistries API
type PolicyRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicyRegistrySpec   `json:"spec,omitempty"`
	Status PolicyRegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyRegistryList contains a list of PolicyRegistry
type PolicyRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicyRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolicyRegistry{}, &PolicyRegistryList{})
}

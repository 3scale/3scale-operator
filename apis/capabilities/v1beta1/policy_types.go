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
	PolicyKind = "Policy"

	// PolicyInvalidConditionType represents that the combination of configuration
	// in the PolicySpec is not supported. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	PolicyInvalidConditionType common.ConditionType = "Invalid"

	// PolicyReadyConditionType indicates the policy has been successfully synchronized.
	// Steady state
	PolicyReadyConditionType common.ConditionType = "Ready"

	// PolicyFailedConditionType indicates that an error occurred during synchronization.
	// The operator will retry.
	PolicyFailedConditionType common.ConditionType = "Failed"
)

// PolicySchemaSpec defines the desired Policy schema
type PolicySchemaSpec struct {
	// Name is the name of the policy schema
	Name string `json:"name"`

	// Version is the version of the policy schema
	Version string `json:"version"`

	// Summary is the summary of the policy schema
	Summary string `json:"summary"`

	// Description is an array of description messages for the policy schema
	Description *[]string `json:"description,omitempty"`

	// Schema the $schema keyword is used to declare that this is a JSON Schema.
	Schema string `json:"$schema"`

	// Configuration defines the structural schema
	Configuration runtime.RawExtension `json:"configuration"`
}

// PolicySpec defines the desired state of Policy
type PolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`

	// Name is the name of the policy
	Name string `json:"name"`

	// Version is the version of the policy
	Version string `json:"version"`

	// Schema is the schema of the policy
	Schema PolicySchemaSpec `json:"schema"`
}

// PolicyStatus defines the observed state of Policy
type PolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	ID *int64 `json:"policyID,omitempty"`

	// ProviderAccountHost contains the 3scale account's provider URL
	// +optional
	ProviderAccountHost string `json:"providerAccountHost,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Backend Spec.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Current state of the activedoc resource.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

func (p *PolicyStatus) Equals(other *PolicyStatus, logger logr.Logger) bool {
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

// Policy is the Schema for the policies API
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec,omitempty"`
	Status PolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Policy{}, &PolicyList{})
}

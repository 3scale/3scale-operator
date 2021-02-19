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
	"github.com/3scale/3scale-operator/pkg/helper"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	DeveloperUserKind = "DeveloperUser"

	// DeveloperUserInvalidConditionType represents that the combination of configuration
	// in the spec is not supported. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	DeveloperUserInvalidConditionType common.ConditionType = "Invalid"

	// DeveloperUserOrphanConditionType represents that the configuration in the spec
	// contains reference to non existing resource.
	// This is (should be) a transient error, but
	// indicates a state that must be fixed before progress can be made.
	// Example: the DeveloperUserSpec references non existing product resource
	DeveloperUserOrphanConditionType common.ConditionType = "Orphan"

	// DeveloperUserReadyConditionType indicates the activedoc has been successfully synchronized.
	// Steady state
	DeveloperUserReadyConditionType common.ConditionType = "Ready"

	// DeveloperUserFailedConditionType indicates that an error occurred during synchronization.
	// The operator will retry.
	DeveloperUserFailedConditionType common.ConditionType = "Failed"

	// DeveloperUserPasswordSecretField indicates the secret field name with developer user's password
	DeveloperUserPasswordSecretField = "password"
)

// DeveloperUserSpec defines the desired state of DeveloperUser
type DeveloperUserSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Username
	Username string `json:"username"`

	// Email
	Email string `json:"email"`

	// Password
	PasswordCredentialsRef corev1.SecretReference `json:"passwordCredentialsRef"`

	// DeveloperAccountRef is the reference to the parent developer account
	DeveloperAccountRef corev1.LocalObjectReference `json:"developerAccountRef"`

	// State defines the desired state. Defaults to "false", ie, active
	// +optional
	Suspended bool `json:"suspended,omitempty"`

	// Role defines the desired 3scale role. Defaults to "member"
	// +kubebuilder:validation:Enum=admin;member
	// +optional
	Role *string `json:"role,omitempty"`

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`
}

// DeveloperUserStatus defines the observed state of DeveloperUser
type DeveloperUserStatus struct {
	// +optional
	ID *int64 `json:"developerUserID,omitempty"`

	// +optional
	AccountID *int64 `json:"accoundID,omitempty"`

	// +optional
	DeveloperUserState *string `json:"developerUserState,omitempty"`

	// 3scale control plane host
	// +optional
	ProviderAccountHost string `json:"providerAccountHost,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Backend Spec.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Current state of the 3scale backend.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

func (a *DeveloperUserStatus) Equals(other *DeveloperUserStatus, logger logr.Logger) bool {
	if !reflect.DeepEqual(a.ID, other.ID) {
		diff := cmp.Diff(a.ID, other.ID)
		logger.V(1).Info("ID not equal", "difference", diff)
		return false
	}

	if !reflect.DeepEqual(a.AccountID, other.AccountID) {
		diff := cmp.Diff(a.AccountID, other.AccountID)
		logger.V(1).Info("AccountID not equal", "difference", diff)
		return false
	}

	if a.ProviderAccountHost != other.ProviderAccountHost {
		diff := cmp.Diff(a.ProviderAccountHost, other.ProviderAccountHost)
		logger.V(1).Info("ProviderAccountHost not equal", "difference", diff)
		return false
	}

	if !reflect.DeepEqual(a.DeveloperUserState, other.DeveloperUserState) {
		diff := cmp.Diff(a.DeveloperUserState, other.DeveloperUserState)
		logger.V(1).Info("DeveloperUserState not equal", "difference", diff)
		return false
	}

	if a.ObservedGeneration != other.ObservedGeneration {
		diff := cmp.Diff(a.ObservedGeneration, other.ObservedGeneration)
		logger.V(1).Info("ObservedGeneration not equal", "difference", diff)
		return false
	}

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

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DeveloperUser is the Schema for the developerusers API
type DeveloperUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeveloperUserSpec   `json:"spec,omitempty"`
	Status DeveloperUserStatus `json:"status,omitempty"`
}

func (s *DeveloperUser) IsOrphan() bool {
	return s.Status.Conditions.IsTrueFor(DeveloperUserOrphanConditionType)
}

func (s *DeveloperUser) IsAdmin() bool {
	// Role defaults to member
	return s.Spec.Role != nil && *s.Spec.Role == "admin"
}

func (a *DeveloperUser) Validate() field.ErrorList {
	errors := field.ErrorList{}

	// Email validation
	emailFldPath := field.NewPath("spec").Child("email")
	if !helper.IsEmailValid(a.Spec.Email) {
		errors = append(errors, field.Invalid(emailFldPath, a.Spec.Email, "Email address not valid"))
	}

	return errors
}

// +kubebuilder:object:root=true

// DeveloperUserList contains a list of DeveloperUser
type DeveloperUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeveloperUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeveloperUser{}, &DeveloperUserList{})
}

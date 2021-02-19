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
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	DeveloperAccountKind = "DeveloperAccount"

	// DeveloperAccountInvalidConditionType represents that the combination of configuration
	// in the Spec is not supported. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	DeveloperAccountInvalidConditionType common.ConditionType = "Invalid"

	// DeveloperAccountReadyConditionType indicates the account has been successfully synchronized.
	// Steady state
	DeveloperAccountReadyConditionType common.ConditionType = "Ready"

	// DeveloperAccountWaitingConditionType indicates the account is waiting for
	// some async event to happen that is needed to be ready.
	// The operator will retry.
	DeveloperAccountWaitingConditionType common.ConditionType = "Waiting"

	// DeveloperAccountFailedConditionType indicates that an error occurred during synchronization.
	// The operator will retry.
	DeveloperAccountFailedConditionType common.ConditionType = "Failed"
)

// DeveloperAccountSpec defines the desired state of DeveloperAccount
type DeveloperAccountSpec struct {
	// OrgName is the organization name
	OrgName string `json:"orgName"`

	// MonthlyBillingEnabled sets the billing status. Defaults to "true", ie., active
	// +optional
	MonthlyBillingEnabled *bool `json:"monthlyBillingEnabled,omitempty"`

	// MonthlyChargingEnabled Defaults to "true"
	// +optional
	MonthlyChargingEnabled *bool `json:"monthlyChargingEnabled,omitempty"`

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`
}

// DeveloperAccountStatus defines the observed state of DeveloperAccount
type DeveloperAccountStatus struct {
	// +optional
	ID *int64 `json:"accountID,omitempty"`

	// +optional
	AccountState *string `json:"accountState,omitempty"`

	// +optional
	CreditCardStored *bool `json:"creditCardStored,omitempty"`

	// ProviderAccountHost contains the 3scale account's provider URL
	// +optional
	ProviderAccountHost string `json:"providerAccountHost,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Backend Spec.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Current state of the policy resource.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

func (a *DeveloperAccountStatus) Equals(other *DeveloperAccountStatus, logger logr.Logger) bool {
	if !reflect.DeepEqual(a.ID, other.ID) {
		diff := cmp.Diff(a.ID, other.ID)
		logger.V(1).Info("ID not equal", "difference", diff)
		return false
	}

	if a.ProviderAccountHost != other.ProviderAccountHost {
		diff := cmp.Diff(a.ProviderAccountHost, other.ProviderAccountHost)
		logger.V(1).Info("ProviderAccountHost not equal", "difference", diff)
		return false
	}

	if !reflect.DeepEqual(a.AccountState, other.AccountState) {
		diff := cmp.Diff(a.AccountState, other.AccountState)
		logger.V(1).Info("AccountState not equal", "difference", diff)
		return false
	}

	if !reflect.DeepEqual(a.CreditCardStored, other.CreditCardStored) {
		diff := cmp.Diff(a.CreditCardStored, other.CreditCardStored)
		logger.V(1).Info("CreditCardStored not equal", "difference", diff)
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

func (s *DeveloperAccountStatus) IsReady() bool {
	return s.Conditions.IsTrueFor(DeveloperAccountReadyConditionType)
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DeveloperAccount is the Schema for the developeraccounts API
type DeveloperAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeveloperAccountSpec   `json:"spec,omitempty"`
	Status DeveloperAccountStatus `json:"status,omitempty"`
}

func (a *DeveloperAccount) Validate() field.ErrorList {
	errors := field.ErrorList{}
	return errors
}

// +kubebuilder:object:root=true

// DeveloperAccountList contains a list of DeveloperAccount
type DeveloperAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeveloperAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeveloperAccount{}, &DeveloperAccountList{})
}

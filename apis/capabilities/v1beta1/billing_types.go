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
	"k8s.io/apimachinery/pkg/util/validation/field"
	"reflect"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	BillingKind = "Billing"

	// BillingInvalidConditionType represents that the combination of configuration
	// in the Spec is not supported. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	BillingInvalidConditionType common.ConditionType = "Invalid"

	// BillingReadyConditionType indicates the Billing has been successfully synchronized.
	// Steady state
	BillingReadyConditionType common.ConditionType = "Ready"

	// BillingWaitingConditionType indicates the Billing is waiting for
	// some async event to happen that is needed to be ready.
	// The operator will retry.
	BillingWaitingConditionType common.ConditionType = "Waiting"

	// BillingFailedConditionType indicates that an error occurred during synchronization.
	// The operator will retry.
	BillingFailedConditionType common.ConditionType = "Failed"
)

type BillingConfig struct {
	// Bill All Tenants, Defaults to "true"
	// +optional
	BillAllTenants *bool `json:"billAllTenants,omitempty"`

	// Bill Day of Month, Defaults - 1st day of the month
	// +optional
	BillDayOfMonth int `json:"billDayOfMonth,omitempty"`

	// Bill in next cycle, Defaults - false. After the bill - set to false
	// +optional
	BillNow *bool `json:"billNow,omitempty"`
}

// BillingSpec defines the desired state of Billing
type BillingSpec struct {
	// Tenant Account ID
	TenantAccountID *int64 `json:"accountID,omitempty"`

	// ID of the developer account
	// +optional
	DeveloperAccountID *int64 `json:"developerAccountID,omitempty"`

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`

	// BillingConfig describes how to configure Billing process
	BillingConfig BillingConfig `json:"billingConfig"`
}

// BillingStatus defines the observed state of Billing
type BillingStatus struct {

	// +optional
	ID *int64 `json:"billingId,omitempty"`

	// Tenant Account ID
	// +optional
	TenantAccountID *int64 `json:"accountID,omitempty"`

	// ID of the developer account
	// +optional
	DeveloperAccountID *int64 `json:"developerAccountID,omitempty"`

	// +optional
	BillDate *string `json:"billDate,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Billing Spec.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Current state of the Billing.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Billing is the Schema for the Billings API
type Billing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BillingSpec   `json:"spec,omitempty"`
	Status BillingStatus `json:"status,omitempty"`
}

func (a *Billing) Validate() field.ErrorList {
	errors := field.ErrorList{}
	return errors
}

// +kubebuilder:object:root=true

// BillingList contains a list of Billing
type BillingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Billing `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Billing{}, &BillingList{})
}

func (a *BillingStatus) Equals(other *BillingStatus, logger logr.Logger) bool {
	if !reflect.DeepEqual(a.TenantAccountID, other.TenantAccountID) {
		diff := cmp.Diff(a.TenantAccountID, other.TenantAccountID)
		logger.V(1).Info("TenantAccountID not equal", "difference", diff)
		return false
	}

	return true
}

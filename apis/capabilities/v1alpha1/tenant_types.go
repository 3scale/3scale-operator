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

package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/3scale/3scale-operator/pkg/apispkg/common"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// TenantReadyConditionType indicates the tenant has been successfully created.
	// Steady state
	TenantReadyConditionType common.ConditionType = "Ready"

	// TenantFailedConditionType indicates that an error occurred during creation.
	// The operator will retry.
	TenantFailedConditionType common.ConditionType = "Failed"
)

// TenantSpec defines the desired state of Tenant
type TenantSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Username               string             `json:"username"`
	Email                  string             `json:"email"`
	OrganizationName       string             `json:"organizationName"`
	SystemMasterUrl        string             `json:"systemMasterUrl"`
	TenantSecretRef        v1.SecretReference `json:"tenantSecretRef"`
	PasswordCredentialsRef v1.SecretReference `json:"passwordCredentialsRef"`
	MasterCredentialsRef   v1.SecretReference `json:"masterCredentialsRef"`
}

// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	TenantId int64 `json:"tenantId"`
	AdminId  int64 `json:"adminId"`

	// Current state of the tenant resource.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Tenant is the Schema for the tenants API
// +kubebuilder:resource:path=tenants,scope=Namespaced
// +operator-sdk:csv:customresourcedefinitions:displayName="Tenant"
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}

// SetDefaults sets the default vaules for the tenant spec and returns true if the spec was changed
func (t *Tenant) SetDefaults() bool {
	changed := false
	ts := &t.Spec
	if ts.TenantSecretRef.Name == "" {
		ts.TenantSecretRef.Name = fmt.Sprintf("%s-%s", strings.ToLower(t.Name), strings.ToLower(t.Spec.OrganizationName))
		changed = true
	}
	if ts.TenantSecretRef.Namespace == "" {
		ts.TenantSecretRef.Namespace = t.Namespace
		changed = true
	}
	return changed
}

func (t *Tenant) MasterSecretKey() client.ObjectKey {
	namespace := t.Spec.MasterCredentialsRef.Namespace

	if namespace == "" {
		namespace = t.Namespace
	}

	return client.ObjectKey{
		Name:      t.Spec.MasterCredentialsRef.Name,
		Namespace: namespace,
	}
}

func (t *Tenant) AdminPassSecretKey() client.ObjectKey {
	namespace := t.Spec.PasswordCredentialsRef.Namespace

	if namespace == "" {
		namespace = t.Namespace
	}

	return client.ObjectKey{
		Name:      t.Spec.PasswordCredentialsRef.Name,
		Namespace: namespace,
	}
}

func (t *Tenant) TenantSecretKey() client.ObjectKey {
	namespace := t.Spec.TenantSecretRef.Namespace

	if namespace == "" {
		namespace = t.Namespace
	}

	return client.ObjectKey{
		Name:      t.Spec.TenantSecretRef.Name,
		Namespace: namespace,
	}
}

func (b *Tenant) SpecEqual(other *Tenant, logger logr.Logger) bool {
	equal := true

	if !reflect.DeepEqual(b.ObjectMeta, other.ObjectMeta) || !reflect.DeepEqual(b.Spec, other.Spec) {
		equal = false
	}

	return equal
}

func (b *TenantStatus) StatusEqual(other *TenantStatus, logger logr.Logger) bool {
	equal := true

	if b.TenantId != other.TenantId {
		equal = false
	}

	if b.AdminId != other.AdminId {
		equal = false
	}

	if other.Conditions == nil {
		equal = false
	}

	equal = conditionsEqual(TenantReadyConditionType, b.Conditions, other.Conditions) && conditionsEqual(TenantFailedConditionType, b.Conditions, other.Conditions)

	return equal
}

// Compare conditions of a specific type
func conditionsEqual(typeToCompare common.ConditionType, currentConditions, incomingConditions []common.Condition) bool {
	for _, condition1 := range incomingConditions {
		if condition1.Type != typeToCompare {
			continue
		}

		// Find the corresponding condition in conditions2
		var condition2 common.Condition
		for _, c := range currentConditions {
			if c.Type == typeToCompare {
				condition2 = c
				break
			}
		}

		// Compare the status and message
		if condition1.Status != condition2.Status || condition1.Message != condition2.Message {
			return false
		}
	}
	return true
}

// +kubebuilder:object:root=true

// TenantList contains a list of Tenant
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}

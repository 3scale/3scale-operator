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
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

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

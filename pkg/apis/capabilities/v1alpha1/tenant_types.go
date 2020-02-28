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
// +k8s:openapi-gen=true
type TenantSpec struct {
	Username               string             `json:"username"`
	Email                  string             `json:"email"`
	OrganizationName       string             `json:"organizationName"`
	SystemMasterUrl        string             `json:"systemMasterUrl"`
	TenantSecretRef        v1.SecretReference `json:"tenantSecretRef"`
	PasswordCredentialsRef v1.SecretReference `json:"passwordCredentialsRef"`
	MasterCredentialsRef   v1.SecretReference `json:"masterCredentialsRef"`
}

// TenantStatus defines the observed state of Tenant
// +k8s:openapi-gen=true
type TenantStatus struct {
	TenantId int64 `json:"tenantId"`
	AdminId  int64 `json:"adminId"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Tenant is the Schema for the tenants API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=tenants,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Tenant"
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TenantList contains a list of Tenant
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}

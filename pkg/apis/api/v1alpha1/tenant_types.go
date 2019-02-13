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
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	UserName             string             `json:"username"`
	Email                string             `json:"email"`
	OrgName              string             `json:"organizationName"`
	TenantSecretRef      v1.SecretReference `json:"tenantSecretRef"`
	AdminPasswordRef     v1.SecretReference `json:"passwordCredentialsRef"`
	MasterCredentialsRef v1.SecretReference `json:"masterCredentialsRef"`
}

// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	TenantID    int64 `json:"tenantID"`
	AdminUserID int64 `json:"adminID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Tenant is the Schema for the tenants API
// +k8s:openapi-gen=true
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
		ts.TenantSecretRef.Name = fmt.Sprintf("%s-%s", strings.ToLower(t.Name), strings.ToLower(t.Spec.OrgName))
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

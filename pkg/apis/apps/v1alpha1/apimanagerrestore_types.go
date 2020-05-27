package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// APIManagerRestoreSpec defines the desired state of APIManagerRestore
// +k8s:openapi-gen=true
type APIManagerRestoreSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	RestoreSource APIManagerRestoreSource `json:"restoreSource"`
}

type APIManagerRestoreSource struct {
	// +optional
	PersistentVolumeClaim *PersistentVolumeClaimRestoreSource `json:"persistentVolumeClaim,omitempty"`
}

type PersistentVolumeClaimRestoreSource struct {
	ClaimSource v1.PersistentVolumeClaimVolumeSource `json:"claimSource"`
}

// APIManagerRestoreStatus defines the observed state of APIManagerRestore
// +k8s:openapi-gen=true
type APIManagerRestoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +optional
	APIManagerToRestoreRef *v1.LocalObjectReference `json:"apimanagerToRestore,omitempty"`

	// +optional
	Completed *bool `json:"completed,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIManagerRestore is the Schema for the apimanagerrestores API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=apimanagerrestores,scope=Namespaced
type APIManagerRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIManagerRestoreSpec   `json:"spec,omitempty"`
	Status APIManagerRestoreStatus `json:"status,omitempty"`
}

func (a *APIManagerRestore) SetDefaults() (bool, error) {
	return false, nil
}

func (a *APIManagerRestore) RestoreCompleted() bool {
	return a.Status.Completed != nil && *a.Status.Completed
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIManagerRestoreList contains a list of APIManagerRestore
type APIManagerRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIManagerRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIManagerRestore{}, &APIManagerRestoreList{})
}

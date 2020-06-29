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

// APIManagerRestoreSource defines the backup data restore source
// configurability. It is a union type. Only one of the fields can be
// set
type APIManagerRestoreSource struct {
	// +optional
	// Restore data soure configuration
	PersistentVolumeClaim *PersistentVolumeClaimRestoreSource `json:"persistentVolumeClaim,omitempty"`
}

// PersistentVolumeClaimRestoreSource defines the configuration
// of the PersistentVolumeClaim to be used as the restore data source
// for an APIManager restore
type PersistentVolumeClaimRestoreSource struct {
	// PersistentVolumeClaim source of an existing PersistentVolumeClaim.
	// See
	ClaimSource v1.PersistentVolumeClaimVolumeSource `json:"claimSource"`
}

// APIManagerRestoreStatus defines the observed state of APIManagerRestore
// +k8s:openapi-gen=true
type APIManagerRestoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Name of the APIManager to be restored
	// +optional
	APIManagerToRestoreRef *v1.LocalObjectReference `json:"apiManagerToRestoreRef,omitempty"`

	// Set to true when backup has been completed
	// +optional
	Completed *bool `json:"completed,omitempty"`

	// Set to true when main steps have been completed. At this point
	// restore still cannot be considered fully completed due to some remaining
	// post-backup tasks are pending (cleanup, ...)
	// +optional
	MainStepsCompleted *bool `json:"mainStepsCompleted,omitempty"`

	// Restore start time. It is represented in RFC3339 form and is in UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// Restore completion time. It is represented in RFC3339 form and is in UTC.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIManagerRestore represents an APIManager restore
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=apimanagerrestores,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="APIManagerRestore"
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

func (a *APIManagerRestore) MainStepsCompleted() bool {
	return a.Status.MainStepsCompleted != nil && *a.Status.MainStepsCompleted
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

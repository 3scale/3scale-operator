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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// APIManagerRestoreSpec defines the desired state of APIManagerRestore
type APIManagerRestoreSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

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
type APIManagerRestoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

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

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// APIManagerRestore represents an APIManager restore
// +kubebuilder:resource:path=apimanagerrestores,scope=Namespaced
// +operator-sdk:csv:customresourcedefinitions:displayName="APIManagerRestore"
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

// +kubebuilder:object:root=true

// APIManagerRestoreList contains a list of APIManagerRestore
type APIManagerRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIManagerRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIManagerRestore{}, &APIManagerRestoreList{})
}

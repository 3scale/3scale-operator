package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// APIManagerBackupSpec defines the desired state of APIManagerBackup
// +k8s:openapi-gen=true
type APIManagerBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	BackupDestination APIManagerBackupDestination `json:"backupDestination"`
}

type APIManagerBackupDestination struct {
	// +optional
	PersistentVolumeClaim *PersistentVolumeClaimBackupDestination `json:"persistentVolumeClaim,omitempty"`
}

// Ways to define a PVC creation:
// Define VolumeName OR Define Resources. When VolumeName is specified resources is not needed:
// Detailed info:
// https://docs.okd.io/3.11/dev_guide/persistent_volumes.html#persistent-volumes-volumes-and-claim-prebinding
type PersistentVolumeClaimBackupDestination struct {
	// +optional
	Resources *PersistentVolumeClaimResources `json:"resources,omitempty"`
	// +optional
	VolumeName *string `json:"volumeName,omitempty"`
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`
}

type PersistentVolumeClaimResources struct {
	Requests resource.Quantity `json:"requests"` // Should this be a string or a resoure.Quantity? it seems it is serialized as a string
}

type APIManagerBackupConditionType string

// APIManagerBackupCondition describes the state of an APIManagerBackup at a certain point.
type APIManagerBackupCondition struct {
	// Type of APIManagerBackup condition
	Type APIManagerBackupConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown
	Status v1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	// +optional
	Reason *string `json:"reason,omitempty" description:"one-word CamelCase reason for the condition's last transition"`
	// A human readable message indicating details about the transition.
	// +optional
	Message *string `json:"message,omitempty" description:"human-readable message indicating details about last transition"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,7,opt,name=lastTransitionTime"`
}

// APIManagerBackupStatus defines the observed state of APIManagerBackup
// +k8s:openapi-gen=true
type APIManagerBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +optional
	Completed *bool `json:"completed,omitempty"`

	// +optional
	APIManagerSourceName *string `json:"apiManagerSourceName,omitempty"`

	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// +optional
	BackupPersistentVolumeClaimName *string `json:"backupPersistentVolumeClaimName,omitempty"`

	// Represents the latest available observations of an APIManagerBackup's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []APIManagerBackupCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIManagerBackup is the Schema for the apimanagerbackups API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=apimanagerbackups,scope=Namespaced
type APIManagerBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIManagerBackupSpec   `json:"spec,omitempty"`
	Status APIManagerBackupStatus `json:"status,omitempty"`
}

func (a *APIManagerBackup) SetDefaults() (bool, error) {
	return false, nil
}

func (a *APIManagerBackup) BackupCompleted() bool {
	return a.Status.Completed != nil && *a.Status.Completed
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIManagerBackupList contains a list of APIManagerBackup
type APIManagerBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIManagerBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIManagerBackup{}, &APIManagerBackupList{})
}

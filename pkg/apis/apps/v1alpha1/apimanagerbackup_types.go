package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// APIManagerBackupSpec defines the desired state of APIManagerBackup
type APIManagerBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// +optional
	// +kubebuilder:validation:MinLength=1
	APIManagerName *string                `json:"apiManagerName,omitemtpy"`
	BackupSource   APIManagerBackupSource `json:"backupSource"`
}

type APIManagerBackupSource struct {
	// +optional
	SimpleStorageService *APIManagerBackupS3Source `json:"simpleStorageService,omitempty"`
	// +optional
	PersistentVolumeClaim *PersistentVolumeClaimBackupSource `json:"persistentVolumeClaim,omitempty"`
}

type APIManagerBackupS3Source struct {
	CredentialsSecretRef v1.LocalObjectReference `json:"credentialsSecretRef"`
	Bucket               string                  `json:"bucket"`
	// +optional
	Endpoint *string `json:"endpoint,omitempty"`
	// +optional
	ForcePathStyle *bool `json:"forcePathStyle,omitempty"`
	// +optional
	Region *string `json:"region,omitempty"`
	// +optional
	Path *string `json:"path,omitempty"`
}

// Ways to define a PVC creation:
// Define VolumeName OR Define Resources. When VolumeName is specified resources is not needed:
// Detailed info:
// https://docs.okd.io/3.11/dev_guide/persistent_volumes.html#persistent-volumes-volumes-and-claim-prebinding
type PersistentVolumeClaimBackupSource struct {
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
type APIManagerBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Alternatives to keep track of backup state:
	// 1 - Use bools in top-level of status section. Requires namespacing of the names of the variables. PVC and S3 backup share same "space"
	// 2 - Use conditions to codify the same as point 1. Requires namespacing of the names of the variables. PVC and S3 backup share same "space"
	// 3 - Create a struct to represent a state. All
	// In all cases all fields should be optional to avoid backward-incompatible changes

	// +optional
	SecretsAndConfigMapsBackupSubStepFinished *bool `json:"secretsAndConfigMapsBackupSubStepFinished,omitempty"`
	// +optional
	SecretsAndConfigMapsCleanupSubStepFinished *bool `json:"secretsAndConfigMapsCleanupSubStepFinished,omitempty"`
	// +optional
	APIManagerCustomResourceBackupSubStepFinished *bool `json:"apiManagerCustomResourceBackupSubStepFinished,omitempty"`
	// +optional
	APIManagerCustomResourceCleanupSubStepFinished *bool `json:"apiManagerCustomResourceCleanupSubStepFinished,omitempty"`
	// +optional
	SystemFileStorageBackupSubStepFinished *bool `json:"systemFileStorageBackupSubStepFinished,omitempty"`
	// +optional
	SystemFileStorageCleanupSubStepFinished *bool `json:"systemFileStorageCleanupSubStepFinished,omitempty"`
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

func (a *APIManagerBackup) SecretsAndConfigMapsBackupStepFinished() bool {
	return a.SecretsAndConfigMapsBackupSubStepFinished() && a.SecretsAndConfigMapsCleanupSubStepFinished()
}

func (a *APIManagerBackup) SecretsAndConfigMapsBackupSubStepFinished() bool {
	return a.Status.SecretsAndConfigMapsBackupSubStepFinished != nil && *a.Status.SecretsAndConfigMapsBackupSubStepFinished
}

func (a *APIManagerBackup) SecretsAndConfigMapsCleanupSubStepFinished() bool {
	return a.Status.SecretsAndConfigMapsCleanupSubStepFinished != nil && *a.Status.SecretsAndConfigMapsCleanupSubStepFinished
}

func (a *APIManagerBackup) APIManagerCustomResourceBackupStepFinished() bool {
	return a.APIManagerCustomResourceBackupSubStepFinished() && a.APIManagerCustomResourceCleanupSubStepFinished()
}

func (a *APIManagerBackup) APIManagerCustomResourceBackupSubStepFinished() bool {
	return a.Status.APIManagerCustomResourceBackupSubStepFinished != nil && *a.Status.APIManagerCustomResourceBackupSubStepFinished
}

func (a *APIManagerBackup) APIManagerCustomResourceCleanupSubStepFinished() bool {
	return a.Status.APIManagerCustomResourceCleanupSubStepFinished != nil && *a.Status.APIManagerCustomResourceCleanupSubStepFinished
}

func (a *APIManagerBackup) SystemFileStorageBackupStepFinished() bool {
	return a.SystemFileStorageBackupSubStepFinished() && a.SystemFileStorageCleanupSubStepFinished()
}

func (a *APIManagerBackup) SystemFileStorageBackupSubStepFinished() bool {
	return a.Status.SystemFileStorageBackupSubStepFinished != nil && *a.Status.SystemFileStorageBackupSubStepFinished
}

func (a *APIManagerBackup) SystemFileStorageCleanupSubStepFinished() bool {
	return a.Status.SystemFileStorageCleanupSubStepFinished != nil && *a.Status.SystemFileStorageCleanupSubStepFinished
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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

// APIManagerBackupSpec defines the desired state of APIManagerBackup
// +k8s:openapi-gen=true
type APIManagerBackupSpec struct {
	// Backup data destination configuration
	BackupDestination APIManagerBackupDestination `json:"backupDestination"`
}

// APIManagerBackupDestination defines the backup data destination
// configurability. It is a union type. Only one of the fields can be
// set
type APIManagerBackupDestination struct {
	// PersistentVolumeClaim as backup data destination configuration
	// +optional
	PersistentVolumeClaim *PersistentVolumeClaimBackupDestination `json:"persistentVolumeClaim,omitempty"`
}

// PersistentVolumeClaimBackupDestination defines the configuration
// of the PersistentVolumeClaim to be used as the backup data destination
// Ways to define a PVC creation:
// Define VolumeName OR Define Resources. When VolumeName is specified resources is not needed:
// Detailed information:
// https://docs.okd.io/3.11/dev_guide/persistent_volumes.html#persistent-volumes-volumes-and-claim-prebinding
type PersistentVolumeClaimBackupDestination struct {
	// Resources configuration for the backup data PersistentVolumeClaim.
	// Ignored when VolumeName field is set
	// +optional
	Resources *PersistentVolumeClaimResources `json:"resources,omitempty"`
	// Name of an existing PersistentVolume to be bound to the
	// backup data PersistentVolumeClaim
	// +optional
	VolumeName *string `json:"volumeName,omitempty"`
	// Storage class to be used by the PersistentVolumeClaim. Ignored
	// when VolumeName field is set
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`
}

// PersistentVolumeClaimResources defines the resources configuration
// of the backup data destination PersistentVolumeClaim
type PersistentVolumeClaimResources struct {
	// Storage Resource requests to be used on the PersistentVolumeClaim.
	// To learn more about resource requests see:
	// https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	Requests resource.Quantity `json:"requests"` // Should this be a string or a resoure.Quantity? it seems it is serialized as a string
}

// APIManagerBackupStatus defines the observed state of APIManagerBackup
// +k8s:openapi-gen=true
type APIManagerBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Set to true when backup has been completed
	// +optional
	Completed *bool `json:"completed,omitempty"`

	// Set to true when main steps have been completed. At this point
	// backup still cannot be considered  fully completed due to some remaining
	// post-backup tasks are pending (cleanup, ...)
	// +optional
	MainStepsCompleted *bool `json:"mainStepsCompleted,omitempty"`

	// Name of the APIManager from which the backup has been performed
	// +optional
	APIManagerSourceName *string `json:"apiManagerSourceName,omitempty"`

	// Backup start time. It is represented in RFC3339 form and is in UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// Backup completion time. It is represented in RFC3339 form and is in UTC.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Name of the backup data PersistentVolumeClaim. Only set when
	// PersistentVolumeClaim is used as the backup data destination
	// +optional
	BackupPersistentVolumeClaimName *string `json:"backupPersistentVolumeClaimName,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIManagerBackup represents an APIManager backup
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=apimanagerbackups,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="APIManagerBackup"
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

func (a *APIManagerBackup) MainStepsCompleted() bool {
	return a.Status.MainStepsCompleted != nil && *a.Status.MainStepsCompleted
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

package v1alpha1

import (
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/olm"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// APIManagerSpec defines the desired state of APIManager
// +k8s:openapi-gen=true
type APIManagerSpec struct {
	APIManagerCommonSpec `json:",inline"`
	// +optional
	Apicast *ApicastSpec `json:"apicast,omitempty"`
	// +optional
	Backend *BackendSpec `json:"backend,omitempty"`
	// +optional
	System *SystemSpec `json:"system,omitempty"`
	// +optional
	Zync *ZyncSpec `json:"zync,omitempty"`
	// +optional
	HighAvailability *HighAvailabilitySpec `json:"highAvailability,omitempty"`
}

// APIManagerStatus defines the observed state of APIManager
// +k8s:openapi-gen=true
type APIManagerStatus struct {
	Conditions  []APIManagerCondition `json:"conditions,omitempty" protobuf:"bytes,4,rep,name=conditions"`
	Deployments olm.DeploymentStatus  `json:"deployments"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIManager is the Schema for the apimanagers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type APIManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIManagerSpec   `json:"spec,omitempty"`
	Status APIManagerStatus `json:"status,omitempty"`
}

type APIManagerConditionType string

const (
	// Ready means the APIManager is available. This is, when all of its
	// elements are up and running
	APIManagerReady APIManagerConditionType = "Ready"
	// Progressing means the APIManager is being deployed
	APIManagerProgressing APIManagerConditionType = "Progressing"
)

type APIManagerCondition struct {
	Type   APIManagerConditionType `json:"type" description:"type of APIManager condition"`
	Status v1.ConditionStatus      `json:"status" description:"status of the condition, one of True, False, Unknown"` //TODO should be a custom ConditionStatus or the core v1 one?

	// The Reason, Message, LastHeartbeatTime and LastTransitionTime fields are
	// optional. Unless we really use them they should directly not be used even
	// if they are optional

	// +optional
	//Reason *string `json:"reason,omitempty" description:"one-word CamelCase reason for the condition's last transition"`
	// +optional
	//Message *string `json:"message,omitempty" description:"human-readable message indicating details about last transition"`

	// +optional
	//LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty" description:"last time we got an update on a given condition"` // TODO the Kubernetes API convention guide says *unversioned.Time should be used but that seems to be a client-side package. I've seen that objects like PersistentVolumeClaim use metav1.Time
	// +optional
	//metav1.Time        `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	//LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" description:"last time the condition transit from one status to another"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIManagerList contains a list of APIManager
type APIManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIManager `json:"items"`
}

type APIManagerCommonSpec struct {
	WildcardDomain string `json:"wildcardDomain"`
	// +optional
	AppLabel *string `json:"appLabel,omitempty"`
	// +optional
	TenantName *string `json:"tenantName,omitempty"`
	// +optional
	ImageStreamTagImportInsecure *bool `json:"imageStreamTagImportInsecure,omitempty"`
	// +optional
	ResourceRequirementsEnabled *bool `json:"resourceRequirementsEnabled,omitempty"`
}

type ApicastSpec struct {
	// +optional
	ApicastManagementAPI *string `json:"managementAPI,omitempty"`
	// +optional
	OpenSSLVerify *bool `json:"openSSLVerify,omitempty"`
	// +optional
	IncludeResponseCodes *bool `json:"responseCodes,omitempty"`
	// +optional
	RegistryURL *string `json:"registryURL,omitempty"`
	// +optional
	Image *string `json:"image,omitempty"`
}

type BackendSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`

	// +optional
	RedisImage *string `json:"redisImage,omitempty"`
}

type SystemSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`

	// +optional
	MemcachedImage *string `json:"memcachedImage,omitempty"`

	// +optional
	RedisImage *string `json:"redisImage,omitempty"`

	// TODO should this field be optional? We have different approaches in Kubernetes.
	// For example, in v1.Volume it is optional and there's an implied behaviour
	// on which one is the default VolumeSource of the ones available. However,
	// on v1.Probe the Handler field is mandatory and says that one of the
	// available values and only one should be specified (it mandates to write
	// something)

	// +optional
	FileStorageSpec *SystemFileStorageSpec `json:"fileStorage,omitempty"`

	// TODO should union fields be optional?

	// +optional
	DatabaseSpec *SystemDatabaseSpec `json:"database,omitempty"`
}

type SystemFileStorageSpec struct {
	// Union type. Only one of the fields can be set.
	// +optional
	PVC *SystemPVCSpec `json:"persistentVolumeClaim,omitempty"`
	// +optional
	S3 *SystemS3Spec `json:"amazonSimpleStorageService,omitempty"`
}

type SystemPVCSpec struct {
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
}

type SystemS3Spec struct {
	AWSBucket      string                  `json:"awsBucket"`
	AWSRegion      string                  `json:"awsRegion"`
	AWSCredentials v1.LocalObjectReference `json:"awsCredentialsSecret"`
}

type SystemDatabaseSpec struct {
	// Union type. Only one of the fields can be set
	// +optional
	MySQL *SystemMySQLSpec `json:"mysql,omitempty"`
	// +optional
	PostgreSQL *SystemPostgreSQLSpec `json:"postgresql,omitempty"`
}

type SystemMySQLSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`
}

type SystemPostgreSQLSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`
}

type ZyncSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`
	// +optional
	PostgreSQLImage *string `json:"postgreSQLImage,omitempty"`
}

type HighAvailabilitySpec struct {
	Enabled bool `json:"enabled,omitempty"`
}

func init() {
	SchemeBuilder.Register(&APIManager{}, &APIManagerList{})
}

// SetDefaults sets the default values for the APIManager spec and returns true if the spec was changed
func (apimanager *APIManager) SetDefaults() (bool, error) {
	var err error
	changed := false
	commonChanged := apimanager.setAPIManagerCommonSpecDefaults()
	apicastChanged := apimanager.setApicastSpecDefaults()
	systemChanged, err := apimanager.setSystemSpecDefaults()

	if commonChanged || apicastChanged || systemChanged {
		changed = true
	}

	return changed, err
}

func (apimanager *APIManager) setApicastSpecDefaults() bool {
	changed := false
	spec := &apimanager.Spec

	defaultApicastManagementAPI := "status"
	defaultApicastOpenSSLVerify := false
	defaultApicastResponseCodes := true
	defaultApicastRegistryURL := "http://apicast-staging:8090/policies"
	if spec.Apicast == nil {

		changed = true
		spec.Apicast = &ApicastSpec{
			ApicastManagementAPI: &defaultApicastManagementAPI,
			OpenSSLVerify:        &defaultApicastOpenSSLVerify,
			IncludeResponseCodes: &defaultApicastResponseCodes,
			RegistryURL:          &defaultApicastRegistryURL,
		}
	} else {
		if spec.Apicast.ApicastManagementAPI == nil {
			spec.Apicast.ApicastManagementAPI = &defaultApicastManagementAPI
		}
		if spec.Apicast.OpenSSLVerify == nil {
			spec.Apicast.OpenSSLVerify = &defaultApicastOpenSSLVerify
		}
		if spec.Apicast.IncludeResponseCodes == nil {
			spec.Apicast.IncludeResponseCodes = &defaultApicastResponseCodes
		}
		if spec.Apicast.RegistryURL == nil {
			spec.Apicast.RegistryURL = &defaultApicastRegistryURL
		}
	}

	return changed
}

func (apimanager *APIManager) setAPIManagerCommonSpecDefaults() bool {
	changed := false
	spec := &apimanager.Spec

	if spec.AppLabel == nil {
		defaultAppLabel := "3scale-api-management"
		spec.AppLabel = &defaultAppLabel
		changed = true
	}

	if spec.TenantName == nil {
		defaultTenantName := "3scale"
		spec.TenantName = &defaultTenantName
		changed = true
	}

	if spec.ImageStreamTagImportInsecure == nil {
		defaultImageStreamTagImportInsecure := false
		spec.ImageStreamTagImportInsecure = &defaultImageStreamTagImportInsecure
		changed = true
	}

	if spec.ResourceRequirementsEnabled == nil {
		defaultResourceRequirementsEnabled := true
		spec.ResourceRequirementsEnabled = &defaultResourceRequirementsEnabled
		changed = true
	}

	// TODO do something with mandatory parameters?
	// TODO check that only compatible ProductRelease versions are compatible?

	return changed
}

func (apimanager *APIManager) setSystemSpecDefaults() (bool, error) {
	// TODO fix how should be shown
	changed := false
	spec := &apimanager.Spec

	if spec.System == nil {
		spec.System = &SystemSpec{}
	}

	changed, err := apimanager.setSystemFileStorageSpecDefaults()
	if err != nil {
		return changed, err
	}

	changed, err = apimanager.setSystemDatabaseSpecDefaults()
	if err != nil {
		return changed, err
	}

	return changed, nil
}

func (apimanager *APIManager) setSystemFileStorageSpecDefaults() (bool, error) {
	changed := false
	systemSpec := apimanager.Spec.System

	defaultFileStorageSpec := &SystemFileStorageSpec{
		PVC: &SystemPVCSpec{
			StorageClassName: nil,
		},
	}
	if systemSpec.FileStorageSpec == nil {
		systemSpec.FileStorageSpec = defaultFileStorageSpec
		changed = true
	} else {
		if systemSpec.FileStorageSpec.PVC != nil && systemSpec.FileStorageSpec.S3 != nil {
			return changed, fmt.Errorf("Only one FileStorage can be chosen at the same time")
		}
		if systemSpec.FileStorageSpec.PVC == nil && systemSpec.FileStorageSpec.S3 == nil {
			systemSpec.FileStorageSpec.PVC = defaultFileStorageSpec.PVC
			changed = true
		}
	}

	return changed, nil
}

func (apimanager *APIManager) setSystemDatabaseSpecDefaults() (bool, error) {
	changed := false
	systemSpec := apimanager.Spec.System
	defaultDatabaseSpec := &SystemDatabaseSpec{
		MySQL: &SystemMySQLSpec{
			Image: nil,
		},
	}

	if systemSpec.DatabaseSpec == nil {
		systemSpec.DatabaseSpec = defaultDatabaseSpec
		changed = true
	} else {
		if systemSpec.DatabaseSpec.MySQL != nil && systemSpec.DatabaseSpec.PostgreSQL != nil {
			return changed, fmt.Errorf("Only one System Database can be chosen at the same time")
		}
		if systemSpec.DatabaseSpec.MySQL == nil && systemSpec.DatabaseSpec.PostgreSQL == nil {
			systemSpec.DatabaseSpec.MySQL = defaultDatabaseSpec.MySQL
		}
	}

	return changed, nil
}

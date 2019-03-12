package v1alpha1

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// We use pointers in a spec field when the field is optional, to allow differentiate
// Between an unset value from the zero value of the field.
// This is a common convention
// used in kubernetes: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#optional-vs-required

// APIManagerSpec defines the desired state of APIManager
type APIManagerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

	APIManagerCommonSpec `json:",inline"`
	// +optional
	ApicastSpec *ApicastSpec `json:"apicast,omitempty"`
	// +optional
	BackendSpec *BackendSpec `json:"backend,omitempty"`
	// +optional
	SystemSpec *SystemSpec `json:"system,omitempty"`
	// +optional
	ZyncSpec *ZyncSpec `json:"zync,omitempty"`
	// +optional
	WildcardRouterSpec *WildcardRouterSpec `json:"wildcardRouter,omitempty"`
	// +optional
	HighAvailabilitySpec *HighAvailabilitySpec `json:"highAvailability,omitempty"`
}

// APIManagerStatus defines the observed state of APIManager
type APIManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Conditions []APIManagerCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,4,rep,name=conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIManager is the Schema for the apimanagers API
// +k8s:openapi-gen=true
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
	ProductVersion product.Version `json:"productVersion"`
	WildcardDomain string          `json:"wildcardDomain"`
	// +optional
	AppLabel *string `json:"appLabel,omitempty"`
	// +optional
	TenantName *string `json:"tenantName,omitempty"`
	// +optional
	WildcardPolicy *string `json:"wildcardPolicy,omitempty"`
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
	AWSBucket         string                  `json:"awsBucket"`
	AWSRegion         string                  `json:"awsRegion"`
	AWSCredentials    v1.LocalObjectReference `json:"awsCredentialsSecret"`
	FileUploadStorage string                  `json:"fileUploadStorage"`
}

type SystemDatabaseSpec struct {
	// Union type. Only one of the fields can be set
	// +optional
	MySQLSpec *SystemMySQLSpec `json:"mysql,omitempty"`
	// In the future PostgreSQL support will be added
}

type SystemMySQLSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`
}

type ZyncSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`
	// +optional
	PostgreSQLImage *string `json:"postgreSQLImage,omitempty"`
}

type WildcardRouterSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`
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
	changed = apimanager.setAPIManagerCommonSpecDefaults()
	changed = apimanager.setApicastSpecDefaults()
	changed, err = apimanager.setSystemSpecDefaults()

	return changed, err
}

func (apimanager *APIManager) setApicastSpecDefaults() bool {
	changed := false
	spec := &apimanager.Spec

	defaultApicastManagementAPI := "status"
	defaultApicastOpenSSLVerify := false
	defaultApicastResponseCodes := true
	defaultApicastRegistryURL := "http://apicast-staging:8090/policies"
	if spec.ApicastSpec == nil {

		changed = true
		spec.ApicastSpec = &ApicastSpec{
			ApicastManagementAPI: &defaultApicastManagementAPI,
			OpenSSLVerify:        &defaultApicastOpenSSLVerify,
			IncludeResponseCodes: &defaultApicastResponseCodes,
			RegistryURL:          &defaultApicastRegistryURL,
		}
	} else {
		if spec.ApicastSpec.ApicastManagementAPI == nil {
			spec.ApicastSpec.ApicastManagementAPI = &defaultApicastManagementAPI
		}
		if spec.ApicastSpec.OpenSSLVerify == nil {
			spec.ApicastSpec.OpenSSLVerify = &defaultApicastOpenSSLVerify
		}
		if spec.ApicastSpec.IncludeResponseCodes == nil {
			spec.ApicastSpec.IncludeResponseCodes = &defaultApicastResponseCodes
		}
		if spec.ApicastSpec.RegistryURL == nil {
			spec.ApicastSpec.RegistryURL = &defaultApicastRegistryURL
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

	if spec.WildcardPolicy == nil {
		defaultWildcardPolicy := "None" //TODO should be a set of predefined values (a custom type enum-like to be used)
		spec.WildcardPolicy = &defaultWildcardPolicy
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

	if spec.SystemSpec == nil {
		spec.SystemSpec = &SystemSpec{}
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
	systemSpec := apimanager.Spec.SystemSpec

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
			systemSpec.FileStorageSpec.PVC = &SystemPVCSpec{
				StorageClassName: nil,
			}
			changed = true
		}
	}

	return changed, nil
}

func (apimanager *APIManager) setSystemDatabaseSpecDefaults() (bool, error) {
	changed := false
	systemSpec := apimanager.Spec.SystemSpec
	defaultDatabaseSpec := &SystemDatabaseSpec{
		MySQLSpec: &SystemMySQLSpec{
			Image: nil,
		},
	}

	if systemSpec.DatabaseSpec == nil {
		systemSpec.DatabaseSpec = defaultDatabaseSpec
		changed = true
	}

	return changed, nil
}

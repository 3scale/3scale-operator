package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
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

	AmpRelease string `json:"ampRelease"`

	// +optional
	AppLabel *string `json:"appLabel,omitempty"`

	// +optional
	TenantName *string `json:"tenantName,omitempty"`

	// +optional
	RwxStorageClass *string `json:"rwxStorageClass,omitempty"` // TODO is this correct? maybe a pointer to String?

	// +optional
	AmpBackendImage *string `json:"ampBackendImage,omitempty"`

	// +optional
	AmpZyncImage *string `json:"ampZyncImage,omitempty"`

	// +optional
	AmpApicastImage *string `json:"ampApicastImage,omitempty"`

	// +optional
	AmpRouterImage *string `json:"ampRouterImage,omitempty"`

	// +optional
	AmpSystemImage *string `json:"ampSystemImage,omitempty"`

	// +optional
	PostgreSQLImage *string `json:"postgreSQLImage,omitempty"`

	// +optional
	MysqlImage *string `json:"mysqlImage,omitempty"`

	// +optional
	MemcachedImage *string `json:"memcachedImage,omitempty"`

	// +optional
	ImageStreamTagImportInsecure *bool `json:"imageStreamTagImportInsecure,omitempty"`

	// +optional
	RedisImage *string `json:"redisImage,omitempty"`

	// TODO. This should be moved to a Secret.
	// +optional
	MysqlUser *string `json:"mysqlUser,omitempty"`

	// TODO. This should be moved to a Secret.
	// +optional
	MysqlPassword *string `json:"mysqlPassword,omitempty"`

	// TODO. This should be moved to a Secret.
	// +optional
	MysqlRootPassword *string `json:"mysqlRootPassword,omitempty"`

	// TODO. This should be moved to a Secret.
	// +optional
	MysqlDatabase *string `json:"mysqlDatabase,omitempty"`

	// +optional
	ApicastManagementApi *string `json:"apicastManagementApi,omitempty"`

	// +optional
	ApicastOpenSSLVerify *bool `json:"apicastOpenSSLVerify,omitempty"`

	// +optional
	ApicastResponseCodes *bool `json:"apicastResponseCodes,omitempty"`

	// +optional
	ApicastRegistryURL *string `json:"apicastRegistryURL,omitempty"`

	WildcardDomain string `json:"wildcardDomain"`

	// +optional
	WildcardPolicy *string `json:"wildcardPolicy,omitempty"`

	Productized bool `json:"productized"`

	Evaluation bool `json:"evaluation"`

	S3Version bool `json:"s3version"`

	HAVersion bool `json:"haversion"`

	// S3 optional configuration TODO redo this
	// +optional
	AwsRegion         *string `json:"awsRegion, omitempty"`
	AwsBucket         *string `json:"awsBucket, omitempty"`
	FileUploadStorage *string `json:"fileUploadStorage, omitempty"`
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

func init() {
	SchemeBuilder.Register(&APIManager{}, &APIManagerList{})
}

// SetDefaults sets the default values for the APIManager spec and returns true if the spec was changed
func (apimanager *APIManager) SetDefaults() bool {
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

	if spec.RwxStorageClass == nil { // Needed??
		spec.RwxStorageClass = nil //in OpenShift template would be "null" in the parameter and nul in the field
		changed = true
	}

	if spec.AmpBackendImage == nil {
		defaultAmpBackendImage := "quay.io/3scale/apisonator:nightly"
		spec.AmpBackendImage = &defaultAmpBackendImage
		changed = true
	}

	if spec.AmpZyncImage == nil {
		defaultAmpZyncImage := "quay.io/3scale/zync:nightly"
		spec.AmpZyncImage = &defaultAmpZyncImage
		changed = true
	}

	if spec.AmpApicastImage == nil {
		defaultAmpApicastImage := "quay.io/3scale/apicast:nightly"
		spec.AmpApicastImage = &defaultAmpApicastImage
		changed = true
	}

	if spec.AmpRouterImage == nil {
		defaultAmpRouterImage := "quay.io/3scale/wildcard-router:nightly"
		spec.AmpRouterImage = &defaultAmpRouterImage
		changed = true
	}

	if spec.AmpSystemImage == nil {
		defaultAmpSystemImage := "quay.io/3scale/porta:nightly"
		spec.AmpSystemImage = &defaultAmpSystemImage
		changed = true
	}

	if spec.PostgreSQLImage == nil {
		defaultPostgreSQLImage := "registry.access.redhat.com/rhscl/postgresql-95-rhel7:9.5"
		spec.PostgreSQLImage = &defaultPostgreSQLImage
		changed = true
	}

	if spec.MysqlImage == nil {
		defaultMysqlImage := "registry.access.redhat.com/rhscl/mysql-57-rhel7:5.7"
		spec.MysqlImage = &defaultMysqlImage
		changed = true
	}

	if spec.MemcachedImage == nil {
		defaultMemcachedImage := "registry.access.redhat.com/3scale-amp20/memcached"
		spec.MemcachedImage = &defaultMemcachedImage
		changed = true
	}

	if spec.ImageStreamTagImportInsecure == nil {
		defaultImageStreamTagImportInsecure := false
		spec.ImageStreamTagImportInsecure = &defaultImageStreamTagImportInsecure
		changed = true
	}

	if spec.RedisImage == nil {
		defaultRedisImage := "registry.access.redhat.com/rhscl/redis-32-rhel7:3.2"
		spec.RedisImage = &defaultRedisImage
		changed = true
	}

	if spec.MysqlUser == nil {
		defaultMysqlUser := "mysql"
		spec.MysqlUser = &defaultMysqlUser
		changed = true
	}

	if spec.MysqlDatabase == nil {
		defaultMysqlDatabase := "system"
		spec.MysqlDatabase = &defaultMysqlDatabase
		changed = true
	}

	if spec.MysqlPassword == nil {
		defaultMysqlPassword := oprand.String(8)
		spec.MysqlPassword = &defaultMysqlPassword
		changed = true
	}

	if spec.MysqlRootPassword == nil {
		defaultMysqlRootPassword := oprand.String(8)
		spec.MysqlRootPassword = &defaultMysqlRootPassword
		changed = true
	}

	if spec.ApicastManagementApi == nil {
		defaultApicastManagementApi := "status"
		spec.ApicastManagementApi = &defaultApicastManagementApi
		changed = true
	}

	if spec.ApicastOpenSSLVerify == nil {
		defaultApicastOpenSSLVerify := false
		spec.ApicastOpenSSLVerify = &defaultApicastOpenSSLVerify
		changed = true
	}

	if spec.ApicastResponseCodes == nil {
		defaultApicastResponseCodes := true
		spec.ApicastResponseCodes = &defaultApicastResponseCodes
		changed = true
	}

	if spec.ApicastRegistryURL == nil {
		defaultApicastRegistryURL := "http://apicast-staging:8090/policies"
		spec.ApicastRegistryURL = &defaultApicastRegistryURL
		changed = true
	}

	if spec.WildcardPolicy == nil {
		defaultWildcardPolicy := "None" //TODO should be a set of predefined values (a custom type enum-like to be used)
		spec.WildcardPolicy = &defaultWildcardPolicy
		changed = true
	}

	return changed
}

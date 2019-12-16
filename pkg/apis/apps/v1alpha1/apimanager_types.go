package v1alpha1

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/version"
	"github.com/RHsyseng/operator-utils/pkg/olm"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

const (
	ThreescaleVersionAnnotation = "apps.3scale.net/apimanager-threescale-version"
	OperatorVersionAnnotation   = "apps.3scale.net/threescale-operator-version"
	Default3scaleAppLabel       = "3scale-api-management"
)

const (
	defaultTenantName                  = "3scale"
	defaultImageStreamImportInsecure   = false
	defaultResourceRequirementsEnabled = true
)

const (
	defaultApicastManagementAPI = "status"
	defaultApicastOpenSSLVerify = false
	defaultApicastResponseCodes = true
	defaultApicastRegistryURL   = "http://apicast-staging:8090/policies"
)

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
	// +optional
	PodDisruptionBudget *PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`
}

// APIManagerStatus defines the observed state of APIManager
// +k8s:openapi-gen=true
type APIManagerStatus struct {
	Conditions []APIManagerCondition `json:"conditions,omitempty" protobuf:"bytes,4,rep,name=conditions"`

	// APIManager Deployment Configs
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="Deployments"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:podStatuses"
	Deployments olm.DeploymentStatus `json:"deployments"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIManager is the Schema for the apimanagers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=apimanagers,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="APIManager"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="DeploymentConfig,apps.openshift.io/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="PersistentVolumeClaim,v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Service,v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Route,route.openshift.io/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="ImageStream,image.openshift.io/v1"
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
	// Wildcard domain as configured in the API Manager object
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Wildcard Domain"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:label"
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
	// +optional
	ProductionSpec *ApicastProductionSpec `json:"productionSpec,omitempty"`
	// +optional
	StagingSpec *ApicastStagingSpec `json:"stagingSpec,omitempty"`
}

type ApicastProductionSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
}

type ApicastStagingSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
}

type BackendSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`
	// +optional
	RedisImage *string `json:"redisImage,omitempty"`
	// +optional
	ListenerSpec *BackendListenerSpec `json:"listenerSpec,omitempty"`
	// +optional
	WorkerSpec *BackendWorkerSpec `json:"workerSpec,omitempty"`
	// +optional
	CronSpec *BackendCronSpec `json:"cronSpec,omitempty"`
}

type BackendListenerSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
}

type BackendWorkerSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
}

type BackendCronSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
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

	AppSpec     *SystemAppSpec     `json:"appSpec,omitempty"`
	SidekiqSpec *SystemSidekiqSpec `json:"sidekiqSpec,omitempty"`
}

type SystemAppSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
}

type SystemSidekiqSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
}

type SystemFileStorageSpec struct {
	// Union type. Only one of the fields can be set.
	// +optional
	PVC *SystemPVCSpec `json:"persistentVolumeClaim,omitempty"`
	// +optional
	// Deprecated
	DeprecatedS3 *DeprecatedSystemS3Spec `json:"amazonSimpleStorageService,omitempty"`
	// +optional
	S3 *SystemS3Spec `json:"simpleStorageService,omitempty"`
}

type SystemPVCSpec struct {
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
}

type DeprecatedSystemS3Spec struct {
	// Deprecated
	AWSBucket string `json:"awsBucket"`
	// Deprecated
	AWSRegion string `json:"awsRegion"`
	// Deprecated
	AWSCredentials v1.LocalObjectReference `json:"awsCredentialsSecret"`
}

type SystemS3Spec struct {
	ConfigurationSecretRef v1.LocalObjectReference `json:"configurationSecretRef"`
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

	// +optional
	AppSpec *ZyncAppSpec `json:"appSpec,omitempty"`

	// +optional
	QueSpec *ZyncQueSpec `json:"queSpec,omitempty"`
}

type ZyncAppSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
}

type ZyncQueSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
}

type HighAvailabilitySpec struct {
	Enabled bool `json:"enabled,omitempty"`
}

type PodDisruptionBudgetSpec struct {
	Enabled bool `json:"enabled,omitempty"`
}

func init() {
	SchemeBuilder.Register(&APIManager{}, &APIManagerList{})
}

// SetDefaults sets the default values for the APIManager spec and returns true if the spec was changed
func (apimanager *APIManager) SetDefaults() (bool, error) {
	var err error
	changed := false

	tmpChanged := apimanager.setAPIManagerAnnotationsDefaults()
	changed = changed || tmpChanged

	tmpChanged = apimanager.setAPIManagerCommonSpecDefaults()
	changed = changed || tmpChanged

	tmpChanged = apimanager.setBackendSpecDefaults()
	changed = changed || tmpChanged

	tmpChanged = apimanager.setApicastSpecDefaults()
	changed = changed || tmpChanged

	tmpChanged, err = apimanager.setSystemSpecDefaults()
	changed = changed || tmpChanged
	if err != nil {
		return changed, err
	}

	tmpChanged = apimanager.setZyncDefaults()
	changed = changed || tmpChanged

	return changed, err
}

func (apimanager *APIManager) setAPIManagerAnnotationsDefaults() bool {
	changed := false

	if apimanager.Annotations == nil {
		apimanager.Annotations = map[string]string{}
		changed = true
	}

	if _, ok := apimanager.Annotations[OperatorVersionAnnotation]; !ok {
		apimanager.Annotations[OperatorVersionAnnotation] = version.Version
		changed = true
	}

	if _, ok := apimanager.Annotations[ThreescaleVersionAnnotation]; !ok {
		apimanager.Annotations[ThreescaleVersionAnnotation] = product.ThreescaleRelease
		changed = true
	}

	return changed
}

func (apimanager *APIManager) setApicastSpecDefaults() bool {
	changed := false
	spec := &apimanager.Spec

	tmpDefaultApicastManagementAPI := defaultApicastManagementAPI
	tmpDefaultApicastOpenSSLVerify := defaultApicastOpenSSLVerify
	tmpDefaultApicastResponseCodes := defaultApicastResponseCodes
	tmpDefaultApicastRegistryURL := defaultApicastRegistryURL
	if spec.Apicast == nil {
		spec.Apicast = &ApicastSpec{}
		changed = true
	}

	if spec.Apicast.ApicastManagementAPI == nil {
		spec.Apicast.ApicastManagementAPI = &tmpDefaultApicastManagementAPI
		changed = true
	}
	if spec.Apicast.OpenSSLVerify == nil {
		spec.Apicast.OpenSSLVerify = &tmpDefaultApicastOpenSSLVerify
		changed = true
	}
	if spec.Apicast.IncludeResponseCodes == nil {
		spec.Apicast.IncludeResponseCodes = &tmpDefaultApicastResponseCodes
		changed = true
	}
	if spec.Apicast.RegistryURL == nil {
		spec.Apicast.RegistryURL = &tmpDefaultApicastRegistryURL
		changed = true
	}

	if spec.Apicast.StagingSpec == nil {
		spec.Apicast.StagingSpec = &ApicastStagingSpec{}
		changed = true
	}

	if spec.Apicast.ProductionSpec == nil {
		spec.Apicast.ProductionSpec = &ApicastProductionSpec{}
		changed = true
	}

	if spec.Apicast.StagingSpec.Replicas == nil {
		spec.Apicast.StagingSpec.Replicas = apimanager.defaultReplicas()
		changed = true
	}

	if spec.Apicast.ProductionSpec.Replicas == nil {
		spec.Apicast.ProductionSpec.Replicas = apimanager.defaultReplicas()
		changed = true
	}

	return changed
}

func (apimanager *APIManager) defaultReplicas() *int64 {
	var defaultReplicas int64 = 1
	return &defaultReplicas
}

func (apimanager *APIManager) setBackendSpecDefaults() bool {
	changed := false
	spec := &apimanager.Spec

	if spec.Backend == nil {
		spec.Backend = &BackendSpec{}
		changed = true
	}

	if spec.Backend.ListenerSpec == nil {
		spec.Backend.ListenerSpec = &BackendListenerSpec{}
		changed = true
	}

	if spec.Backend.CronSpec == nil {
		spec.Backend.CronSpec = &BackendCronSpec{}
		changed = true
	}

	if spec.Backend.WorkerSpec == nil {
		spec.Backend.WorkerSpec = &BackendWorkerSpec{}
		changed = true
	}

	if spec.Backend.ListenerSpec.Replicas == nil {
		spec.Backend.ListenerSpec.Replicas = apimanager.defaultReplicas()
		changed = true
	}

	if spec.Backend.CronSpec.Replicas == nil {
		spec.Backend.CronSpec.Replicas = apimanager.defaultReplicas()
		changed = true
	}

	if spec.Backend.WorkerSpec.Replicas == nil {
		spec.Backend.WorkerSpec.Replicas = apimanager.defaultReplicas()
		changed = true
	}

	return changed
}

func (apimanager *APIManager) setAPIManagerCommonSpecDefaults() bool {
	changed := false
	spec := &apimanager.Spec

	tmpDefaultAppLabel := Default3scaleAppLabel
	tmpDefaultTenantName := defaultTenantName
	tmpDefaultImageStreamTagImportInsecure := defaultImageStreamImportInsecure
	tmpDefaultResourceRequirementsEnabled := defaultResourceRequirementsEnabled

	if spec.AppLabel == nil {
		spec.AppLabel = &tmpDefaultAppLabel
		changed = true
	}

	if spec.TenantName == nil {
		spec.TenantName = &tmpDefaultTenantName
		changed = true
	}

	if spec.ImageStreamTagImportInsecure == nil {
		spec.ImageStreamTagImportInsecure = &tmpDefaultImageStreamTagImportInsecure
		changed = true
	}

	if spec.ResourceRequirementsEnabled == nil {
		spec.ResourceRequirementsEnabled = &tmpDefaultResourceRequirementsEnabled
		changed = true
	}

	// TODO do something with mandatory parameters?
	// TODO check that only compatible ProductRelease versions are compatible?

	return changed
}

func (apimanager *APIManager) setSystemSpecDefaults() (bool, error) {
	changed := false
	spec := &apimanager.Spec

	if spec.System == nil {
		spec.System = &SystemSpec{}
		changed = true
	}

	tmpChanged, err := apimanager.setSystemFileStorageSpecDefaults()
	changed = changed || tmpChanged
	if err != nil {
		return changed, err
	}

	tmpChanged, err = apimanager.setSystemDatabaseSpecDefaults()
	changed = changed || tmpChanged
	if err != nil {
		return changed, err
	}

	if spec.System.AppSpec == nil {
		spec.System.AppSpec = &SystemAppSpec{}
		changed = true
	}

	if spec.System.SidekiqSpec == nil {
		spec.System.SidekiqSpec = &SystemSidekiqSpec{}
		changed = true
	}

	if spec.System.AppSpec.Replicas == nil {
		spec.System.AppSpec.Replicas = apimanager.defaultReplicas()
		changed = true
	}

	if spec.System.SidekiqSpec.Replicas == nil {
		spec.System.SidekiqSpec.Replicas = apimanager.defaultReplicas()
		changed = true
	}

	return changed, nil
}

func (apimanager *APIManager) setSystemFileStorageSpecDefaults() (bool, error) {
	systemSpec := apimanager.Spec.System

	if systemSpec.FileStorageSpec != nil &&
		systemSpec.FileStorageSpec.PVC != nil &&
		systemSpec.FileStorageSpec.S3 != nil {
		return true, fmt.Errorf("Only one FileStorage can be chosen at the same time")
	}

	return false, nil
}

func (apimanager *APIManager) setSystemDatabaseSpecDefaults() (bool, error) {
	changed := false
	systemSpec := apimanager.Spec.System

	if apimanager.IsExternalDatabaseEnabled() {
		if systemSpec.DatabaseSpec != nil {
			systemSpec.DatabaseSpec = nil
			changed = true
		}
		return changed, nil
	}

	// databases managed internally
	if systemSpec.DatabaseSpec != nil &&
		systemSpec.DatabaseSpec.MySQL != nil &&
		systemSpec.DatabaseSpec.PostgreSQL != nil {
		return changed, fmt.Errorf("Only one System Database can be chosen at the same time")
	}

	return changed, nil
}

func (apimanager *APIManager) setZyncDefaults() bool {
	changed := false
	spec := &apimanager.Spec

	if spec.Zync == nil {
		spec.Zync = &ZyncSpec{}
		changed = true
	}

	if spec.Zync.AppSpec == nil {
		spec.Zync.AppSpec = &ZyncAppSpec{}
		changed = true
	}

	if spec.Zync.QueSpec == nil {
		spec.Zync.QueSpec = &ZyncQueSpec{}
	}

	if spec.Zync.AppSpec.Replicas == nil {
		spec.Zync.AppSpec.Replicas = apimanager.defaultReplicas()
		changed = true
	}

	if spec.Zync.QueSpec.Replicas == nil {
		spec.Zync.QueSpec.Replicas = apimanager.defaultReplicas()
		changed = true
	}

	return changed
}

func (apimanager *APIManager) IsExternalDatabaseEnabled() bool {
	return apimanager.Spec.HighAvailability != nil && apimanager.Spec.HighAvailability.Enabled
}

func (apimanager *APIManager) IsPDBEnabled() bool {
	return apimanager.Spec.PodDisruptionBudget != nil && apimanager.Spec.PodDisruptionBudget.Enabled
}

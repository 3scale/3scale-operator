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
	"fmt"
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/olm"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/version"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

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
type APIManagerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

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
	// +optional
	Monitoring *MonitoringSpec `json:"monitoring,omitempty"`
}

// APIManagerStatus defines the observed state of APIManager
type APIManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Current state of the APIManager resource.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`

	// APIManager Deployment Configs
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Deployments",xDescriptors="urn:alm:descriptor:com.tectonic.ui:podStatuses"
	Deployments olm.DeploymentStatus `json:"deployments"`
}

func (s *APIManagerStatus) Equals(other *APIManagerStatus, logger logr.Logger) bool {
	// Marshalling sorts by condition type
	currentMarshaledJSON, _ := s.Conditions.MarshalJSON()
	otherMarshaledJSON, _ := other.Conditions.MarshalJSON()
	if string(currentMarshaledJSON) != string(otherMarshaledJSON) {
		diff := cmp.Diff(string(currentMarshaledJSON), string(otherMarshaledJSON))
		logger.V(1).Info("Conditions not equal", "difference", diff)
		return false
	}

	// Deployments should already be sorted at this point so there's no need
	// to sort them and we can compare directly
	if !reflect.DeepEqual(s.Deployments, other.Deployments) {
		diff := cmp.Diff(s.Deployments, s.Deployments)
		logger.V(1).Info("Deployments not equal", "difference", diff)
		return false
	}

	return true
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// APIManager is the Schema for the apimanagers API
// +kubebuilder:resource:path=apimanagers,scope=Namespaced
// +operator-sdk:csv:customresourcedefinitions:displayName="APIManager"
// +operator-sdk:csv:customresourcedefinitions:resources={{"DeploymentConfig","apps.openshift.io/v1"}}
// +operator-sdk:csv:customresourcedefinitions:resources={{"PersistentVolumeClaim","v1"}}
// +operator-sdk:csv:customresourcedefinitions:resources={{"Service","v1"}}
// +operator-sdk:csv:customresourcedefinitions:resources={{"Route","route.openshift.io/v1"}}
// +operator-sdk:csv:customresourcedefinitions:resources={{"ImageStream","image.openshift.io/v1"}}
type APIManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIManagerSpec   `json:"spec,omitempty"`
	Status APIManagerStatus `json:"status,omitempty"`
}

const (
	APIManagerAvailableConditionType common.ConditionType = "Available"
)

type APIManagerCommonSpec struct {
	// Wildcard domain as configured in the API Manager object
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Wildcard Domain",xDescriptors="urn:alm:descriptor:com.tectonic.ui:label"
	WildcardDomain string `json:"wildcardDomain"`
	// +optional
	AppLabel *string `json:"appLabel,omitempty"`
	// +optional
	TenantName *string `json:"tenantName,omitempty"`
	// +optional
	ImageStreamTagImportInsecure *bool `json:"imageStreamTagImportInsecure,omitempty"`
	// +optional
	ResourceRequirementsEnabled *bool `json:"resourceRequirementsEnabled,omitempty"`
	// +optional
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// CustomPolicySpec contains or has reference to an APIcast custom policy
type CustomPolicySpec struct {
	// Name specifies the name of the custom policy
	Name string `json:"name"`
	// Version specifies the name of the custom policy
	Version string `json:"version"`
	// SecretRef specifies the secret holding the custom policy metadata and lua code
	SecretRef *v1.LocalObjectReference `json:"secretRef"`
}

func (c *CustomPolicySpec) VersionName() string {
	return fmt.Sprintf("%s%s", c.Name, c.Version)
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
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	// +kubebuilder:validation:Minimum=1
	Workers *int32 `json:"workers,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum=debug;info;notice;warn;error;crit;alert;emerg
	LogLevel *string `json:"logLevel,omitempty"` // APICAST_LOG_LEVEL
	// CustomPolicies specifies an array of defined custome policies to be loaded
	// +optional
	CustomPolicies []CustomPolicySpec `json:"customPolicies,omitempty"`
}

type ApicastStagingSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum=debug;info;notice;warn;error;crit;alert;emerg
	LogLevel *string `json:"logLevel,omitempty"` // APICAST_LOG_LEVEL
	// CustomPolicies specifies an array of defined custome policies to be loaded
	// +optional
	CustomPolicies []CustomPolicySpec `json:"customPolicies,omitempty"`
}

type BackendSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`
	// +optional
	RedisImage *string `json:"redisImage,omitempty"`
	// +optional
	RedisPersistentVolumeClaimSpec *BackendRedisPersistentVolumeClaimSpec `json:"redisPersistentVolumeClaim,omitempty"`
	// +optional
	RedisAffinity *v1.Affinity `json:"redisAffinity,omitempty"`
	// +optional
	RedisTolerations []v1.Toleration `json:"redisTolerations,omitempty"`
	// +optional
	RedisResources *v1.ResourceRequirements `json:"redisResources,omitempty"`
	// +optional
	ListenerSpec *BackendListenerSpec `json:"listenerSpec,omitempty"`
	// +optional
	WorkerSpec *BackendWorkerSpec `json:"workerSpec,omitempty"`
	// +optional
	CronSpec *BackendCronSpec `json:"cronSpec,omitempty"`
}

type BackendRedisPersistentVolumeClaimSpec struct {
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
}

type BackendListenerSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type BackendWorkerSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type BackendCronSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type SystemSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`

	// +optional
	MemcachedImage *string `json:"memcachedImage,omitempty"`

	// +optional
	MemcachedAffinity *v1.Affinity `json:"memcachedAffinity,omitempty"`
	// +optional
	MemcachedTolerations []v1.Toleration `json:"memcachedTolerations,omitempty"`
	// +optional
	MemcachedResources *v1.ResourceRequirements `json:"memcachedResources,omitempty"`

	// +optional
	RedisImage *string `json:"redisImage,omitempty"`

	// +optional
	RedisPersistentVolumeClaimSpec *SystemRedisPersistentVolumeClaimSpec `json:"redisPersistentVolumeClaim,omitempty"`
	// +optional
	RedisAffinity *v1.Affinity `json:"redisAffinity,omitempty"`
	// +optional
	RedisTolerations []v1.Toleration `json:"redisTolerations,omitempty"`
	// +optional
	RedisResources *v1.ResourceRequirements `json:"redisResources,omitempty"`

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

	// +optional
	AppSpec *SystemAppSpec `json:"appSpec,omitempty"`
	// +optional
	SidekiqSpec *SystemSidekiqSpec `json:"sidekiqSpec,omitempty"`
	// +optional
	SphinxSpec *SystemSphinxSpec `json:"sphinxSpec,omitempty"`
}

type SystemAppSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	MasterContainerResources *v1.ResourceRequirements `json:"masterContainerResources,omitempty"`
	// +optional
	ProviderContainerResources *v1.ResourceRequirements `json:"providerContainerResources,omitempty"`
	// +optional
	DeveloperContainerResources *v1.ResourceRequirements `json:"developerContainerResources,omitempty"`
}

type SystemSidekiqSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type SystemSphinxSpec struct {
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
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

type SystemRedisPersistentVolumeClaimSpec struct {
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
}

type SystemPVCSpec struct {
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
	// Resources represents the minimum resources the volume should have.
	// Ignored when VolumeName field is set
	// +optional
	Resources *PersistentVolumeClaimResources `json:"resources,omitempty"`
	// VolumeName is the binding reference to the PersistentVolume backing this claim.
	// +optional
	VolumeName *string `json:"volumeName,omitempty"`
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

	// +optional
	PersistentVolumeClaimSpec *SystemMySQLPVCSpec `json:"persistentVolumeClaim,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type SystemPostgreSQLSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`

	// +optional
	PersistentVolumeClaimSpec *SystemPostgreSQLPVCSpec `json:"persistentVolumeClaim,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type SystemMySQLPVCSpec struct {
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
	// Resources represents the minimum resources the volume should have.
	// Ignored when VolumeName field is set
	// +optional
	Resources *PersistentVolumeClaimResources `json:"resources,omitempty"`
	// VolumeName is the binding reference to the PersistentVolume backing this claim.
	// +optional
	VolumeName *string `json:"volumeName,omitempty"`
}

type SystemPostgreSQLPVCSpec struct {
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
	// Resources represents the minimum resources the volume should have.
	// Ignored when VolumeName field is set
	// +optional
	Resources *PersistentVolumeClaimResources `json:"resources,omitempty"`
	// VolumeName is the binding reference to the PersistentVolume backing this claim.
	// +optional
	VolumeName *string `json:"volumeName,omitempty"`
}

type ZyncSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`
	// +optional
	PostgreSQLImage *string `json:"postgreSQLImage,omitempty"`

	// +optional
	DatabaseAffinity *v1.Affinity `json:"databaseAffinity,omitempty"`
	// +optional
	DatabaseTolerations []v1.Toleration `json:"databaseTolerations,omitempty"`
	// +optional
	DatabaseResources *v1.ResourceRequirements `json:"databaseResources,omitempty"`

	// +optional
	AppSpec *ZyncAppSpec `json:"appSpec,omitempty"`

	// +optional
	QueSpec *ZyncQueSpec `json:"queSpec,omitempty"`
}

type ZyncAppSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type ZyncQueSpec struct {
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type HighAvailabilitySpec struct {
	Enabled bool `json:"enabled,omitempty"`
	// +optional
	ExternalZyncDatabaseEnabled *bool `json:"externalZyncDatabaseEnabled,omitempty"`
}

type PodDisruptionBudgetSpec struct {
	Enabled bool `json:"enabled,omitempty"`
}

type MonitoringSpec struct {
	Enabled bool `json:"enabled,omitempty"`
	// +optional
	EnablePrometheusRules *bool `json:"enablePrometheusRules,omitempty"`
}

// PersistentVolumeClaimResources defines the resources configuration
// of the backup data destination PersistentVolumeClaim
type PersistentVolumeClaimResources struct {
	// Storage Resource requests to be used on the PersistentVolumeClaim.
	// To learn more about resource requests see:
	// https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	Requests resource.Quantity `json:"requests"` // Should this be a string or a resoure.Quantity? it seems it is serialized as a string
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

	if spec.System.SphinxSpec == nil {
		spec.System.SphinxSpec = &SystemSphinxSpec{}
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
		changed = true
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

func (apimanager *APIManager) IsZyncExternalDatabaseEnabled() bool {
	return apimanager.IsExternalDatabaseEnabled() &&
		apimanager.Spec.HighAvailability.ExternalZyncDatabaseEnabled != nil &&
		*apimanager.Spec.HighAvailability.ExternalZyncDatabaseEnabled
}
func (apimanager *APIManager) IsPDBEnabled() bool {
	return apimanager.Spec.PodDisruptionBudget != nil && apimanager.Spec.PodDisruptionBudget.Enabled
}

func (apimanager *APIManager) IsSystemPostgreSQLEnabled() bool {
	return !apimanager.IsExternalDatabaseEnabled() &&
		apimanager.Spec.System.DatabaseSpec != nil &&
		apimanager.Spec.System.DatabaseSpec.PostgreSQL != nil
}

func (apimanager *APIManager) IsSystemMysqlEnabled() bool {
	return !apimanager.IsExternalDatabaseEnabled() &&
		apimanager.Spec.System.DatabaseSpec != nil &&
		apimanager.Spec.System.DatabaseSpec.MySQL != nil
}

func (apimanager *APIManager) IsMonitoringEnabled() bool {
	return apimanager.Spec.Monitoring != nil && apimanager.Spec.Monitoring.Enabled
}

func (apimanager *APIManager) IsPrometheusRulesEnabled() bool {
	return (apimanager.IsMonitoringEnabled() &&
		(apimanager.Spec.Monitoring.EnablePrometheusRules == nil || *apimanager.Spec.Monitoring.EnablePrometheusRules))
}

func (apimanager *APIManager) Validate() field.ErrorList {
	fieldErrors := field.ErrorList{}

	specFldPath := field.NewPath("spec")

	if apimanager.Spec.Apicast != nil {
		apicastFldPath := specFldPath.Child("apicast")

		if apimanager.Spec.Apicast.ProductionSpec != nil {
			prodSpecFldPath := apicastFldPath.Child("productionSpec")
			customPoliciesFldPath := prodSpecFldPath.Child("customPolicies")
			duplicateMap := make(map[string]int)
			for idx, customPolicySpec := range apimanager.Spec.Apicast.ProductionSpec.CustomPolicies {
				customPoliciesIdxFldPath := customPoliciesFldPath.Index(idx)

				// check custom policy secret is set
				if customPolicySpec.SecretRef == nil {
					fieldErrors = append(fieldErrors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret is mandatory"))
				} else if customPolicySpec.SecretRef.Name == "" {
					fieldErrors = append(fieldErrors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret name is empty"))
				}

				// check duplicated custom policy version name
				if _, ok := duplicateMap[customPolicySpec.VersionName()]; ok {
					fieldErrors = append(fieldErrors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret name version tuple is duplicated"))
					break
				}
				duplicateMap[customPolicySpec.VersionName()] = 0
			}
		}

		if apimanager.Spec.Apicast.StagingSpec != nil {
			stagingSpecFldPath := apicastFldPath.Child("stagingSpec")
			customPoliciesFldPath := stagingSpecFldPath.Child("customPolicies")
			duplicateMap := make(map[string]int)
			for idx, customPolicySpec := range apimanager.Spec.Apicast.StagingSpec.CustomPolicies {
				// TODO(eastizle): DRY!!
				customPoliciesIdxFldPath := customPoliciesFldPath.Index(idx)

				// check custom policy secret is set
				if customPolicySpec.SecretRef == nil {
					fieldErrors = append(fieldErrors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret is mandatory"))
				} else if customPolicySpec.SecretRef.Name == "" {
					fieldErrors = append(fieldErrors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret name is empty"))
				}

				// check duplicated custom policy version name
				if _, ok := duplicateMap[customPolicySpec.VersionName()]; ok {
					fieldErrors = append(fieldErrors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret name version tuple is duplicated"))
					break
				}
				duplicateMap[customPolicySpec.VersionName()] = 0
			}
		}

	}

	return fieldErrors
}

// +kubebuilder:object:root=true

// APIManagerList contains a list of APIManager
type APIManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIManager{}, &APIManagerList{})
}

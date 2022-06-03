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

	"github.com/3scale/3scale-operator/pkg/common"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

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

// CustomEnvironmentSpec contains or has reference to an APIcast custom environment
type CustomEnvironmentSpec struct {
	SecretRef *v1.LocalObjectReference `json:"secretRef"`
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
	// OpenTracing contains the OpenTracing integration configuration
	// with APIcast in the production environment.
	// +optional
	OpenTracing *APIcastOpenTracingSpec `json:"openTracing,omitempty"`
	// CustomEnvironments specifies an array of defined custom environments to be loaded
	// +optional
	CustomEnvironments []CustomEnvironmentSpec `json:"customEnvironments,omitempty"` // APICAST_ENVIRONMENT
	// HttpsPort controls on which port APIcast should start listening for HTTPS connections.
	// If this clashes with HTTP port it will be used only for HTTPS.
	// Enable TLS at APIcast pod level setting either `httpsPort` or `httpsCertificateSecretRef` fields or both.
	// +optional
	HTTPSPort *int32 `json:"httpsPort,omitempty"` // APICAST_HTTPS_PORT
	// HTTPSVerifyDepth defines the maximum length of the client certificate chain.
	// +kubebuilder:validation:Minimum=0
	// +optional
	HTTPSVerifyDepth *int64 `json:"httpsVerifyDepth,omitempty"` // APICAST_HTTPS_VERIFY_DEPTH
	// HTTPSCertificateSecretRef references secret containing the X.509 certificate in the PEM format and the X.509 certificate secret key.
	// Enable TLS at APIcast pod level setting either `httpsPort` or `httpsCertificateSecretRef` fields or both.
	// +optional
	HTTPSCertificateSecretRef *v1.LocalObjectReference `json:"httpsCertificateSecretRef,omitempty"`
	// AllProxy specifies a HTTP(S) proxy to be used for connecting to services if
	// a protocol-specific proxy is not specified. Authentication is not supported.
	// Format is <scheme>://<host>:<port>
	// +optional
	AllProxy *string `json:"allProxy,omitempty"` // ALL_PROXY
	// HTTPProxy specifies a HTTP(S) Proxy to be used for connecting to HTTP services.
	// Authentication is not supported. Format is <scheme>://<host>:<port>
	// +optional
	HTTPProxy *string `json:"httpProxy,omitempty"` // HTTP_PROXY
	// HTTPSProxy specifies a HTTP(S) Proxy to be used for connecting to HTTPS services.
	// Authentication is not supported. Format is <scheme>://<host>:<port>
	// +optional
	HTTPSProxy *string `json:"httpsProxy,omitempty"` // HTTPS_PROXY
	// NoProxy specifies a comma-separated list of hostnames and domain
	// names for which the requests should not be proxied. Setting to a single
	// * character, which matches all hosts, effectively disables the proxy.
	// +optional
	NoProxy *string `json:"noProxy,omitempty"` // NO_PROXY
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
	// OpenTracing contains the OpenTracing integration configuration
	// with APIcast in the staging environment.
	// +optional
	OpenTracing *APIcastOpenTracingSpec `json:"openTracing,omitempty"`
	// CustomEnvironments specifies an array of defined custom environments to be loaded
	// +optional
	CustomEnvironments []CustomEnvironmentSpec `json:"customEnvironments,omitempty"` // APICAST_ENVIRONMENT
	// HttpsPort controls on which port APIcast should start listening for HTTPS connections.
	// If this clashes with HTTP port it will be used only for HTTPS.
	// Enable TLS at APIcast pod level setting either `httpsPort` or `httpsCertificateSecretRef` fields or both.
	// +optional
	HTTPSPort *int32 `json:"httpsPort,omitempty"` // APICAST_HTTPS_PORT
	// HTTPSVerifyDepth defines the maximum length of the client certificate chain.
	// +kubebuilder:validation:Minimum=0
	// +optional
	HTTPSVerifyDepth *int64 `json:"httpsVerifyDepth,omitempty"` // APICAST_HTTPS_VERIFY_DEPTH
	// HTTPSCertificateSecretRef references secret containing the X.509 certificate in the PEM format and the X.509 certificate secret key.
	// Enable TLS at APIcast pod level setting either `httpsPort` or `httpsCertificateSecretRef` fields or both.
	// +optional
	HTTPSCertificateSecretRef *v1.LocalObjectReference `json:"httpsCertificateSecretRef,omitempty"`
	// AllProxy specifies a HTTP(S) proxy to be used for connecting to services if
	// a protocol-specific proxy is not specified. Authentication is not supported.
	// Format is <scheme>://<host>:<port>
	// +optional
	AllProxy *string `json:"allProxy,omitempty"` // ALL_PROXY
	// HTTPProxy specifies a HTTP(S) Proxy to be used for connecting to HTTP services.
	// Authentication is not supported. Format is <scheme>://<host>:<port>
	// +optional
	HTTPProxy *string `json:"httpProxy,omitempty"` // HTTP_PROXY
	// HTTPSProxy specifies a HTTP(S) Proxy to be used for connecting to HTTPS services.
	// Authentication is not supported. Format is <scheme>://<host>:<port>
	// +optional
	HTTPSProxy *string `json:"httpsProxy,omitempty"` // HTTPS_PROXY
	// NoProxy specifies a comma-separated list of hostnames and domain
	// names for which the requests should not be proxied. Setting to a single
	// * character, which matches all hosts, effectively disables the proxy.
	// +optional
	NoProxy *string `json:"noProxy,omitempty"` // NO_PROXY
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

type ExternalComponentsSpec struct {
	// +optional
	System *ExternalSystemComponents `json:"system,omitempty"`
	// +optional
	Backend *ExternalBackendComponents `json:"backend,omitempty"`
	// +optional
	Zync *ExternalZyncComponents `json:"zync,omitempty"`
}

type ExternalSystemComponents struct {
	// +optional
	Redis *bool `json:"redis,omitempty"`
	// +optional
	Database *bool `json:"database,omitempty"`
}

type ExternalBackendComponents struct {
	// +optional
	Redis *bool `json:"redis,omitempty"`
}

type ExternalZyncComponents struct {
	// +optional
	Database *bool `json:"database,omitempty"`
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

type APIcastOpenTracingSpec struct {
	// Enabled controls whether OpenTracing integration with APIcast is enabled.
	// By default it is not enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	// TracingLibrary controls which OpenTracing library is loaded. At the moment
	// the only supported tracer is `jaeger`. If not set, `jaeger` will be used.
	TracingLibrary *string `json:"tracingLibrary,omitempty"`
	// TracingConfigSecretRef contains a secret reference the OpenTracing configuration.
	// Each supported tracing library provides a default configuration file
	// that is used if TracingConfig is not specified.
	// +optional
	TracingConfigSecretRef *v1.LocalObjectReference `json:"tracingConfigSecretRef,omitempty"`
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

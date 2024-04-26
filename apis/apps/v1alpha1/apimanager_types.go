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

	"github.com/3scale/3scale-operator/apis/apps"
	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	"github.com/3scale/3scale-operator/version"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	ThreescaleVersionAnnotation     = "apps.3scale.net/apimanager-threescale-version"
	OperatorVersionAnnotation       = "apps.3scale.net/threescale-operator-version"
	Default3scaleAppLabel           = "3scale-api-management"
	ThreescaleRequirementsConfirmed = "apps.3scale.net/apimanager-confirmed-requirements-version"
)

const (
	defaultTenantName                  = "3scale"
	defaultResourceRequirementsEnabled = true
)

const (
	defaultApicastManagementAPI = "status"
	defaultApicastOpenSSLVerify = false
	defaultApicastResponseCodes = true
	defaultApicastRegistryURL   = "http://apicast-staging:8090/policies"
)

const (
	DefaultHTTPPort  int32 = 8080
	DefaultHTTPSPort int32 = 8443
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
	ExternalComponents *ExternalComponentsSpec `json:"externalComponents,omitempty"`

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

	// APIManager Deployments
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
// +operator-sdk:csv:customresourcedefinitions:resources={{"Deployment","apps/v1"}}
// +operator-sdk:csv:customresourcedefinitions:resources={{"ConfigMap","v1"}}
// +operator-sdk:csv:customresourcedefinitions:resources={{"PersistentVolumeClaim","v1"}}
// +operator-sdk:csv:customresourcedefinitions:resources={{"Service","v1"}}
// +operator-sdk:csv:customresourcedefinitions:resources={{"Route","route.openshift.io/v1"}}
type APIManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// XPreserveUnknownFields Does not work at type level. Tested on OCP 4.8

	// +kubebuilder:validation:XPreserveUnknownFields
	Spec APIManagerSpec `json:"spec,omitempty"`

	// +kubebuilder:validation:XPreserveUnknownFields
	Status APIManagerStatus `json:"status,omitempty"`
}

const (
	APIManagerAvailableConditionType  common.ConditionType = "Available"
	APIManagerWarningConditionType    common.ConditionType = "Warning"
	APIManagerPreflightsConditionType common.ConditionType = "Preflights"
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
	// Hpa specifies an array of defined HPA values
	//+optional
	Hpa bool `json:"hpa,omitempty"`
	// OpenTracing contains the OpenTracing integration configuration
	// with APIcast in the production environment.
	// Deprecated
	// +optional
	OpenTracing *APIcastOpenTracingSpec `json:"openTracing,omitempty"`
	// OpenTelemetry contains the gateway instrumentation configuration
	// with APIcast.
	// +optional
	OpenTelemetry *OpenTelemetrySpec `json:"openTelemetry,omitempty"`
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
	// ServiceCacheSize specifies the number of services that APICast can store in the internal cache
	// +optional
	ServiceCacheSize *int32 `json:"serviceCacheSize,omitempty"` // APICAST_SERVICE_CACHE_SIZE
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
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
	// Deprecated
	// +optional
	OpenTracing *APIcastOpenTracingSpec `json:"openTracing,omitempty"`
	// OpenTelemetry contains the gateway instrumentation configuration
	// with APIcast.
	// +optional
	OpenTelemetry *OpenTelemetrySpec `json:"openTelemetry,omitempty"`
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
	// ServiceCacheSize specifies the number of services that APICast can store in the internal cache
	// +optional
	ServiceCacheSize *int32 `json:"serviceCacheSize,omitempty"` // APICAST_SERVICE_CACHE_SIZE
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
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
	RedisPriorityClassName *string `json:"redisPriorityClassName,omitempty"`
	// +optional
	RedisTopologySpreadConstraints []v1.TopologySpreadConstraint `json:"redisTopologySpreadConstraints,omitempty"`
	// +optional
	RedisLabels map[string]string `json:"redisLabels,omitempty"`
	// +optional
	RedisAnnotations map[string]string `json:"redisAnnotations,omitempty"`
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
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// Hpa specifies an array of defined HPA values
	//+optional
	Hpa bool `json:"hpa,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
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
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// Hpa specifies an array of defined HPA values
	//+optional
	Hpa bool `json:"hpa,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
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
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
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
	MemcachedPriorityClassName *string `json:"memcachedPriorityClassName,omitempty"`
	// +optional
	MemcachedTopologySpreadConstraints []v1.TopologySpreadConstraint `json:"memcachedTopologySpreadConstraints,omitempty"`
	// +optional
	MemcachedLabels map[string]string `json:"memcachedLabels,omitempty"`
	// +optional
	MemcachedAnnotations map[string]string `json:"memcachedAnnotations,omitempty"`

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
	// +optional
	RedisPriorityClassName *string `json:"redisPriorityClassName,omitempty"`
	// +optional
	RedisTopologySpreadConstraints []v1.TopologySpreadConstraint `json:"redisTopologySpreadConstraints,omitempty"`
	// +optional
	RedisLabels map[string]string `json:"redisLabels,omitempty"`
	// +optional
	RedisAnnotations map[string]string `json:"redisAnnotations,omitempty"`

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
	// Deprecated
	SphinxSpec *SystemSphinxSpec `json:"sphinxSpec,omitempty"`

	// +optional
	SearchdSpec *SystemSearchdSpec `json:"searchdSpec,omitempty"`
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
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
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
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

type SystemSearchdSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	PVC *PVCGenericSpec `json:"persistentVolumeClaim,omitempty"`
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

type SystemSphinxSpec struct {
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
}

type SystemFileStorageSpec struct {
	// Union type. Only one of the fields can be set.
	// +optional
	PVC *PVCGenericSpec `json:"persistentVolumeClaim,omitempty"`
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

type PVCGenericSpec struct {
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
	// STS authentication spec
	// +optional
	STS *STSSpec `json:"sts,omitempty"`
}

type STSSpec struct {
	// Enable Secure Token Service for  short-term, limited-privilege security credentials
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// The ID the token is intended for
	// +optional
	Audience *string `json:"audience,omitempty"`
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
	PersistentVolumeClaimSpec *PVCGenericSpec `json:"persistentVolumeClaim,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

type SystemPostgreSQLSpec struct {
	// +optional
	Image *string `json:"image,omitempty"`

	// +optional
	PersistentVolumeClaimSpec *PVCGenericSpec `json:"persistentVolumeClaim,omitempty"`
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
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

	// +optional
	DatabasePriorityClassName *string `json:"databasePriorityClassName,omitempty"`
	// +optional
	DatabaseTopologySpreadConstraints []v1.TopologySpreadConstraint `json:"databaseTopologySpreadConstraints,omitempty"`
	// +optional
	DatabaseLabels map[string]string `json:"databaseLabels,omitempty"`
	// +optional
	DatabaseAnnotations map[string]string `json:"databaseAnnotations,omitempty"`
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
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
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
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// +optional
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
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

type OpenTelemetrySpec struct {
	// Enabled controls whether OpenTelemetry integration with APIcast is enabled.
	// By default it is not enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// TracingConfigSecretRef contains a Secret reference the Opentelemetry configuration.
	// The configuration file specification is defined in the Nginx instrumentation library repo
	// https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx
	// +optional
	TracingConfigSecretRef *v1.LocalObjectReference `json:"tracingConfigSecretRef,omitempty"`

	// TracingConfigSecretKey contains the key of the secret to select the configuration from.
	// if unspecified, the first secret key in lexicographical order will be selected.
	// +optional
	TracingConfigSecretKey *string `json:"tracingConfigSecretKey,omitempty"`
}

func (a *APIManager) OpenTelemetryEnabledForStaging() bool {
	return a.Spec.Apicast != nil && a.Spec.Apicast.StagingSpec != nil && a.Spec.Apicast.StagingSpec.OpenTelemetry != nil && a.Spec.Apicast.StagingSpec.OpenTelemetry.Enabled != nil && *a.Spec.Apicast.StagingSpec.OpenTelemetry.Enabled
}

func (a *APIManager) OpenTelemetryEnabledForProduction() bool {
	return a.Spec.Apicast != nil && a.Spec.Apicast.ProductionSpec != nil && a.Spec.Apicast.ProductionSpec.OpenTelemetry != nil && a.Spec.Apicast.ProductionSpec.OpenTelemetry.Enabled != nil && *a.Spec.Apicast.ProductionSpec.OpenTelemetry.Enabled
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

func (apimanager *APIManager) IsInFreshInstallationScenario() bool {
	threescaleAnnotationFound := true

	if _, ok := apimanager.Annotations[ThreescaleVersionAnnotation]; ok {
		threescaleAnnotationFound = false
	}

	return threescaleAnnotationFound
}

func (apimanager *APIManager) IsHealthyUpgradeScenario() bool {
	threescaleAnnotationFound := false

	// is the version annotation present
	if _, ok := apimanager.Annotations[ThreescaleVersionAnnotation]; ok {
		threescaleAnnotationFound = true
	}

	// is the condition set to "available" = "true"
	conditionAvailable := apimanager.IsExistingInstallationHealthy()

	if threescaleAnnotationFound && conditionAvailable {
		return true
	}

	return false
}

func (apimanager *APIManager) IsExistingInstallationHealthy() bool {
	// Fetch ready condition
	availableCondition := apimanager.Status.Conditions.GetCondition(APIManagerAvailableConditionType)

	// return true if the ready condition is set to true
	if availableCondition != nil && availableCondition.Status == v1.ConditionTrue {
		return true
	}

	return false
}

func (apimanager *APIManager) RequirementsConfirmed(requirementsConfigMapResourceVersion string) bool {
	if val, ok := apimanager.Annotations[ThreescaleRequirementsConfirmed]; ok && val == requirementsConfigMapResourceVersion {
		return true
	}

	return false
}

func (apimanager *APIManager) RetrieveRHTVersion() string {
	if val, ok := apimanager.Annotations[ThreescaleVersionAnnotation]; ok && val != "" {
		return val
	}

	return ""
}

func (apimanager *APIManager) IsMultiMinorHopDetected() (bool, error) {
	var currentlyInstalledVersion string
	var multiMinorHopDetected bool

	if val, ok := apimanager.Annotations[ThreescaleVersionAnnotation]; ok && val != "" {
		currentlyInstalledVersion = val
	}

	if currentlyInstalledVersion != "" {
		multiMinorHop, err := common.CompareMinorVersions(currentlyInstalledVersion, version.ThreescaleVersionMajorMinor())
		if err != nil {
			return true, err
		}
		multiMinorHopDetected = multiMinorHop
	}

	return multiMinorHopDetected, nil
}

func (apimanager *APIManager) setAPIManagerAnnotationsDefaults() bool {
	changed := false

	if apimanager.Annotations == nil {
		apimanager.Annotations = map[string]string{}
		changed = true
	}

	if v, ok := apimanager.Annotations[OperatorVersionAnnotation]; !ok || v != version.Version {
		apimanager.Annotations[OperatorVersionAnnotation] = version.Version
		changed = true
	}

	if v, ok := apimanager.Annotations[ThreescaleVersionAnnotation]; !ok || v != version.ThreescaleVersionMajorMinorPatch() {
		apimanager.Annotations[ThreescaleVersionAnnotation] = version.ThreescaleVersionMajorMinorPatch()
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

	return changed
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

	return changed
}

func (apimanager *APIManager) setAPIManagerCommonSpecDefaults() bool {
	changed := false
	spec := &apimanager.Spec

	tmpDefaultAppLabel := Default3scaleAppLabel
	tmpDefaultTenantName := defaultTenantName
	tmpDefaultResourceRequirementsEnabled := defaultResourceRequirementsEnabled

	if spec.AppLabel == nil {
		spec.AppLabel = &tmpDefaultAppLabel
		changed = true
	}

	if spec.TenantName == nil {
		spec.TenantName = &tmpDefaultTenantName
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

	if spec.System.SearchdSpec == nil {
		spec.System.SearchdSpec = &SystemSearchdSpec{}
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

	if apimanager.IsExternal(SystemDatabase) {
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

	return changed
}

func (apimanager *APIManager) UpdateExternalComponentsFromHighAvailability() bool {
	// The external components field is already populated. Nothing to do
	if apimanager.Spec.ExternalComponents != nil {
		return false
	}

	updated := false
	// When the info comes from the deprecated .spec.highAvailability field
	e := mapHighAvailabilityToExternalComponents(apimanager)

	if e != nil {
		apimanager.Spec.ExternalComponents = e
		updated = true
	}

	// Remove the deprecated field
	if apimanager.Spec.HighAvailability != nil {
		apimanager.Spec.HighAvailability = nil
		updated = true
	}

	return updated
}

func (apimanager *APIManager) IsExternal(selector func(*ExternalComponentsSpec) bool) bool {
	// ExternalComponents has precedence over HighAvailability
	if apimanager.Spec.ExternalComponents != nil {
		return selector(apimanager.Spec.ExternalComponents)
	}

	// When the info comes from the deprecated .spec.highAvailability field
	e := mapHighAvailabilityToExternalComponents(apimanager)

	if e != nil {
		return selector(e)
	}

	return false
}

func mapHighAvailabilityToExternalComponents(apiManager *APIManager) *ExternalComponentsSpec {
	if apiManager.Spec.HighAvailability == nil {
		return nil
	}

	if !apiManager.Spec.HighAvailability.Enabled {
		return nil
	}

	trueVal := true

	e := &ExternalComponentsSpec{
		System:  &ExternalSystemComponents{Redis: &trueVal, Database: &trueVal},
		Backend: &ExternalBackendComponents{Redis: &trueVal},
	}

	if apiManager.Spec.HighAvailability.ExternalZyncDatabaseEnabled != nil &&
		*apiManager.Spec.HighAvailability.ExternalZyncDatabaseEnabled {
		e.Zync = &ExternalZyncComponents{Database: &trueVal}
	}

	return e
}

func AllComponentsExternal() *ExternalComponentsSpec {
	trueVal := true
	return &ExternalComponentsSpec{
		System:  &ExternalSystemComponents{Redis: &trueVal, Database: &trueVal},
		Backend: &ExternalBackendComponents{Redis: &trueVal},
		Zync:    &ExternalZyncComponents{Database: &trueVal},
	}
}

func SystemDatabase(e *ExternalComponentsSpec) bool {
	return e != nil && e.System != nil && e.System.Database != nil && *e.System.Database
}

func SystemRedis(e *ExternalComponentsSpec) bool {
	return e != nil && e.System != nil && e.System.Redis != nil && *e.System.Redis
}

func BackendRedis(e *ExternalComponentsSpec) bool {
	return e != nil && e.Backend != nil && e.Backend.Redis != nil && *e.Backend.Redis
}

func ZyncDatabase(e *ExternalComponentsSpec) bool {
	return e != nil && e.Zync != nil && e.Zync.Database != nil && *e.Zync.Database
}

func (apimanager *APIManager) IsPDBEnabled() bool {
	return apimanager.Spec.PodDisruptionBudget != nil && apimanager.Spec.PodDisruptionBudget.Enabled
}

func (apimanager *APIManager) IsSystemPostgreSQLEnabled() bool {
	return !apimanager.IsExternal(SystemDatabase) &&
		apimanager.Spec.System.DatabaseSpec != nil &&
		apimanager.Spec.System.DatabaseSpec.PostgreSQL != nil
}

func (apimanager *APIManager) IsSystemMysqlEnabled() bool {
	return !apimanager.IsExternal(SystemDatabase) && !apimanager.IsSystemPostgreSQLEnabled()
}

func (apimanager *APIManager) IsMonitoringEnabled() bool {
	return apimanager.Spec.Monitoring != nil && apimanager.Spec.Monitoring.Enabled
}

func (apimanager *APIManager) IsPrometheusRulesEnabled() bool {
	return (apimanager.IsMonitoringEnabled() &&
		(apimanager.Spec.Monitoring.EnablePrometheusRules == nil || *apimanager.Spec.Monitoring.EnablePrometheusRules))
}

func (apimanager *APIManager) IsAPIcastProductionOpenTracingEnabled() bool {
	return apimanager.Spec.Apicast != nil && apimanager.Spec.Apicast.ProductionSpec != nil &&
		apimanager.Spec.Apicast.ProductionSpec.OpenTracing != nil &&
		apimanager.Spec.Apicast.ProductionSpec.OpenTracing.Enabled != nil &&
		*apimanager.Spec.Apicast.ProductionSpec.OpenTracing.Enabled
}

func (apimanager *APIManager) IsAPIcastStagingOpenTracingEnabled() bool {
	return apimanager.Spec.Apicast != nil && apimanager.Spec.Apicast.StagingSpec != nil &&
		apimanager.Spec.Apicast.StagingSpec.OpenTracing != nil &&
		apimanager.Spec.Apicast.StagingSpec.OpenTracing.Enabled != nil &&
		*apimanager.Spec.Apicast.StagingSpec.OpenTracing.Enabled
}

func (apimanager *APIManager) IsS3Enabled() bool {
	return apimanager.Spec.System.FileStorageSpec != nil &&
		apimanager.Spec.System.FileStorageSpec.S3 != nil
}

func (apimanager *APIManager) IsS3STSEnabled() bool {
	return apimanager.IsS3Enabled() &&
		apimanager.Spec.System.FileStorageSpec.S3.STS != nil &&
		// Defined here the default value when Enabled is not specified.
		// when Enabled is not set in the CR, IsS3STSEnabled() returns false
		apimanager.Spec.System.FileStorageSpec.S3.STS.Enabled != nil &&
		*apimanager.Spec.System.FileStorageSpec.S3.STS.Enabled
}

func (apimanager *APIManager) IsS3IAMEnabled() bool {
	return apimanager.IsS3Enabled() && !apimanager.IsS3STSEnabled()
}

func (apimanager *APIManager) Validate() field.ErrorList {
	fieldErrors := field.ErrorList{}

	specFldPath := field.NewPath("spec")

	if apimanager.Spec.Apicast != nil {
		apicastFldPath := specFldPath.Child("apicast")

		if apimanager.Spec.Apicast.ProductionSpec != nil {
			prodSpecFldPath := apicastFldPath.Child("productionSpec")

			customPoliciesFldPath := prodSpecFldPath.Child("customPolicies")
			duplicatePolicyMap := make(map[string]int)
			for idx, customPolicySpec := range apimanager.Spec.Apicast.ProductionSpec.CustomPolicies {
				customPoliciesIdxFldPath := customPoliciesFldPath.Index(idx)

				// check custom policy secret is set
				if customPolicySpec.SecretRef == nil {
					fieldErrors = append(fieldErrors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret is mandatory"))
				} else if customPolicySpec.SecretRef.Name == "" {
					fieldErrors = append(fieldErrors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret name is empty"))
				}

				// check duplicated custom policy version name
				if _, ok := duplicatePolicyMap[customPolicySpec.VersionName()]; ok {
					fieldErrors = append(fieldErrors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret name version tuple is duplicated"))
					break
				}
				duplicatePolicyMap[customPolicySpec.VersionName()] = 0
			}

			if apimanager.OpenTelemetryEnabledForProduction() {
				openTelemetrySpec := apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry
				if openTelemetrySpec.TracingConfigSecretRef != nil {
					if openTelemetrySpec.TracingConfigSecretRef.Name == "" {
						apicastOpenTelemtryTracingFldPath := prodSpecFldPath.Child("openTelemetry")
						customTracingConfigFldPath := apicastOpenTelemtryTracingFldPath.Child("tracingConfigSecretRef")
						fieldErrors = append(fieldErrors, field.Invalid(customTracingConfigFldPath, openTelemetrySpec, "custom tracing library secret name is empty"))
					}
				}
			}

			if apimanager.IsAPIcastProductionOpenTracingEnabled() {
				openTracingConfigSpec := apimanager.Spec.Apicast.ProductionSpec.OpenTracing
				if openTracingConfigSpec.TracingConfigSecretRef != nil {
					if openTracingConfigSpec.TracingConfigSecretRef.Name == "" {
						apicastProductioOpenTracingFldPath := prodSpecFldPath.Child("openTracing")
						customTracingConfigFldPath := apicastProductioOpenTracingFldPath.Child("tracingConfigSecretRef")
						fieldErrors = append(fieldErrors, field.Invalid(customTracingConfigFldPath, apimanager.Spec.Apicast.ProductionSpec.OpenTracing, "custom tracing library secret name is empty"))
					}
				}

				// For now only "jaeger" is accepted" as the tracing library
				if openTracingConfigSpec.TracingLibrary != nil && *openTracingConfigSpec.TracingLibrary != apps.APIcastDefaultTracingLibrary {
					tracingLibraryFldPath := field.NewPath("spec").
						Child("apicast").
						Child("productionSpec").
						Child("openTracing").
						Child("tracingLibrary")
					fieldErrors = append(fieldErrors, field.Invalid(tracingLibraryFldPath, openTracingConfigSpec, "invalid tracing library specified"))
				}
			}

			customEnvsFldPath := prodSpecFldPath.Child("customEnvironments")
			duplicateEnvMap := make(map[string]int)
			// check custom environment secret is set
			for idx, customEnvSpec := range apimanager.Spec.Apicast.ProductionSpec.CustomEnvironments {
				customEnvsIdxFldPath := customEnvsFldPath.Index(idx)

				if customEnvSpec.SecretRef == nil {
					fieldErrors = append(fieldErrors, field.Invalid(customEnvsIdxFldPath, customEnvSpec, "custom environment secret is mandatory"))
				} else if customEnvSpec.SecretRef.Name == "" {
					fieldErrors = append(fieldErrors, field.Invalid(customEnvsIdxFldPath, customEnvSpec, "custom environment secret name is empty"))
				} else {
					// check duplicated custom env secret
					if _, ok := duplicateEnvMap[customEnvSpec.SecretRef.Name]; ok {
						fieldErrors = append(fieldErrors, field.Invalid(customEnvsIdxFldPath, customEnvSpec.SecretRef.Name, "custom env secret name is duplicated"))
						break
					}
					duplicateEnvMap[customEnvSpec.SecretRef.Name] = 0
				}
			}

			// check HTTPSPort does not conflict with default HTTPPort
			httpsPortFldPath := prodSpecFldPath.Child("httpsPort")

			if apimanager.Spec.Apicast.ProductionSpec.HTTPSPort != nil && *apimanager.Spec.Apicast.ProductionSpec.HTTPSPort == DefaultHTTPPort {
				fieldErrors = append(fieldErrors, field.Invalid(httpsPortFldPath, apimanager.Spec.Apicast.ProductionSpec.HTTPSPort, "HTTPS port conflicts with HTTP port"))
			}
		}

		if apimanager.Spec.Apicast.StagingSpec != nil {
			stagingSpecFldPath := apicastFldPath.Child("stagingSpec")
			customPoliciesFldPath := stagingSpecFldPath.Child("customPolicies")
			duplicatePolicyMap := make(map[string]int)
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
				if _, ok := duplicatePolicyMap[customPolicySpec.VersionName()]; ok {
					fieldErrors = append(fieldErrors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret name version tuple is duplicated"))
					break
				}
				duplicatePolicyMap[customPolicySpec.VersionName()] = 0
			}

			if apimanager.OpenTelemetryEnabledForStaging() {
				openTelemetrySpec := apimanager.Spec.Apicast.StagingSpec.OpenTelemetry
				if openTelemetrySpec.TracingConfigSecretRef != nil {
					if openTelemetrySpec.TracingConfigSecretRef.Name == "" {
						apicastStagingOpenTelemtryTracingFldPath := stagingSpecFldPath.Child("openTelemetry")
						customTracingConfigFldPath := apicastStagingOpenTelemtryTracingFldPath.Child("tracingConfigSecretRef")
						fieldErrors = append(fieldErrors, field.Invalid(customTracingConfigFldPath, openTelemetrySpec, "custom tracing library secret name is empty"))
					}
				}
			}

			if apimanager.IsAPIcastStagingOpenTracingEnabled() {
				openTracingConfigSpec := apimanager.Spec.Apicast.StagingSpec.OpenTracing
				if openTracingConfigSpec.TracingConfigSecretRef != nil {
					if openTracingConfigSpec.TracingConfigSecretRef.Name == "" {
						apicastStagingOpenTracingFldPath := stagingSpecFldPath.Child("openTracing")
						customTracingConfigFldPath := apicastStagingOpenTracingFldPath.Child("tracingConfigSecretRef")
						fieldErrors = append(fieldErrors, field.Invalid(customTracingConfigFldPath, openTracingConfigSpec, "custom tracing library secret name is empty"))
					}
				}
				// For now only "jaeger" is accepted" as the tracing library
				if openTracingConfigSpec.TracingLibrary != nil && *openTracingConfigSpec.TracingLibrary != apps.APIcastDefaultTracingLibrary {
					tracingLibraryFldPath := field.NewPath("spec").
						Child("apicast").
						Child("stagingSpec").
						Child("openTracing").
						Child("tracingLibrary")
					fieldErrors = append(fieldErrors, field.Invalid(tracingLibraryFldPath, openTracingConfigSpec, "invalid tracing library specified"))
				}
			}

			customEnvsFldPath := stagingSpecFldPath.Child("customEnvironments")
			duplicateEnvMap := make(map[string]int)
			// check custom environment secret is set
			for idx, customEnvSpec := range apimanager.Spec.Apicast.StagingSpec.CustomEnvironments {
				customEnvsIdxFldPath := customEnvsFldPath.Index(idx)

				if customEnvSpec.SecretRef == nil {
					fieldErrors = append(fieldErrors, field.Invalid(customEnvsIdxFldPath, customEnvSpec, "custom environment secret is mandatory"))
				} else if customEnvSpec.SecretRef.Name == "" {
					fieldErrors = append(fieldErrors, field.Invalid(customEnvsIdxFldPath, customEnvSpec, "custom environment secret name is empty"))
				} else {
					// check duplicated custom env secret
					if _, ok := duplicateEnvMap[customEnvSpec.SecretRef.Name]; ok {
						fieldErrors = append(fieldErrors, field.Invalid(customEnvsIdxFldPath, customEnvSpec.SecretRef.Name, "custom env secret name is duplicated"))
						break
					}
					duplicateEnvMap[customEnvSpec.SecretRef.Name] = 0
				}
			}

			// check HTTPSPort does not conflict with default HTTPPort
			httpsPortFldPath := stagingSpecFldPath.Child("httpsPort")

			if apimanager.Spec.Apicast.StagingSpec.HTTPSPort != nil && *apimanager.Spec.Apicast.StagingSpec.HTTPSPort == DefaultHTTPPort {
				fieldErrors = append(fieldErrors, field.Invalid(httpsPortFldPath, apimanager.Spec.Apicast.StagingSpec.HTTPSPort, "HTTPS port conflicts with HTTP port"))
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

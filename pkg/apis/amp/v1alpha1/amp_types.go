package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AMPSpec defines the desired state of AMP

// We use pointers to a Type when the field is optional, to allow differentiate
// Between an unset value from the zero value of the type.
// This is a common convention
// used in kubernetes: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#optional-vs-required
type AMPSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	AmpRelease                   string  `json:"ampRelease"`
	AppLabel                     *string `json:"appLabel,omitempty,+optional"`
	TenantName                   *string `json:"tenantName,omitempty,+optional"`
	RwxStorageClass              *string `json:"rwxStorageClass,omitempty,+optional"` // TODO is this correct? maybe a pointer to String?
	AmpBackendImage              *string `json:"ampBackendImage,omitempty,+optional"`
	AmpZyncImage                 *string `json:"ampZyncImage,omitempty,+optional"`
	AmpApicastImage              *string `json:"ampApicastImage,omitempty,+optional"`
	AmpRouterImage               *string `json:"ampRouterImage,omitempty,+optional"`
	AmpSystemImage               *string `json:"ampSystemImage,omitempty,+optional"`
	PostgreSQLImage              *string `json:"postgreSQLImage,omitempty,+optional"`
	MysqlImage                   *string `json:"mysqlImage,omitempty,+optional"`
	MemcachedImage               *string `json:"memcachedImage,omitempty,+optional"`
	ImageStreamTagImportInsecure *bool   `json:"imageStreamTagImportInsecure,omitempty,+optional"`
	RedisImage                   *string `json:"redisImage,omitempty,+optional"`
	MysqlUser                    *string `json:"mysqlUser,omitempty,+optional"`
	MysqlPassword                *string `json:"mysqlPassword,omitempty,+optional"`
	MysqlDatabase                *string `json:"mysqlDatabase,omitempty,+optional"`
	MysqlRootPassword            *string `json:"mysqlRootPassword,omitempty,+optional"` // TODO this should be gathered from a secret
	SystemBackendUsername        *string `json:"systemBackendUsername,omitempty,+optional"`
	SystemBackendPassword        *string `json:"systemBackendPassword,omitempty,+optional"`     // TODO this should be gathered from a secret
	SystemBackendSharedSecret    *string `json:"systemBackendSharedSecret,omitempty,+optional"` // TODO this should be gathered from a secret
	SystemAppSecretKeyBase       *string `json:"systemAppSecretKeyBase,omitempty,+optional"`    // TODO this should be gathered from a secret
	AdminPassword                *string `json:"adminPassword,omitempty,+optional"`             // TODO this should be gathered from a secret
	AdminUsername                *string `json:"adminUsername,omitempty,+optional"`
	AdminAccessToken             *string `json:"adminAccessToken,omitempty,+optional"` // TODO this should be gathered from a secret
	MasterName                   *string `json:"masterName,omitempty,+optional"`
	MasterUser                   *string `json:"masterUser,omitempty,+optional"`
	MasterPassword               *string `json:"masterPassword,omitempty,+optional"`    // TODO this should be gathered from a secret
	MasterAccessToken            *string `json:"masterAccessToken,omitempty,+optional"` // TODO this should be gathered from a secret
	RecaptchaPublicKey           *string `json:"recaptchaPublicKey,omitempty,+optional"`
	RecaptchaPrivateKey          *string `json:"recaptchaPrivateKey,omitempty,+optional"`     // TODO this should be gathered from a secret
	ZyncDatabasePassword         *string `json:"zyncDatabasePassword,omitempty,+optional"`    // TODO this should be gathered from a secret
	ZyncSecretKeyBase            *string `json:"zyncSecretKeyBase,omitempty,+optional"`       // TODO this should be gathered from a secret
	ZyncAuthenticationToken      *string `json:"zyncAuthenticationToken,omitempty,+optional"` // TODO this should be gathered from a secret
	ApicastAccessToken           *string `json:"apicastAccessToken,omitempty,+optional"`      // TODO this should be gathered from a secret
	ApicastManagementApi         *string `json:"apicastManagementApi,omitempty,+optional"`
	ApicastOpenSSLVerify         *bool   `json:"apicastOpenSSLVerify,omitempty,+optional"`
	ApicastResponseCodes         *bool   `json:"apicastResponseCodes,omitempty,+optional"`
	ApicastRegistryURL           *string `json:"apicastRegistryURL,omitempty,+optional"`
	WildcardDomain               string  `json:"wildcardDomain"`
	WildcardPolicy               *string `json:"wildcardPolicy,omitempty,+optional"`
}

// AMPStatus defines the observed state of AMP
type AMPStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AMP is the Schema for the amps API
// +k8s:openapi-gen=true
type AMP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AMPSpec   `json:"spec,omitempty"`
	Status AMPStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AMPList contains a list of AMP
type AMPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AMP `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AMP{}, &AMPList{})
}

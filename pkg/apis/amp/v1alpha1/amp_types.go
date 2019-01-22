package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	optrand "github.com/3scale/3scale-operator/pkg/crypto/rand"
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

	// +optional
	MysqlUser *string `json:"mysqlUser,omitempty"`

	// +optional
	MysqlPassword *string `json:"mysqlPassword,omitempty"`

	// +optional
	MysqlDatabase *string `json:"mysqlDatabase,omitempty"`

	// +optional
	MysqlRootPassword *string `json:"mysqlRootPassword,omitempty"` // TODO this should be gathered from a secret

	// +optional
	SystemBackendUsername *string `json:"systemBackendUsername,omitempty"`

	// +optional
	SystemBackendPassword *string `json:"systemBackendPassword,omitempty"` // TODO this should be gathered from a secret

	// +optional
	SystemBackendSharedSecret *string `json:"systemBackendSharedSecret,omitempty"` // TODO this should be gathered from a secret

	// +optional
	SystemAppSecretKeyBase *string `json:"systemAppSecretKeyBase,omitempty"` // TODO this should be gathered from a secret

	// +optional
	AdminPassword *string `json:"adminPassword,omitempty"` // TODO this should be gathered from a secret

	// +optional
	AdminUsername *string `json:"adminUsername,omitempty"`

	// +optional
	AdminAccessToken *string `json:"adminAccessToken,omitempty"` // TODO this should be gathered from a secret

	// +optional
	MasterName *string `json:"masterName,omitempty"`

	// +optional
	MasterUser *string `json:"masterUser,omitempty"`

	// +optional
	MasterPassword *string `json:"masterPassword,omitempty"` // TODO this should be gathered from a secret

	// +optional
	MasterAccessToken *string `json:"masterAccessToken,omitempty"` // TODO this should be gathered from a secret

	// +optional
	RecaptchaPublicKey *string `json:"recaptchaPublicKey,omitempty"`

	// +optional
	RecaptchaPrivateKey *string `json:"recaptchaPrivateKey,omitempty"` // TODO this should be gathered from a secret

	// +optional
	ZyncDatabasePassword *string `json:"zyncDatabasePassword,omitempty"` // TODO this should be gathered from a secret

	// +optional
	ZyncSecretKeyBase *string `json:"zyncSecretKeyBase,omitempty"` // TODO this should be gathered from a secret

	// +optional
	ZyncAuthenticationToken *string `json:"zyncAuthenticationToken,omitempty"` // TODO this should be gathered from a secret

	// +optional
	ApicastAccessToken *string `json:"apicastAccessToken,omitempty"` // TODO this should be gathered from a secret

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

// SetDefaults sets the default values for the AMP spec and returns true if the spec was changed
func (amp *AMP) SetDefaults() bool {
	changed := false
	ampSpec := &amp.Spec
	if ampSpec.AppLabel == nil {
		defaultAppLabel := "3scale-api-management"
		ampSpec.AppLabel = &defaultAppLabel
		changed = true
	}

	if ampSpec.TenantName == nil {
		defaultTenantName := "3scale"
		ampSpec.TenantName = &defaultTenantName
		changed = true
	}

	if ampSpec.RwxStorageClass == nil { // Needed??
		ampSpec.RwxStorageClass = nil //in OpenShift template would be "null" in the parameter and nul in the field
		changed = true
	}

	if ampSpec.AmpBackendImage == nil {
		defaultAmpBackendImage := "quay.io/3scale/apisonator:nightly"
		ampSpec.AmpBackendImage = &defaultAmpBackendImage
		changed = true
	}

	if ampSpec.AmpZyncImage == nil {
		defaultAmpZyncImage := "quay.io/3scale/zync:nightly"
		ampSpec.AmpZyncImage = &defaultAmpZyncImage
		changed = true
	}

	if ampSpec.AmpApicastImage == nil {
		defaultAmpApicastImage := "quay.io/3scale/apicast:nightly"
		ampSpec.AmpApicastImage = &defaultAmpApicastImage
		changed = true
	}

	if ampSpec.AmpRouterImage == nil {
		defaultAmpRouterImage := "quay.io/3scale/wildcard-router:nightly"
		ampSpec.AmpRouterImage = &defaultAmpRouterImage
		changed = true
	}

	if ampSpec.AmpSystemImage == nil {
		defaultAmpSystemImage := "quay.io/3scale/porta:nightly"
		ampSpec.AmpSystemImage = &defaultAmpSystemImage
		changed = true
	}

	if ampSpec.PostgreSQLImage == nil {
		defaultPostgreSQLImage := "registry.access.redhat.com/rhscl/postgresql-95-rhel7:9.5"
		ampSpec.PostgreSQLImage = &defaultPostgreSQLImage
		changed = true
	}

	if ampSpec.MysqlImage == nil {
		defaultMysqlImage := "registry.access.redhat.com/rhscl/mysql-57-rhel7:5.7"
		ampSpec.MysqlImage = &defaultMysqlImage
		changed = true
	}

	if ampSpec.MemcachedImage == nil {
		defaultMemcachedImage := "registry.access.redhat.com/3scale-amp20/memcached"
		ampSpec.MemcachedImage = &defaultMemcachedImage
		changed = true
	}

	if ampSpec.ImageStreamTagImportInsecure == nil {
		defaultImageStreamTagImportInsecure := false
		ampSpec.ImageStreamTagImportInsecure = &defaultImageStreamTagImportInsecure
		changed = true
	}

	if ampSpec.RedisImage == nil {
		defaultRedisImage := "registry.access.redhat.com/rhscl/redis-32-rhel7:3.2"
		ampSpec.RedisImage = &defaultRedisImage
		changed = true
	}

	if ampSpec.MysqlUser == nil {
		defaultMysqlUser := "mysql"
		ampSpec.MysqlUser = &defaultMysqlUser
		changed = true
	}

	if ampSpec.MysqlPassword == nil {
		defaultMysqlPassword := optrand.String(8)
		ampSpec.MysqlPassword = &defaultMysqlPassword
		changed = true
	}

	if ampSpec.MysqlDatabase == nil {
		defaultMysqlDatabase := "system"
		ampSpec.MysqlDatabase = &defaultMysqlDatabase
		changed = true
	}

	if ampSpec.MysqlRootPassword == nil {
		defaultMysqlRootPassword := optrand.String(8)
		ampSpec.MysqlRootPassword = &defaultMysqlRootPassword
		changed = true
	}

	if ampSpec.SystemBackendUsername == nil {
		defaultSystemBackendUsername := "3scale_api_user"
		ampSpec.SystemBackendUsername = &defaultSystemBackendUsername
		changed = true
	}

	if ampSpec.SystemBackendPassword == nil {
		defaultSystemBackendPassword := optrand.String(8)
		ampSpec.SystemBackendPassword = &defaultSystemBackendPassword
		changed = true
	}

	if ampSpec.SystemBackendSharedSecret == nil {
		defaultSystemBackendSharedSecret := optrand.String(8)
		ampSpec.SystemBackendSharedSecret = &defaultSystemBackendSharedSecret
		changed = true
	}

	if ampSpec.SystemAppSecretKeyBase == nil {
		// TODO is not exactly what we were generating
		// in OpenShift templates. We were generating
		// '[a-f0-9]{128}' . Ask system if there's some reason
		// for that and if we can change it. If must be that range
		// then we should create another function to generate
		// hexadecimal string output
		defaultSystemAppSecretKeyBase := optrand.String(128)
		ampSpec.SystemAppSecretKeyBase = &defaultSystemAppSecretKeyBase
		changed = true
	}

	if ampSpec.AdminPassword == nil {
		defaultAdminPassword := optrand.String(8)
		ampSpec.AdminPassword = &defaultAdminPassword
		changed = true
	}

	if ampSpec.AdminUsername == nil {
		defaultAdminUsername := "admin"
		ampSpec.AdminUsername = &defaultAdminUsername
		changed = true
	}

	if ampSpec.AdminAccessToken == nil {
		defaultAdminAccessToken := optrand.String(16)
		ampSpec.AdminAccessToken = &defaultAdminAccessToken
		changed = true
	}

	if ampSpec.MasterName == nil {
		defaultMasterName := "master"
		ampSpec.MasterName = &defaultMasterName
		changed = true
	}

	if ampSpec.MasterUser == nil {
		defaultMasterUser := "master"
		ampSpec.MasterUser = &defaultMasterUser
		changed = true
	}

	if ampSpec.MasterPassword == nil {
		defaultMasterPassword := optrand.String(8)
		ampSpec.MasterPassword = &defaultMasterPassword
		changed = true
	}

	if ampSpec.MasterAccessToken == nil {
		defaultMasterAccessToken := optrand.String(8)
		ampSpec.MasterAccessToken = &defaultMasterAccessToken
		changed = true
	}

	if ampSpec.RecaptchaPublicKey == nil {
		defaultRecaptchaPublicKey := "" // TODO is this correct? is an empty OpenShift parameter equal to the empty string? or null/nil?
		ampSpec.RecaptchaPublicKey = &defaultRecaptchaPublicKey
		changed = true
	}

	if ampSpec.RecaptchaPrivateKey == nil {
		defaultRecaptchaPrivateKey := "" // TODO is this correct? is an empty OpenShift parameter equal to the empty string? or null/nil?
		ampSpec.RecaptchaPrivateKey = &defaultRecaptchaPrivateKey
		changed = true
	}

	if ampSpec.ZyncDatabasePassword == nil {
		defaultZyncDatabasePassword := optrand.String(16)
		ampSpec.ZyncDatabasePassword = &defaultZyncDatabasePassword
		changed = true
	}

	if ampSpec.ZyncSecretKeyBase == nil {
		defaultZyncSecretKeyBase := optrand.String(16)
		ampSpec.ZyncSecretKeyBase = &defaultZyncSecretKeyBase
		changed = true
	}

	if ampSpec.ZyncAuthenticationToken == nil {
		defaultZyncAuthenticationToken := optrand.String(16)
		ampSpec.ZyncAuthenticationToken = &defaultZyncAuthenticationToken
		changed = true
	}

	if ampSpec.ApicastAccessToken == nil {
		defaultApicastAccessToken := optrand.String(8)
		ampSpec.ApicastAccessToken = &defaultApicastAccessToken
		changed = true
	}

	if ampSpec.ApicastManagementApi == nil {
		defaultApicastManagementApi := "status"
		ampSpec.ApicastManagementApi = &defaultApicastManagementApi
		changed = true
	}

	if ampSpec.ApicastOpenSSLVerify == nil {
		defaultApicastOpenSSLVerify := false
		ampSpec.ApicastOpenSSLVerify = &defaultApicastOpenSSLVerify
		changed = true
	}

	if ampSpec.ApicastResponseCodes == nil {
		defaultApicastResponseCodes := true
		ampSpec.ApicastResponseCodes = &defaultApicastResponseCodes
		changed = true
	}

	if ampSpec.ApicastRegistryURL == nil {
		defaultApicastRegistryURL := "http://apicast-staging:8090/policies"
		ampSpec.ApicastRegistryURL = &defaultApicastRegistryURL
		changed = true
	}

	if ampSpec.WildcardPolicy == nil {
		defaultWildcardPolicy := "None" //TODO should be a set of predefined values (a custom type enum-like to be used)
		ampSpec.WildcardPolicy = &defaultWildcardPolicy
		changed = true
	}

	return changed
}

// Generate random alphanumeric string of size 'size'.
// TODO move variables to constants, the creation of the randomgenerator
// to some kind of singleton/global variable, move it into its own package
// maybe make it concurrencysafe etc...
func generateRandString(size int) string {
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	alphanumeric := alphabet + numbers

	result := make([]byte, size)

	randgen := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	for i := range result {
		result[i] = alphanumeric[randgen.Int63()%int64(len(alphanumeric))]
	}

	return string(result)
}

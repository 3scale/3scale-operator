package component

import (
	"fmt"

	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type S3FileStorageOptions struct {
	ConfigurationSecretName string `validate:"required"`
}

type SystemSMTPSecretOptions struct {
	Address           *string `validate:"required"`
	Authentication    *string `validate:"required"`
	Domain            *string `validate:"required"`
	OpenSSLVerifyMode *string `validate:"required"`
	Password          *string `validate:"required"`
	Port              *string `validate:"required"`
	Username          *string `validate:"required"`
}

type PVCFileStorageOptions struct {
	StorageClass *string
}

type SystemOptions struct {
	MemcachedServers                       string  `validate:"required"`
	EventHooksURL                          string  `validate:"required"`
	RedisURL                               string  `validate:"required"`
	RedisSentinelHosts                     *string `validate:"required"`
	RedisSentinelRole                      *string `validate:"required"`
	RedisNamespace                         *string `validate:"required"`
	MessageBusRedisURL                     *string `validate:"required"`
	MessageBusRedisSentinelHosts           *string `validate:"required"`
	MessageBusRedisSentinelRole            *string `validate:"required"`
	MessageBusRedisNamespace               *string `validate:"required"`
	ApicastSystemMasterProxyConfigEndpoint string  `validate:"required"`
	AdminEmail                             *string `validate:"required"`

	ImageTag string `validate:"required"`

	AppMasterContainerResourceRequirements    *v1.ResourceRequirements `validate:"required"`
	AppProviderContainerResourceRequirements  *v1.ResourceRequirements `validate:"required"`
	AppDeveloperContainerResourceRequirements *v1.ResourceRequirements `validate:"required"`
	SidekiqContainerResourceRequirements      *v1.ResourceRequirements `validate:"required"`
	SphinxContainerResourceRequirements       *v1.ResourceRequirements `validate:"required"`

	S3FileStorageOptions  *S3FileStorageOptions  `validate:"required_without=PvcFileStorageOptions"`
	PvcFileStorageOptions *PVCFileStorageOptions `validate:"required_without=S3FileStorageOptions"`

	AppReplicas     *int32 `validate:"required"`
	SidekiqReplicas *int32 `validate:"required"`

	AdminAccessToken    string  `validate:"required"`
	AdminPassword       string  `validate:"required"`
	AdminUsername       string  `validate:"required"`
	AmpRelease          string  `validate:"required"`
	ApicastAccessToken  string  `validate:"required"`
	ApicastRegistryURL  string  `validate:"required"`
	MasterAccessToken   string  `validate:"required"`
	MasterName          string  `validate:"required"`
	MasterUsername      string  `validate:"required"`
	MasterPassword      string  `validate:"required"`
	RecaptchaPublicKey  *string `validate:"required"`
	RecaptchaPrivateKey *string `validate:"required"`
	AppSecretKeyBase    string  `validate:"required"`
	BackendSharedSecret string  `validate:"required"`
	TenantName          string  `validate:"required"`
	WildcardDomain      string  `validate:"required"`
	SmtpSecretOptions   SystemSMTPSecretOptions

	AppAffinity        *v1.Affinity    `validate:"-"`
	AppTolerations     []v1.Toleration `validate:"-"`
	SidekiqAffinity    *v1.Affinity    `validate:"-"`
	SidekiqTolerations []v1.Toleration `validate:"-"`
	SphinxAffinity     *v1.Affinity    `validate:"-"`
	SphinxTolerations  []v1.Toleration `validate:"-"`

	CommonLabels             map[string]string `validate:"required"`
	CommonAppLabels          map[string]string `validate:"required"`
	AppPodTemplateLabels     map[string]string `validate:"required"`
	CommonSidekiqLabels      map[string]string `validate:"required"`
	SidekiqPodTemplateLabels map[string]string `validate:"required"`
	ProviderUILabels         map[string]string `validate:"required"`
	MasterUILabels           map[string]string `validate:"required"`
	DeveloperUILabels        map[string]string `validate:"required"`
	SphinxLabels             map[string]string `validate:"required"`
	SphinxPodTemplateLabels  map[string]string `validate:"required"`
	MemcachedLabels          map[string]string `validate:"required"`
	SMTPLabels               map[string]string `validate:"required"`
}

func NewSystemOptions() *SystemOptions {
	return &SystemOptions{}
}

func (s *SystemOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

func DefaultApicastSystemMasterProxyConfigEndpoint(token string) string {
	return fmt.Sprintf("http://%s@system-master:3000/master/api/proxy/configs", token)
}

func DefaultMemcachedServers() string {
	return "system-memcache:11211"
}

func DefaultRecaptchaPublickey() string {
	return ""
}

func DefaultRecaptchaPrivatekey() string {
	return ""
}

func DefaultBackendSharedSecret() string {
	return oprand.String(8)
}

func DefaultEventHooksURL() string {
	return "http://system-master:3000/master/events/import"
}

func DefaultSystemRedisURL() string {
	return "redis://system-redis:6379/1"
}

func DefaultSystemRedisSentinelHosts() string {
	return ""
}

func DefaultSystemRedisSentinelRole() string {
	return ""
}

func DefaultSystemMessageBusRedisURL() string {
	return ""
}

func DefaultSystemMessageBusRedisSentinelHosts() string {
	return ""
}

func DefaultSystemMessageBusRedisSentinelRole() string {
	return ""
}

func DefaultSystemRedisNamespace() string {
	return ""
}

func DefaultSystemMessageBusRedisNamespace() string {
	return ""
}

func DefaultSystemAppSecretKeyBase() string {
	// TODO is not exactly what we were generating
	// in OpenShift templates. We were generating
	// '[a-f0-9]{128}' . Ask system if there's some reason
	// for that and if we can change it. If must be that range
	// then we should create another function to generate
	// hexadecimal lowercase string output
	return oprand.String(128)
}

func DefaultSystemMasterName() string {
	return "master"
}

func DefaultSystemMasterUsername() string {
	return "master"
}

func DefaultSystemMasterPassword() string {
	return oprand.String(8)
}

func DefaultSystemAdminUsername() string {
	return "admin"
}

func DefaultSystemAdminPassword() string {
	return oprand.String(8)
}

func DefaultSystemAdminAccessToken() string {
	return oprand.String(16)
}

func DefaultSystemMasterAccessToken() string {
	return oprand.String(8)
}

func DefaultSystemAdminEmail() string {
	return ""
}

func DefaultSystemMasterApicastAccessToken() string {
	return oprand.String(8)
}

func DefaultSystemSMTPAddress() string {
	return ""
}

func DefaultSystemSMTPAuthentication() string {
	return ""
}

func DefaultSystemSMTPDomain() string {
	return ""
}

func DefaultSystemSMTPOpenSSLVerifyMode() string {
	return ""
}

func DefaultSystemSMTPPassword() string {
	return ""
}

func DefaultSystemSMTPPort() string {
	return ""
}

func DefaultSystemSMTPUsername() string {
	return ""
}

func DefaultAppMasterContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("800Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("50m"),
			v1.ResourceMemory: resource.MustParse("600Mi"),
		},
	}
}

func DefaultAppProviderContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("800Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("50m"),
			v1.ResourceMemory: resource.MustParse("600Mi"),
		},
	}
}

func DefaultAppDeveloperContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("800Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("50m"),
			v1.ResourceMemory: resource.MustParse("600Mi"),
		},
	}
}

func DefaultSidekiqContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("2Gi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("500Mi"),
		},
	}
}

func DefaultSphinxContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("80m"),
			v1.ResourceMemory: resource.MustParse("250Mi"),
		},
	}
}

func DefaultAppReplicas() *int32 {
	var defaultReplicas int32 = 1
	return &defaultReplicas
}

func DefaultSidekiqReplicas() *int32 {
	var defaultReplicas int32 = 1
	return &defaultReplicas
}

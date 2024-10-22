package component

import (
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type RedisOptions struct {
	AmpRelease                                string                   `validate:"required"`
	BackendImage                              string                   `validate:"required"`
	BackendImageTag                           string                   `validate:"required"`
	SystemImage                               string                   `validate:"required"`
	SystemImageTag                            string                   `validate:"required"`
	BackendRedisContainerResourceRequirements *v1.ResourceRequirements `validate:"required"`
	SystemRedisContainerResourceRequirements  *v1.ResourceRequirements `validate:"required"`
	BackendRedisPVCStorageClass               *string
	SystemRedisPVCStorageClass                *string

	BackendRedisAffinity    *v1.Affinity    `validate:"-"`
	BackendRedisTolerations []v1.Toleration `validate:"-"`
	SystemRedisAffinity     *v1.Affinity    `validate:"-"`
	SystemRedisTolerations  []v1.Toleration `validate:"-"`

	SystemCommonLabels            map[string]string `validate:"required"`
	SystemRedisLabels             map[string]string `validate:"required"`
	SystemRedisPodTemplateLabels  map[string]string `validate:"required"`
	BackendCommonLabels           map[string]string `validate:"required"`
	BackendRedisLabels            map[string]string `validate:"required"`
	BackendRedisPodTemplateLabels map[string]string `validate:"required"`

	SystemRedisPriorityClassName          string                        `validate:"-"`
	BackendRedisPriorityClassName         string                        `validate:"-"`
	SystemRedisTopologySpreadConstraints  []v1.TopologySpreadConstraint `validate:"-"`
	BackendRedisTopologySpreadConstraints []v1.TopologySpreadConstraint `validate:"-"`
	SystemRedisPodTemplateAnnotations     map[string]string             `validate:"-"`
	BackendRedisPodTemplateAnnotations    map[string]string             `validate:"-"`

	// secrets
	BackendStorageURL                string `validate:"required"`
	BackendQueuesURL                 string `validate:"required"`
	BackendRedisQueuesSentinelHosts  string
	BackendRedisQueuesSentinelRole   string
	BackendRedisStorageSentinelHosts string
	BackendRedisStorageSentinelRole  string
	SystemRedisURL                   string `validate:"required"`
	SystemRedisSentinelsHosts        string
	SystemRedisSentinelsRole         string
	SystemRedisNamespace             string
	// TLS
	SystemRedisCAFile                    string
	SystemRedisClientCertificate         string
	SystemRedisPrivateKey                string
	SystemRedisSSL                       string
	BackendConfigCAFile                  string
	BackendConfigClientCertificate       string
	BackendConfigPrivateKey              string
	BackendConfigSSL                     string
	BackendConfigQueuesCAFile            string
	BackendConfigQueuesClientCertificate string
	BackendConfigQueuesPrivateKey        string
	BackendConfigQueuesSSL               string
}

func NewRedisOptions() *RedisOptions {
	return &RedisOptions{}
}

func (r *RedisOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

func DefaultBackendRedisContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("2000m"),
			v1.ResourceMemory: resource.MustParse("32Gi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("1024Mi"),
		},
	}
}

func DefaultSystemRedisContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("32Gi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("150m"),
			v1.ResourceMemory: resource.MustParse("256Mi"),
		},
	}
}

func DefaultBackendRedisStorageURL() string {
	return "redis://backend-redis:6379/0"
}

func DefaultBackendRedisQueuesURL() string {
	return "redis://backend-redis:6379/1"
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

func DefaultBackendStorageSentinelHosts() string {
	return ""
}

func DefaultBackendStorageSentinelRole() string {
	return ""
}

func DefaultBackendQueuesSentinelHosts() string {
	return ""
}

func DefaultBackendQueuesSentinelRole() string {
	return ""
}

// TLS
func DefaultSystemRedisCAFile() string {
	return ""
}

func DefaultSystemRedisClientCertificate() string {
	return ""
}

func DefaultSystemRedisPrivateKey() string {
	return ""
}

func DefaultSystemRedisSSL() string {
	return "false"
}

// TLS, Backend
func DefaultBackendConfigCAFile() string {
	return ""
}
func DefaultBackendConfigClientCertificate() string {
	return ""
}
func DefaultBackendConfigPrivateKey() string {
	return ""
}
func DefaultBackendConfigSSL() string {
	return "false"
}
func DefaultBackendConfigQueuesCAFile() string {
	return ""
}
func DefaultBackendConfigQueuesClientCertificate() string {
	return ""
}
func DefaultBackendConfigQueuesPrivateKey() string {
	return ""
}
func DefaultBackendConfigQueuesSSL() string {
	return "false"
}

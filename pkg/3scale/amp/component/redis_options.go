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
	InsecureImportPolicy                      *bool                    `validate:"required"`
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

	PriorityClassName string `validate:"-"`
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

func DefaultSystemRedisNamespace() string {
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

package component

import (
	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type BackendOptions struct {
	ServiceEndpoint              string `validate:"required"`
	RouteEndpoint                string `validate:"required"`
	StorageURL                   string `validate:"required"`
	QueuesURL                    string `validate:"required"`
	StorageSentinelHosts         string
	StorageSentinelRole          string
	QueuesSentinelHosts          string
	QueuesSentinelRole           string
	ListenerResourceRequirements v1.ResourceRequirements `validate:"-"`
	WorkerResourceRequirements   v1.ResourceRequirements `validate:"-"`
	CronResourceRequirements     v1.ResourceRequirements `validate:"-"`
	ListenerReplicas             int32
	WorkerReplicas               int32
	CronReplicas                 int32
	AppLabel                     string `validate:"required"`
	SystemBackendUsername        string `validate:"required"`
	SystemBackendPassword        string `validate:"required"`
	TenantName                   string `validate:"required"`
	WildcardDomain               string `validate:"required"`
}

func NewBackendOptions() *BackendOptions {
	return &BackendOptions{}
}

func (b *BackendOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(b)
}

func DefaultBackendServiceEndpoint() string {
	return "http://backend-listener:3000"
}

func DefaultBackendRedisStorageURL() string {
	return "redis://backend-redis:6379/0"
}

func DefaultBackendRedisQueuesURL() string {
	return "redis://backend-redis:6379/1"
}

func DefaultBackendListenerResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("700Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("550Mi"),
		},
	}
}

func DefaultBackendWorkerResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("300Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("150m"),
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
	}
}

func DefaultCronResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("150m"),
			v1.ResourceMemory: resource.MustParse("80Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("50m"),
			v1.ResourceMemory: resource.MustParse("40Mi"),
		},
	}
}

func DefaultSystemBackendUsername() string {
	return "3scale_api_user"
}

func DefaultSystemBackendPassword() string {
	return oprand.String(8)
}

package component

import (
	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type BackendOptions struct {
	ServiceEndpoint              string                  `validate:"required"`
	RouteEndpoint                string                  `validate:"required"`
	ImageTag                     string                  `validate:"required"`
	ListenerResourceRequirements v1.ResourceRequirements `validate:"-"`
	WorkerResourceRequirements   v1.ResourceRequirements `validate:"-"`
	CronResourceRequirements     v1.ResourceRequirements `validate:"-"`
	ListenerReplicas             int32
	WorkerReplicas               int32
	CronReplicas                 int32
	SystemBackendUsername        string            `validate:"required"`
	SystemBackendPassword        string            `validate:"required"`
	TenantName                   string            `validate:"required"`
	WildcardDomain               string            `validate:"required"`
	ListenerAffinity             *v1.Affinity      `validate:"-"`
	ListenerTolerations          []v1.Toleration   `validate:"-"`
	WorkerAffinity               *v1.Affinity      `validate:"-"`
	WorkerTolerations            []v1.Toleration   `validate:"-"`
	CronAffinity                 *v1.Affinity      `validate:"-"`
	CronTolerations              []v1.Toleration   `validate:"-"`
	CommonLabels                 map[string]string `validate:"required"`
	CommonListenerLabels         map[string]string `validate:"required"`
	CommonWorkerLabels           map[string]string `validate:"required"`
	CommonCronLabels             map[string]string `validate:"required"`
	ListenerPodTemplateLabels    map[string]string `validate:"required"`
	WorkerPodTemplateLabels      map[string]string `validate:"required"`
	CronPodTemplateLabels        map[string]string `validate:"required"`
	WorkerMetrics                bool
	ListenerMetrics              bool

	// Used for monitoring objects
	// Those objects are namespaced. However, objects includes labels, rules and expressions
	// that need namespace filtering because they are "global" once imported
	// to the prometheus or grafana services.
	Namespace string `validate:"required"`
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
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("500Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("100Mi"),
		},
	}
}

func DefaultSystemBackendUsername() string {
	return "3scale_api_user"
}

func DefaultSystemBackendPassword() string {
	return oprand.String(8)
}

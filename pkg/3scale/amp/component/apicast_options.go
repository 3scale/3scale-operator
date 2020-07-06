package component

import (
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ApicastOptions struct {
	ManagementAPI                  string `validate:"required"`
	OpenSSLVerify                  string `validate:"required"`
	ResponseCodes                  string `validate:"required"`
	ImageTag                       string `validate:"required"`
	ExtendedMetrics                bool
	ProductionResourceRequirements v1.ResourceRequirements `validate:"-"`
	StagingResourceRequirements    v1.ResourceRequirements `validate:"-"`
	ProductionReplicas             int32
	StagingReplicas                int32
	CommonLabels                   map[string]string `validate:"required"`
	CommonStagingLabels            map[string]string `validate:"required"`
	CommonProductionLabels         map[string]string `validate:"required"`
	StagingPodTemplateLabels       map[string]string `validate:"required"`
	ProductionPodTemplateLabels    map[string]string `validate:"required"`
	ProductionAffinity             *v1.Affinity      `validate:"-"`
	ProductionTolerations          []v1.Toleration   `validate:"-"`
	StagingAffinity                *v1.Affinity      `validate:"-"`
	StagingTolerations             []v1.Toleration   `validate:"-"`
}

func NewApicastOptions() *ApicastOptions {
	return &ApicastOptions{}
}

func (a *ApicastOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}

func DefaultProductionResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("64Mi"),
		},
	}
}

func DefaultStagingResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("50m"),
			v1.ResourceMemory: resource.MustParse("64Mi"),
		},
	}
}

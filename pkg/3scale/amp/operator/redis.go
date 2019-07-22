package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	v1 "k8s.io/api/core/v1"
)

func (o *OperatorRedisOptionsProvider) GetRedisOptions() (*component.RedisOptions, error) {
	optProv := component.RedisOptionsBuilder{}

	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Redis Options - %s", err)
	}
	return res, nil
}

func (o *OperatorRedisOptionsProvider) setResourceRequirementsOptions(b *component.RedisOptionsBuilder) {
	if !*o.APIManagerSpec.ResourceRequirementsEnabled {
		b.SystemRedisContainerResourceRequirements(v1.ResourceRequirements{})
		b.BackendRedisContainerResourceRequirements(v1.ResourceRequirements{})
	}
}

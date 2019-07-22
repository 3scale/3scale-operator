package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	v1 "k8s.io/api/core/v1"
)

func (o *OperatorMemcachedOptionsProvider) GetMemcachedOptions() (*component.MemcachedOptions, error) {
	optProv := component.MemcachedOptionsBuilder{}
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)

	o.setResourceRequirementsOptions(&optProv)

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Memcached Options - %s", err)
	}
	return res, nil
}

func (o *OperatorMemcachedOptionsProvider) setResourceRequirementsOptions(b *component.MemcachedOptionsBuilder) {
	if !*o.APIManagerSpec.ResourceRequirementsEnabled {
		b.ResourceRequirements(v1.ResourceRequirements{})
	}
}

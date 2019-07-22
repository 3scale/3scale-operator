package component

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type RedisOptions struct {
	// required options
	appLabel string

	// non-required options
	backendRedisContainerResourceRequirements *v1.ResourceRequirements
	systemRedisContainerResourceRequirements  *v1.ResourceRequirements
}

type RedisOptionsBuilder struct {
	options RedisOptions
}

func (r *RedisOptionsBuilder) AppLabel(appLabel string) {
	r.options.appLabel = appLabel
}

func (r *RedisOptionsBuilder) SystemRedisContainerResourceRequirements(resourceRequirements v1.ResourceRequirements) {
	r.options.systemRedisContainerResourceRequirements = &resourceRequirements
}

func (r *RedisOptionsBuilder) BackendRedisContainerResourceRequirements(resourceRequirements v1.ResourceRequirements) {
	r.options.backendRedisContainerResourceRequirements = &resourceRequirements
}

func (r *RedisOptionsBuilder) Build() (*RedisOptions, error) {
	err := r.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	r.setNonRequiredOptions()

	return &r.options, nil
}

func (r *RedisOptionsBuilder) setRequiredOptions() error {
	if r.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}

	return nil
}

func (r *RedisOptionsBuilder) setNonRequiredOptions() {
	if r.options.backendRedisContainerResourceRequirements == nil {
		r.options.backendRedisContainerResourceRequirements = r.defaultBackendRedisContainerResourceRequirements()
	}

	if r.options.systemRedisContainerResourceRequirements == nil {
		r.options.systemRedisContainerResourceRequirements = r.defaultSystemRedisContainerResourceRequirements()
	}
}

func (r *RedisOptionsBuilder) defaultBackendRedisContainerResourceRequirements() *v1.ResourceRequirements {
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

func (r *RedisOptionsBuilder) defaultSystemRedisContainerResourceRequirements() *v1.ResourceRequirements {
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

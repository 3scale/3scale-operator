package component

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type MemcachedOptions struct {
	// memcachedRequiredOptions
	appLabel string

	// memcached non-required options
	resourceRequirements *v1.ResourceRequirements
}

type MemcachedOptionsBuilder struct {
	options MemcachedOptions
}

func (m *MemcachedOptionsBuilder) AppLabel(appLabel string) {
	m.options.appLabel = appLabel
}

func (m *MemcachedOptionsBuilder) ResourceRequirements(resourceRequirements v1.ResourceRequirements) {
	m.options.resourceRequirements = &resourceRequirements
}

func (m *MemcachedOptionsBuilder) Build() (*MemcachedOptions, error) {
	err := m.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	m.setNonRequiredOptions()

	return &m.options, nil
}

func (m *MemcachedOptionsBuilder) setRequiredOptions() error {
	if m.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}

	return nil
}

func (m *MemcachedOptionsBuilder) setNonRequiredOptions() {
	if m.options.resourceRequirements == nil {
		m.options.resourceRequirements = m.defaultResourceRequirements()
	}
}

func (m *MemcachedOptionsBuilder) defaultResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("250m"),
			v1.ResourceMemory: resource.MustParse("96Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("50m"),
			v1.ResourceMemory: resource.MustParse("64Mi"),
		},
	}
}

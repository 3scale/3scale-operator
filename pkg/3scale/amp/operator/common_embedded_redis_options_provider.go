package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
)

type CommonEmbeddedRedisOptionProvider struct {
	apimanager *appsv1alpha1.APIManager
	options    *component.CommonEmbeddedRedisOptions
}

func NewCommonEmbeddedRedisOptionProvider(apimanager *appsv1alpha1.APIManager) *CommonEmbeddedRedisOptionProvider {
	return &CommonEmbeddedRedisOptionProvider{
		apimanager: apimanager,
		options:    component.NewCommonEmbeddedRedisOptions(),
	}
}

func (r *CommonEmbeddedRedisOptionProvider) GetCommonEmbeddedRedisOptions() (*component.CommonEmbeddedRedisOptions, error) {
	r.options.ConfigMapLabels = r.configMapLabels()

	err := r.options.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetCommonEmbeddedRedisOptions validating: %w", err)
	}
	return r.options, nil
}

func (r *CommonEmbeddedRedisOptionProvider) systemCommonLabels() map[string]string {
	return map[string]string{
		"app":                  *r.apimanager.Spec.AppLabel,
		"threescale_component": "system",
	}
}

func (r *CommonEmbeddedRedisOptionProvider) configMapLabels() map[string]string {
	labels := r.systemCommonLabels()
	labels["threescale_component_element"] = "redis"
	return labels
}

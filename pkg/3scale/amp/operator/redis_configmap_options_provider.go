package operator

import (
	"fmt"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

type RedisConfigMapOptionsProvider struct {
	apimanager *appsv1alpha1.APIManager
	options    *component.RedisConfigMapOptions
}

func NewRedisConfigMapOptionsProvider(apimanager *appsv1alpha1.APIManager) *RedisConfigMapOptionsProvider {
	return &RedisConfigMapOptionsProvider{
		apimanager: apimanager,
		options:    component.NewRedisConfigMapOptions(),
	}
}

func (r *RedisConfigMapOptionsProvider) GetOptions() (*component.RedisConfigMapOptions, error) {
	r.options.Labels = r.systemRedisLabels()
	r.options.Namespace = r.apimanager.Namespace

	err := r.options.Validate()
	if err != nil {
		return nil, fmt.Errorf("RedisConfigMapOptionsProvider validating: %w", err)
	}
	return r.options, nil
}

func (r *RedisConfigMapOptionsProvider) systemRedisLabels() map[string]string {
	return component.SystemRedisLabels(*r.apimanager.Spec.AppLabel)
}

func (r *RedisConfigMapOptionsProvider) systemCommonLabels() map[string]string {
	return component.SystemCommonLabels(*r.apimanager.Spec.AppLabel)
}

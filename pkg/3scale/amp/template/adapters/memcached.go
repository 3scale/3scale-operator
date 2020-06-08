package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type MemcachedAdapter struct {
}

func NewMemcachedAdapter() Adapter {
	return NewAppenderAdapter(&MemcachedAdapter{})
}

func (m *MemcachedAdapter) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{}
}

func (m *MemcachedAdapter) Objects() ([]common.KubernetesObject, error) {
	memcachedOptions, err := m.options()
	if err != nil {
		return nil, err
	}
	memcachedComponent := component.NewMemcached(memcachedOptions)
	objects := m.componentObjects(memcachedComponent)
	return objects, nil
}

func (m *MemcachedAdapter) componentObjects(c *component.Memcached) []common.KubernetesObject {
	deploymentConfig := c.DeploymentConfig()

	objects := []common.KubernetesObject{
		deploymentConfig,
	}
	return objects
}

func (m *MemcachedAdapter) options() (*component.MemcachedOptions, error) {
	mo := component.NewMemcachedOptions()
	mo.ImageTag = "${AMP_RELEASE}"
	mo.ResourceRequirements = component.DefaultMemcachedResourceRequirements()

	mo.DeploymentLabels = m.deploymentLabels()
	mo.PodTemplateLabels = m.podTemplateLabels()

	err := mo.Validate()
	return mo, err
}

func (m *MemcachedAdapter) deploymentLabels() map[string]string {
	return map[string]string{
		"app":                          "${APP_LABEL}",
		"threescale_component":         "system",
		"threescale_component_element": "memcache",
	}
}

func (m *MemcachedAdapter) podTemplateLabels() map[string]string {
	labels := m.deploymentLabels()
	labels["deploymentConfig"] = "system-memcache"
	return labels
}

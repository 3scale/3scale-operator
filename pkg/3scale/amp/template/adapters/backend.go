package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type Backend struct {
	generatePodDisruptionBudget bool
}

func NewBackendAdapter(generatePDB bool) Adapter {
	return NewAppenderAdapter(&Backend{
		generatePodDisruptionBudget:generatePDB,
	})
}

func (b *Backend) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{}
}

func (b *Backend) Objects() ([]common.KubernetesObject, error) {
	backendOptions, err := b.options()
	if err != nil {
		return nil, err
	}
	backendComponent := component.NewBackend(backendOptions)
	objects := backendComponent.Objects()
	if b.generatePodDisruptionBudget {
		objects = append(objects, backendComponent.PDBObjects()...)
	}
	return objects, nil
}

func (b *Backend) options() (*component.BackendOptions, error) {
	bob := component.BackendOptionsBuilder{}
	bob.AppLabel("${APP_LABEL}")
	bob.SystemBackendUsername("${SYSTEM_BACKEND_USERNAME}")
	bob.SystemBackendPassword("${SYSTEM_BACKEND_PASSWORD}")
	bob.TenantName("${TENANT_NAME}")
	bob.WildcardDomain("${WILDCARD_DOMAIN}")
	return bob.Build()
}

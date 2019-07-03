package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type Backend struct {
}

func NewBackendAdapter(options []string) Adapter {
	return NewAppenderAdapter(&Backend{})
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
	return backendComponent.Objects(), nil
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

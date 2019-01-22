package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorBackendOptionsProvider) GetBackendOptions() (*component.BackendOptions, error) {
	optProv := component.BackendOptionsBuilder{}
	optProv.AppLabel(*o.AmpSpec.AppLabel)
	optProv.SystemBackendUsername(*o.AmpSpec.SystemBackendUsername)
	optProv.SystemBackendPassword(*o.AmpSpec.SystemBackendPassword)
	optProv.TenantName(*o.AmpSpec.TenantName)
	optProv.WildcardDomain(o.AmpSpec.WildcardDomain)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Backend Options - %s", err)
	}
	return res, nil
}

package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorWildcardRouterOptionsProvider) GetWildcardRouterOptions() (*component.WildcardRouterOptions, error) {
	optProv := component.WildcardRouterOptionsBuilder{}
	optProv.AppLabel(*o.AmpSpec.AppLabel)
	optProv.WildcardDomain(o.AmpSpec.WildcardDomain)
	optProv.WildcardPolicy(*o.AmpSpec.WildcardPolicy)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create WildcardRouter Options - %s", err)
	}
	return res, nil
}

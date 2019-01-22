package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorZyncOptionsProvider) GetZyncOptions() (*component.ZyncOptions, error) {
	optProv := component.ZyncOptionsBuilder{}
	optProv.AppLabel(*o.AmpSpec.AppLabel)
	optProv.AuthenticationToken(*o.AmpSpec.ZyncAuthenticationToken)
	optProv.DatabasePassword(*o.AmpSpec.ZyncDatabasePassword)
	optProv.SecretKeyBase(*o.AmpSpec.ZyncSecretKeyBase)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Zync Options - %s", err)
	}
	return res, nil
}

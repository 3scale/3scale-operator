package operator

import (
	"fmt"
	"strconv"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorApicastOptionsProvider) GetApicastOptions() (*component.ApicastOptions, error) {
	optProv := component.ApicastOptionsBuilder{}
	optProv.AppLabel(*o.AmpSpec.AppLabel)
	optProv.ManagementAPI(*o.AmpSpec.ApicastManagementApi)
	optProv.OpenSSLVerify(strconv.FormatBool(*o.AmpSpec.ApicastOpenSSLVerify)) // TODO is this a good place to make the conversion?
	optProv.ResponseCodes(strconv.FormatBool(*o.AmpSpec.ApicastResponseCodes)) // TODO is this a good place to make the conversion?
	optProv.TenantName(*o.AmpSpec.TenantName)
	optProv.WildcardDomain(o.AmpSpec.WildcardDomain)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Apicast Options - %s", err)
	}
	return res, nil
}

package operator

import (
	"fmt"
	"strconv"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorApicastOptionsProvider) GetApicastOptions() (*component.ApicastOptions, error) {
	optProv := component.ApicastOptionsBuilder{}
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.TenantName(*o.APIManagerSpec.TenantName)
	optProv.WildcardDomain(o.APIManagerSpec.WildcardDomain)
	optProv.ManagementAPI(*o.APIManagerSpec.Apicast.ApicastManagementAPI)
	optProv.OpenSSLVerify(strconv.FormatBool(*o.APIManagerSpec.Apicast.OpenSSLVerify))        // TODO is this a good place to make the conversion?
	optProv.ResponseCodes(strconv.FormatBool(*o.APIManagerSpec.Apicast.IncludeResponseCodes)) // TODO is this a good place to make the conversion?

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Apicast Options - %s", err)
	}
	return res, nil
}

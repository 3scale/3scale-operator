package operator

import (
	"fmt"
	"strconv"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	v1 "k8s.io/api/core/v1"
)

func (o *OperatorApicastOptionsProvider) GetApicastOptions() (*component.ApicastOptions, error) {
	optProv := component.ApicastOptionsBuilder{}
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.TenantName(*o.APIManagerSpec.TenantName)
	optProv.WildcardDomain(o.APIManagerSpec.WildcardDomain)
	optProv.ManagementAPI(*o.APIManagerSpec.Apicast.ApicastManagementAPI)
	optProv.OpenSSLVerify(strconv.FormatBool(*o.APIManagerSpec.Apicast.OpenSSLVerify))        // TODO is this a good place to make the conversion?
	optProv.ResponseCodes(strconv.FormatBool(*o.APIManagerSpec.Apicast.IncludeResponseCodes)) // TODO is this a good place to make the conversion?

	o.setResourceRequirementsOptions(&optProv)
	o.setReplicas(&optProv)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Apicast Options - %s", err)
	}
	return res, nil
}

func (o *OperatorApicastOptionsProvider) setResourceRequirementsOptions(b *component.ApicastOptionsBuilder) {
	if !*o.APIManagerSpec.ResourceRequirementsEnabled {
		b.StagingResourceRequirements(v1.ResourceRequirements{})
		b.ProductionResourceRequirements(v1.ResourceRequirements{})
	}
}

func (o *OperatorApicastOptionsProvider) setReplicas(b *component.ApicastOptionsBuilder) {
	if o.APIManagerSpec.HighAvailability != nil && o.APIManagerSpec.HighAvailability.Enabled {
		b.StagingReplicas(2)
		b.ProductionReplicas(2)
	}
}

package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type Apicast struct {
}

func NewApicastAdapter() Adapter {
	return NewAppenderAdapter(&Apicast{})
}

func (a *Apicast) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "APICAST_ACCESS_TOKEN",
			Description: "Read Only Access Token that is APIcast going to use to download its configuration.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "APICAST_MANAGEMENT_API",
			Description: "Scope of the APIcast Management API. Can be disabled, status or debug. At least status required for health checks.",
			Value:       "status",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "APICAST_OPENSSL_VERIFY",
			Description: "Turn on/off the OpenSSL peer verification when downloading the configuration. Can be set to true/false.",
			Value:       "false",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "APICAST_RESPONSE_CODES",
			Description: "Enable logging response codes in APIcast.",
			Value:       "true",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "APICAST_REGISTRY_URL",
			Description: "The URL to point to APIcast policies registry management",
			Value:       "http://apicast-staging:8090/policies",
			Required:    true,
		},
	}
}

func (a *Apicast) Objects() ([]common.KubernetesObject, error) {
	apicastOptions, err := a.options()
	if err != nil {
		return nil, err
	}
	apicastComponent := component.NewApicast(apicastOptions)
	return apicastComponent.Objects(), nil
}

func (a *Apicast) options() (*component.ApicastOptions, error) {
	aob := &component.ApicastOptionsBuilder{}
	aob.AppLabel("${APP_LABEL}")
	aob.ManagementAPI("${APICAST_MANAGEMENT_API}")
	aob.OpenSSLVerify("${APICAST_OPENSSL_VERIFY}")
	aob.ResponseCodes("${APICAST_RESPONSE_CODES}")
	aob.TenantName("${TENANT_NAME}")
	aob.WildcardDomain("${WILDCARD_DOMAIN}")
	return aob.Build()
}

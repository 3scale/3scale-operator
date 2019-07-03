package component

import "fmt"

type ApicastOptions struct {
	appLabel       string
	managementAPI  string
	openSSLVerify  string
	responseCodes  string
	tenantName     string
	wildcardDomain string
}

type ApicastOptionsBuilder struct {
	options ApicastOptions
}

func (a *ApicastOptionsBuilder) AppLabel(appLabel string) {
	a.options.appLabel = appLabel
}

func (a *ApicastOptionsBuilder) ManagementAPI(managementAPI string) {
	a.options.managementAPI = managementAPI
}

func (a *ApicastOptionsBuilder) OpenSSLVerify(openSSLVerify string) {
	a.options.openSSLVerify = openSSLVerify
}

func (a *ApicastOptionsBuilder) ResponseCodes(responseCodes string) {
	a.options.responseCodes = responseCodes
}

func (a *ApicastOptionsBuilder) TenantName(tenantName string) {
	a.options.tenantName = tenantName
}

func (a *ApicastOptionsBuilder) WildcardDomain(wildcardDomain string) {
	a.options.wildcardDomain = wildcardDomain
}

func (a *ApicastOptionsBuilder) Build() (*ApicastOptions, error) {
	err := a.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	return &a.options, nil
}

func (a *ApicastOptionsBuilder) setRequiredOptions() error {
	if a.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if a.options.managementAPI == "" {
		return fmt.Errorf("no management API has been provided")
	}
	if a.options.openSSLVerify == "" {
		return fmt.Errorf("no OpenSSLVerify option has been provided")
	}
	if a.options.responseCodes == "" {
		return fmt.Errorf("no response codes have been provided")
	}
	if a.options.tenantName == "" {
		return fmt.Errorf("no tenant name has been provided")
	}
	if a.options.wildcardDomain == "" {
		return fmt.Errorf("no wildcard domain has been provided")
	}

	return nil
}

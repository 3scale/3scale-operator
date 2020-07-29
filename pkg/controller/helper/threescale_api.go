package helper

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

const (
	HTTP_VERBOSE_ENVVAR = "THREESCALE_DEBUG"
)

type ProviderAccount struct {
	AdminURLStr string
	Token       string
}

// PortaClient instantiate porta_client.ThreeScaleClient from ProviderAccount object
func PortaClient(providerAccount *ProviderAccount) (*threescaleapi.ThreeScaleClient, error) {
	return PortaClientFromURLString(providerAccount.AdminURLStr, providerAccount.Token)
}

func PortaClientFromURLString(adminURLStr, token string) (*threescaleapi.ThreeScaleClient, error) {
	adminURL, err := url.Parse(adminURLStr)
	if err != nil {
		return nil, err
	}
	return PortaClientFromURL(adminURL, token)
}

// PortaClientFromURL instantiates porta_client.ThreeScaleClient from admin url object
func PortaClientFromURL(url *url.URL, token string) (*threescaleapi.ThreeScaleClient, error) {
	adminPortal, err := threescaleapi.NewAdminPortal(url.Scheme, url.Hostname(), helper.PortFromURL(url))
	if err != nil {
		return nil, err
	}

	// TODO By default should not skip verification
	// Activated by some env var or Spec param
	var transport http.RoundTripper = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if helper.GetEnvVar(HTTP_VERBOSE_ENVVAR, "0") == "1" {
		transport = &helper.Transport{Transport: transport}
	}

	return threescaleapi.NewThreeScale(adminPortal, token, &http.Client{Transport: transport}), nil
}

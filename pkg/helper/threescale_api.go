package helper

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// controllerName is the name of this controller
	providerAccountDefaultSecretName = "threescale-provider-account"

	// providerAccountSecretURLFieldName is the field name of the provider account secret where URL can be found
	providerAccountSecretURLFieldName = "adminURL"

	// providerAccountSecretURLFieldName is the field name of the provider account secret where token can be found
	providerAccountSecretTokenFieldName = "token"
)

// LookupThreescaleClient looks up for account provider credentials to build 3scale API client
// If provider_account_reference is provided, it must exist and required fields must exists
// If no provider_account_reference is provided, defaul provider account secret with hardcoded name will be looked up in the namespace.
// If no provider_account_reference is provided AND default provider account secret is not found either, then,
// 3scale default provider account (3scale-admin) will be looked up using system-seed secret in the current namespace.
// If nothing is successfully found, return error
func LookupThreescaleClient(cl client.Client, ns string, providerAccountRef *corev1.LocalObjectReference) (*threescaleapi.ThreeScaleClient, error) {
	secretSource := NewSecretSource(cl, ns)
	if providerAccountRef != nil {
		adminURL, err := secretSource.RequiredFieldValueFromRequiredSecret(providerAccountRef.Name, providerAccountSecretURLFieldName)
		if err != nil {
			return nil, err
		}
		token, err := secretSource.RequiredFieldValueFromRequiredSecret(providerAccountRef.Name, providerAccountSecretTokenFieldName)
		if err != nil {
			return nil, err
		}

		return PortaClientFromURLString(adminURL, token)
	}

	// Default provider account reference
	// if exists, fiels are required.
	defaulSecret, err := GetSecret(providerAccountDefaultSecretName, ns, cl)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		// Not found
		// NO-OP
	} else {
		adminURL := GetSecretDataValue(defaulSecret.Data, providerAccountSecretURLFieldName)
		if adminURL == nil {
			return nil, fmt.Errorf("Secret field '%s' is required in secret '%s'", providerAccountSecretURLFieldName, defaulSecret.Name)
		}
		token := GetSecretDataValue(defaulSecret.Data, providerAccountSecretTokenFieldName)
		if err != nil {
			return nil, fmt.Errorf("Secret field '%s' is required in secret '%s'", providerAccountSecretURLFieldName, defaulSecret.Name)
		}

		return PortaClientFromURLString(*adminURL, *token)
	}

	// Lookup 3scale installation in current namespace
	// TODO: Check apimanger CR exists?

	// TODO implement read from existing 3scale installation
	return nil, nil
}

// PortaClientFromURLString instantiate porta_client.ThreeScaleClient from admin url string
func PortaClientFromURLString(adminURLStr, masterAccessToken string) (*threescaleapi.ThreeScaleClient, error) {
	adminURL, err := url.Parse(adminURLStr)
	if err != nil {
		return nil, err
	}
	return PortaClient(adminURL, masterAccessToken)
}

// PortaClient instantiates porta_client.ThreeScaleClient from admin url object
func PortaClient(url *url.URL, masterAccessToken string) (*threescaleapi.ThreeScaleClient, error) {
	adminPortal, err := threescaleapi.NewAdminPortal(url.Scheme, url.Hostname(), PortFromURL(url))
	if err != nil {
		return nil, err
	}

	// TODO By default should not skip verification
	// Activated by some env var or Spec param
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return threescaleapi.NewThreeScale(adminPortal, masterAccessToken, &http.Client{Transport: tr}), nil
}

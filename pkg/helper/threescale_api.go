package helper

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	HTTP_VERBOSE_ENVVAR = "THREESCALE_DEBUG"
)

var (
	// controllerName is the name of this controller
	providerAccountDefaultSecretName = "threescale-provider-account"

	// providerAccountSecretURLFieldName is the field name of the provider account secret where URL can be found
	providerAccountSecretURLFieldName = "adminURL"

	// providerAccountSecretURLFieldName is the field name of the provider account secret where token can be found
	providerAccountSecretTokenFieldName = "token"
)

type ProviderAccount struct {
	AdminURLStr string
	Token       string
}

// LookupProviderAccount looks up for account provider url and credentials
// If provider_account_reference is provided, it must exist and required fields must exists
// If no provider_account_reference is provided, defaul provider account secret with hardcoded name will be looked up in the namespace.
// If no provider_account_reference is provided AND default provider account secret is not found either, then,
// 3scale default provider account (3scale-admin) will be looked up using system-seed secret in the current namespace.
// If nothing is successfully found, return error
func LookupProviderAccount(cl client.Client, ns string, providerAccountRef *corev1.LocalObjectReference, logger logr.Logger) (*ProviderAccount, error) {
	if providerAccountRef != nil {
		logger.Info("LookupProviderAccount", "ns", ns, "providerAccountRef", providerAccountRef)
		secretSource := NewSecretSource(cl, ns)
		adminURLStr, err := secretSource.RequiredFieldValueFromRequiredSecret(providerAccountRef.Name, providerAccountSecretURLFieldName)
		if err != nil {
			return nil, fmt.Errorf("LookupProviderAccount: %w", err)
		}
		token, err := secretSource.RequiredFieldValueFromRequiredSecret(providerAccountRef.Name, providerAccountSecretTokenFieldName)
		if err != nil {
			return nil, fmt.Errorf("LookupProviderAccount: %w", err)
		}

		logger.Info("LookupProviderAccount providerAccountRef found", "adminURL", adminURLStr)
		return &ProviderAccount{AdminURLStr: adminURLStr, Token: token}, nil
	}

	// Default provider account reference
	// if exists, fiels are required.
	defaulSecret, err := GetSecret(providerAccountDefaultSecretName, ns, cl)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("LookupProviderAccount: %w", err)
		}
		// Not found
		// NO-OP
	} else {
		adminURLStr := GetSecretDataValue(defaulSecret.Data, providerAccountSecretURLFieldName)
		if adminURLStr == nil {
			return nil, fmt.Errorf("LookupProviderAccount: Secret field '%s' is required in secret '%s'", providerAccountSecretURLFieldName, defaulSecret.Name)
		}
		token := GetSecretDataValue(defaulSecret.Data, providerAccountSecretTokenFieldName)
		if err != nil {
			return nil, fmt.Errorf("LookupProviderAccount: Secret field '%s' is required in secret '%s'", providerAccountSecretURLFieldName, defaulSecret.Name)
		}

		logger.Info("LookupProviderAccount default secret found", "adminURL", adminURLStr)
		return &ProviderAccount{AdminURLStr: *adminURLStr, Token: *token}, nil
	}

	// Lookup 3scale installation in current namespace
	// TODO: Check apimanger CR exists?

	// TODO implement read from existing 3scale installation
	logger.Info("LookupProviderAccount no provider account found")
	return nil, nil
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
	adminPortal, err := threescaleapi.NewAdminPortal(url.Scheme, url.Hostname(), PortFromURL(url))
	if err != nil {
		return nil, err
	}

	// TODO By default should not skip verification
	// Activated by some env var or Spec param
	var transport http.RoundTripper = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if GetEnvVar(HTTP_VERBOSE_ENVVAR, "0") == "1" {
		transport = &Transport{Transport: transport}
	}

	return threescaleapi.NewThreeScale(adminPortal, token, &http.Client{Transport: transport}), nil
}

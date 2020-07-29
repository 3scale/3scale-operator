package helper

import (
	"context"
	"errors"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// controllerName is the name of this controller
	providerAccountDefaultSecretName = "threescale-provider-account"

	// providerAccountSecretURLFieldName is the field name of the provider account secret where URL can be found
	providerAccountSecretURLFieldName = "adminURL"

	// providerAccountSecretTokenFieldName is the field name of the provider account secret where token can be found
	providerAccountSecretTokenFieldName = "token"
)

// LookupProviderAccount looks up for account provider url and credentials
// If provider_account_reference is provided, it must exist and required fields must exists
// If no provider_account_reference is provided, defaul provider account secret with hardcoded name will be looked up in the namespace.
// If no provider_account_reference is provided AND default provider account secret is not found either, then,
// 3scale default provider account (3scale-admin) will be looked up using system-seed secret in the current namespace.
// If nothing is successfully found, return error
func LookupProviderAccount(cl client.Client, ns string, providerAccountRef *corev1.LocalObjectReference, logger logr.Logger) (*ProviderAccount, error) {
	if providerAccountRef != nil {
		logger.Info("LookupProviderAccount", "ns", ns, "providerAccountRef", providerAccountRef)
		secretSource := helper.NewSecretSource(cl, ns)
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

	// Default provider account secret
	// if exists, fiels are required.
	defaulSecret, err := helper.GetSecret(providerAccountDefaultSecretName, ns, cl)
	if err == nil {
		adminURLStr := helper.GetSecretDataValue(defaulSecret.Data, providerAccountSecretURLFieldName)
		if adminURLStr == nil {
			return nil, fmt.Errorf("LookupProviderAccount: Secret field '%s' is required in secret '%s'", providerAccountSecretURLFieldName, defaulSecret.Name)
		}
		token := helper.GetSecretDataValue(defaulSecret.Data, providerAccountSecretTokenFieldName)
		if token == nil {
			return nil, fmt.Errorf("LookupProviderAccount: Secret field '%s' is required in secret '%s'", providerAccountSecretTokenFieldName, defaulSecret.Name)
		}

		logger.Info("LookupProviderAccount default secret found", "adminURL", adminURLStr)
		return &ProviderAccount{AdminURLStr: *adminURLStr, Token: *token}, nil
	} else if err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("LookupProviderAccount: %w", err)
	}

	// Lookup 3scale installation in current namespace
	// Read credentials and tenant url for default provider account of 3scale
	listOps := []client.ListOption{client.InNamespace(ns)}
	apimanagerList := &appsv1alpha1.APIManagerList{}
	err = cl.List(context.TODO(), apimanagerList, listOps...)
	if err != nil {
		return nil, err
	}

	if len(apimanagerList.Items) > 0 {
		apimanager := apimanagerList.Items[0]
		wildcardDomain := apimanager.Spec.WildcardDomain
		if apimanager.Spec.TenantName == nil {
			return nil, fmt.Errorf("LookupProviderAccount: apimanager found, '%s', but tenantName is empty", apimanager.Name)
		}

		tenantName := *apimanager.Spec.TenantName
		// TODO read route tls conf to determine HTTP or HTTPS
		adminURL := fmt.Sprintf("%s://%s-admin.%s", "https", tenantName, wildcardDomain)

		// Read access token from secret
		// if exists, fiels are required.
		credSecret, err := helper.GetSecret(component.SystemSecretSystemSeedSecretName, ns, cl)
		if err == nil {
			accessToken := helper.GetSecretDataValue(credSecret.Data, component.SystemSecretSystemSeedAdminAccessTokenFieldName)
			if accessToken == nil {
				return nil, fmt.Errorf("LookupProviderAccount: Secret field '%s' is required in secret '%s'", component.SystemSecretSystemSeedAdminAccessTokenFieldName, component.SystemSecretSystemSeedSecretName)
			}

			logger.Info("LookupProviderAccount 3scale installation found", "adminURL", adminURL)
			return &ProviderAccount{AdminURLStr: adminURL, Token: *accessToken}, nil
		} else if err != nil && !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("LookupProviderAccount: %w", err)
		}
	}

	// not found, return error
	return nil, errors.New("LookupProviderAccount: no provider account found")
}

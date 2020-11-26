package controllers

import (
	"fmt"
	"strconv"

	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

func (t *ProductThreescaleReconciler) syncProxy(_ interface{}) error {
	existing, err := t.productEntity.Proxy()
	if err != nil {
		return fmt.Errorf("Error sync product [%s] proxy: %w", t.resource.Spec.SystemName, err)
	}

	// respect 3scale defaults.
	// If some setting is not set in CR, will not be reconcile, respecting 3scale defaults.

	params := threescaleapi.Params{}

	// Production public base url
	prodPublicBaseURL := t.resource.Spec.ProdPublicBaseURL()
	if prodPublicBaseURL != nil {
		if helper.SetURLDefaultPort(existing.Element.Endpoint) != helper.SetURLDefaultPort(*prodPublicBaseURL) {
			params["endpoint"] = *prodPublicBaseURL
		}
	}

	// Production public base url
	stagingPublicBaseURL := t.resource.Spec.StagingPublicBaseURL()
	if stagingPublicBaseURL != nil {
		if helper.SetURLDefaultPort(existing.Element.SandboxEndpoint) != helper.SetURLDefaultPort(*stagingPublicBaseURL) {
			params["sandbox_endpoint"] = *stagingPublicBaseURL
		}
	}

	// Security: Secret token
	secretToken := t.resource.Spec.SecuritySecretToken()
	if secretToken != nil {
		if existing.Element.SecretToken != *secretToken {
			params["secret_token"] = *secretToken
		}
	}

	// Security: Host rewrite
	hostRewrite := t.resource.Spec.HostRewrite()
	if hostRewrite != nil {
		if existing.Element.HostnameRewrite != *hostRewrite {
			params["hostname_rewrite"] = *hostRewrite
		}
	}

	// Credentials Location
	credentialsLocation := t.resource.Spec.CredentialsLocation()
	if credentialsLocation != nil {
		if existing.Element.CredentialsLocation != *credentialsLocation {
			params["credentials_location"] = *credentialsLocation
		}
	}

	// auth user key
	authUserKey := t.resource.Spec.AuthUserKey()
	if authUserKey != nil {
		if existing.Element.AuthUserKey != *authUserKey {
			params["auth_user_key"] = *authUserKey
		}
	}

	// auth app id
	authAppID := t.resource.Spec.AuthAppID()
	if authAppID != nil {
		if existing.Element.AuthAppID != *authAppID {
			params["auth_app_id"] = *authAppID
		}
	}

	// auth app key
	authAppKey := t.resource.Spec.AuthAppKey()
	if authAppKey != nil {
		if existing.Element.AuthAppKey != *authAppKey {
			params["auth_app_key"] = *authAppKey
		}
	}

	t.syncProxyGatewayResponse(params, existing)

	t.syncProxyOIDC(params, existing)

	if len(params) > 0 {
		err := t.productEntity.UpdateProxy(params)
		if err != nil {
			return fmt.Errorf("Error updating product proxy: %w", err)
		}
	}
	return nil
}

func (t *ProductThreescaleReconciler) syncProxyGatewayResponse(params threescaleapi.Params, existing *threescaleapi.ProxyJSON) {
	gatewayResponse := t.resource.Spec.GatewayResponse()
	if gatewayResponse == nil {
		return
	}

	intCases := []struct {
		desired  *int32
		existing int
		field    string
	}{
		{
			gatewayResponse.ErrorStatusAuthFailed,
			existing.Element.ErrorStatusAuthFailed,
			"error_status_auth_failed",
		},
		{
			gatewayResponse.ErrorStatusAuthMissing,
			existing.Element.ErrorStatusAuthMissing,
			"error_status_auth_missing",
		},
		{
			gatewayResponse.ErrorStatusNoMatch,
			existing.Element.ErrorStatusNoMatch,
			"error_status_no_match",
		},
		{
			gatewayResponse.ErrorStatusLimitsExceeded,
			existing.Element.ErrorStatusLimitsExceeded,
			"error_status_limits_exceeded",
		},
	}
	for _, intCase := range intCases {
		if intCase.desired != nil && int32(intCase.existing) != *intCase.desired {
			params[intCase.field] = strconv.Itoa(int(*intCase.desired))
		}
	}

	strCases := []struct {
		desired  *string
		existing string
		field    string
	}{
		{
			gatewayResponse.ErrorHeadersAuthFailed,
			existing.Element.ErrorHeadersAuthFailed,
			"error_headers_auth_failed",
		},
		{
			gatewayResponse.ErrorAuthFailed,
			existing.Element.ErrorAuthFailed,
			"error_auth_failed",
		},
		{
			gatewayResponse.ErrorHeadersAuthMissing,
			existing.Element.ErrorHeadersAuthMissing,
			"error_headers_auth_missing",
		},
		{
			gatewayResponse.ErrorAuthMissing,
			existing.Element.ErrorAuthMissing,
			"error_auth_missing",
		},
		{
			gatewayResponse.ErrorHeadersNoMatch,
			existing.Element.ErrorHeadersNoMatch,
			"error_headers_no_match",
		},
		{
			gatewayResponse.ErrorNoMatch,
			existing.Element.ErrorNoMatch,
			"error_no_match",
		},
		{
			gatewayResponse.ErrorHeadersLimitsExceeded,
			existing.Element.ErrorHeadersLimitsExceeded,
			"error_headers_limits_exceeded",
		},
		{
			gatewayResponse.ErrorLimitsExceeded,
			existing.Element.ErrorLimitsExceeded,
			"error_limits_exceeded",
		},
	}
	for _, strCase := range strCases {
		if strCase.desired != nil && strCase.existing != *strCase.desired {
			params[strCase.field] = *strCase.desired
		}
	}
}

func (t *ProductThreescaleReconciler) syncProxyOIDC(params threescaleapi.Params, existing *threescaleapi.ProxyJSON) {
	oidcSpec := t.resource.Spec.OIDCSpec()
	if oidcSpec == nil {
		return
	}

	if existing.Element.OidcIssuerEndpoint != oidcSpec.IssuerEndpoint {
		params["oidc_issuer_endpoint"] = oidcSpec.IssuerEndpoint
	}

	if existing.Element.OidcIssuerType != oidcSpec.IssuerType {
		params["oidc_issuer_type"] = oidcSpec.IssuerType
	}

	if oidcSpec.JwtClaimWithClientID != nil && existing.Element.JwtClaimWithClientID != *oidcSpec.JwtClaimWithClientID {
		params["jwt_claim_with_client_id"] = *oidcSpec.JwtClaimWithClientID
	}

	if oidcSpec.JwtClaimWithClientIDType != nil && existing.Element.JwtClaimWithClientIDType != *oidcSpec.JwtClaimWithClientIDType {
		params["jwt_claim_with_client_id_type"] = *oidcSpec.JwtClaimWithClientIDType
	}
}

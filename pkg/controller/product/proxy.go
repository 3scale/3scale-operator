package product

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

func (t *ThreescaleReconciler) syncProxy(_ interface{}) error {
	existing, err := t.entity.Proxy()
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

	if len(params) > 0 {
		err := t.entity.UpdateProxy(params)
		if err != nil {
			return fmt.Errorf("Error updating product proxy: %w", err)
		}
	}
	return nil
}

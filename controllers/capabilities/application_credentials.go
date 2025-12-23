package controllers

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ThreescaleCredentialTypeUserKey = "1"
	ThreescaleCredentialTypeAppID   = "2"
	ThreescaleCredentialTypeOIDC    = "oidc"
)

const (
	ThreescaleCredentialParamUserKey = "user_key"
	ThreescaleCredentialParamAppID   = "application_id"
	ThreescaleCredentialParamAppKey  = "application_key"
)

const (
	CredentialSecretKeyNameUserKey      = "UserKey"
	CredentialSecretKeyNameAppID        = "ApplicationID"
	CredentialSecretKeyNameAppKey       = "ApplicationKey"
	CredentialSecretKeyNameClientID     = "ClientID"
	CredentialSecretKeyNameClientSecret = "ClientSecret"
)

// extractApplicationCredentialType returns the credential type from the product
// setting
func extractApplicationCredentialType(productResource *capabilitiesv1beta1.Product) (string, error) {
	credType := productResource.Spec.AuthenticationMode()
	if credType == nil {
		return "", fmt.Errorf("unable to identify authentication mode from Product CR")
	}
	return *credType, nil
}

func validateApplicationCrendentialSecret(s *corev1.Secret, authMode string) error {
	nn := client.ObjectKeyFromObject(s)

	switch authMode {
	case ThreescaleCredentialTypeUserKey:
		if err := validateSecretForAuthModeUserKey(s); err != nil {
			return err
		}
	case ThreescaleCredentialTypeAppID:
		if err := validateSecretForAuthModeAppIDAppKey(s); err != nil {
			return err
		}
	case ThreescaleCredentialTypeOIDC:
		if err := validateSecretForAuthModeOIDC(s); err != nil {
			return err
		}
	default:
		return fmt.Errorf("secret %s used, but has unsupported type %s", nn, authMode)
	}
	return nil
}

func validateSecretForAuthModeUserKey(s *corev1.Secret) error {
	if _, ok := s.Data[CredentialSecretKeyNameUserKey]; !ok {
		return fmt.Errorf("secret %s used as user-key authentication mode, but lacks %s key",
			client.ObjectKeyFromObject(s), CredentialSecretKeyNameUserKey,
		)
	}
	return nil
}

func validateSecretForAuthModeAppIDAppKey(s *corev1.Secret) error {
	if _, ok := s.Data[CredentialSecretKeyNameAppID]; !ok {
		return fmt.Errorf("secret %s used as app-id/app-key authentication mode, but lacks %s key",
			client.ObjectKeyFromObject(s), CredentialSecretKeyNameAppID,
		)
	}
	if _, ok := s.Data[CredentialSecretKeyNameAppKey]; !ok {
		return fmt.Errorf("secret %s used as app-id/app-key authentication mode, but lacks %s key",
			client.ObjectKeyFromObject(s), CredentialSecretKeyNameAppKey,
		)
	}
	return nil
}

func validateSecretForAuthModeOIDC(s *corev1.Secret) error {
	if _, ok := s.Data[CredentialSecretKeyNameClientID]; !ok {
		return fmt.Errorf("secret %s used as oidc authentication mode, but lacks %s key",
			client.ObjectKeyFromObject(s), CredentialSecretKeyNameClientID,
		)
	}
	if _, ok := s.Data[CredentialSecretKeyNameClientSecret]; !ok {
		return fmt.Errorf("secret %s used as oidc authentication mode, but lacks %s key",
			client.ObjectKeyFromObject(s), CredentialSecretKeyNameClientSecret,
		)
	}
	return nil
}

func handleCredentials(creds *corev1.Secret, authType string) (map[string]string, error) {
	authParams := make(map[string]string)
	switch authType {
	case ThreescaleCredentialTypeUserKey:
		authParams[ThreescaleCredentialParamUserKey] = string(creds.Data[CredentialSecretKeyNameUserKey])
	case ThreescaleCredentialTypeAppID:
		authParams[ThreescaleCredentialParamAppID] = string(creds.Data[CredentialSecretKeyNameAppID])
		authParams[ThreescaleCredentialParamAppKey] = string(creds.Data[CredentialSecretKeyNameAppKey])
	case ThreescaleCredentialTypeOIDC:
		authParams[ThreescaleCredentialParamAppID] = string(creds.Data[CredentialSecretKeyNameClientID])
		authParams[ThreescaleCredentialParamAppKey] = string(creds.Data[CredentialSecretKeyNameClientSecret])
	default:
		return nil, fmt.Errorf("unknown authentication mode")
	}
	return authParams, nil
}

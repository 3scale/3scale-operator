package v1beta1

import (
	"github.com/3scale/3scale-operator/apis/capabilities/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this Product to the Hub version (v1beta2).
func (src *Product) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.Product)

	// Set the ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Set the spec
	dst.Spec.Name = src.Spec.Name
	dst.Spec.SystemName = src.Spec.SystemName
	dst.Spec.Description = src.Spec.Description
	if src.Spec.Deployment != nil {
		dst.Spec.Deployment = &v1beta2.ProductDeploymentSpec{}
		if src.Spec.Deployment.ApicastHosted != nil && src.Spec.Deployment.ApicastHosted.Authentication != nil {
			dst.Spec.Deployment.ApicastHosted = &v1beta2.ApicastHostedSpec{
				Authentication: &v1beta2.AuthenticationSpec{},
			}
			if src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication != nil {
				dst.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Key = src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Key
				dst.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.CredentialsLoc = src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.CredentialsLoc
				if src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Security != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Security = &v1beta2.SecuritySpec{
						HostHeader:  src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Security.HostHeader,
						SecretToken: src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Security.SecretToken,
					}
				}
				if src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse = &v1beta2.GatewayResponseSpec{
						ErrorStatusAuthFailed:      src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorStatusAuthFailed,
						ErrorHeadersAuthFailed:     src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorHeadersAuthFailed,
						ErrorAuthFailed:            src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorAuthFailed,
						ErrorStatusAuthMissing:     src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorStatusAuthMissing,
						ErrorHeadersAuthMissing:    src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorHeadersAuthMissing,
						ErrorAuthMissing:           src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorAuthMissing,
						ErrorStatusNoMatch:         src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorStatusNoMatch,
						ErrorHeadersNoMatch:        src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorHeadersNoMatch,
						ErrorNoMatch:               src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorNoMatch,
						ErrorStatusLimitsExceeded:  src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorStatusLimitsExceeded,
						ErrorHeadersLimitsExceeded: src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorHeadersLimitsExceeded,
						ErrorLimitsExceeded:        src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorLimitsExceeded,
					}
				}

			}
			if src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication != nil {
				dst.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication = &v1beta2.AppKeyAppIDAuthenticationSpec{
					AppID:          src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.AppID,
					AppKey:         src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.AppKey,
					CredentialsLoc: src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.CredentialsLoc,
				}
				if src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.Security != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.Security = &v1beta2.SecuritySpec{
						HostHeader:  src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.Security.HostHeader,
						SecretToken: src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.Security.SecretToken,
					}
				}
				if src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse = &v1beta2.GatewayResponseSpec{
						ErrorStatusAuthFailed:      src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorStatusAuthFailed,
						ErrorHeadersAuthFailed:     src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorHeadersAuthFailed,
						ErrorAuthFailed:            src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorAuthFailed,
						ErrorStatusAuthMissing:     src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorStatusAuthMissing,
						ErrorHeadersAuthMissing:    src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorHeadersAuthMissing,
						ErrorAuthMissing:           src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorAuthMissing,
						ErrorStatusNoMatch:         src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorStatusNoMatch,
						ErrorHeadersNoMatch:        src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorHeadersNoMatch,
						ErrorNoMatch:               src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorNoMatch,
						ErrorStatusLimitsExceeded:  src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorStatusLimitsExceeded,
						ErrorHeadersLimitsExceeded: src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorHeadersLimitsExceeded,
						ErrorLimitsExceeded:        src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorLimitsExceeded,
					}
				}

			}
			if src.Spec.Deployment.ApicastHosted.Authentication.OIDC != nil {
				dst.Spec.Deployment.ApicastHosted.Authentication.OIDC = &v1beta2.OIDCSpec{
					IssuerType:               src.Spec.Deployment.ApicastHosted.Authentication.OIDC.IssuerType,
					IssuerEndpoint:           src.Spec.Deployment.ApicastHosted.Authentication.OIDC.IssuerEndpoint,
					JwtClaimWithClientID:     src.Spec.Deployment.ApicastHosted.Authentication.OIDC.JwtClaimWithClientID,
					JwtClaimWithClientIDType: src.Spec.Deployment.ApicastHosted.Authentication.OIDC.JwtClaimWithClientIDType,
					CredentialsLoc:           src.Spec.Deployment.ApicastHosted.Authentication.OIDC.CredentialsLoc,
				}
				if src.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow = &v1beta2.OIDCAuthenticationFlowSpec{
						StandardFlowEnabled:       src.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow.StandardFlowEnabled,
						ImplicitFlowEnabled:       src.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow.ImplicitFlowEnabled,
						ServiceAccountsEnabled:    src.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow.ServiceAccountsEnabled,
						DirectAccessGrantsEnabled: src.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow.DirectAccessGrantsEnabled,
					}

				}
				if src.Spec.Deployment.ApicastHosted.Authentication.OIDC.Security != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.OIDC.Security = &v1beta2.SecuritySpec{
						HostHeader:  src.Spec.Deployment.ApicastHosted.Authentication.OIDC.Security.HostHeader,
						SecretToken: src.Spec.Deployment.ApicastHosted.Authentication.OIDC.Security.SecretToken,
					}
				}
				if src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse = &v1beta2.GatewayResponseSpec{
						ErrorStatusAuthFailed:      src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorStatusAuthFailed,
						ErrorHeadersAuthFailed:     src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorHeadersAuthFailed,
						ErrorAuthFailed:            src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorAuthFailed,
						ErrorStatusAuthMissing:     src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorStatusAuthMissing,
						ErrorHeadersAuthMissing:    src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorHeadersAuthMissing,
						ErrorAuthMissing:           src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorAuthMissing,
						ErrorStatusNoMatch:         src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorStatusNoMatch,
						ErrorHeadersNoMatch:        src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorHeadersNoMatch,
						ErrorNoMatch:               src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorNoMatch,
						ErrorStatusLimitsExceeded:  src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorStatusLimitsExceeded,
						ErrorHeadersLimitsExceeded: src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorHeadersLimitsExceeded,
						ErrorLimitsExceeded:        src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorLimitsExceeded,
					}
				}
			}
		}
	}

	dst.Spec.MappingRules = []v1beta2.MappingRuleSpec{}
	for i := range src.Spec.MappingRules {
		dst.Spec.MappingRules = append(dst.Spec.MappingRules, v1beta2.MappingRuleSpec{
			HTTPMethod:      src.Spec.MappingRules[i].HTTPMethod,
			Pattern:         src.Spec.MappingRules[i].Pattern,
			MetricMethodRef: src.Spec.MappingRules[i].MetricMethodRef,
			Increment:       src.Spec.MappingRules[i].Increment,
			Last:            src.Spec.MappingRules[i].Last,
		})
	}

	dst.Spec.BackendUsages = map[string]v1beta2.BackendUsageSpec{}
	for s := range src.Spec.BackendUsages {
		dst.Spec.BackendUsages[s] = v1beta2.BackendUsageSpec{
			Path: src.Spec.BackendUsages[s].Path,
		}
	}

	dst.Spec.Metrics = map[string]v1beta2.MetricSpec{}
	for s := range src.Spec.Metrics {
		dst.Spec.Metrics[s] = v1beta2.MetricSpec{
			Name:        src.Spec.Metrics[s].Name,
			Unit:        src.Spec.Metrics[s].Unit,
			Description: src.Spec.Metrics[s].Description,
		}
	}

	dst.Spec.Methods = map[string]v1beta2.MethodSpec{}
	for s := range src.Spec.Methods {
		dst.Spec.Methods[s] = v1beta2.MethodSpec{
			Name:        src.Spec.Methods[s].Name,
			Description: src.Spec.Methods[s].Description,
		}
	}

	dst.Spec.ApplicationPlans = map[string]v1beta2.ApplicationPlanSpec{}
	for s := range src.Spec.ApplicationPlans {
		plan := v1beta2.ApplicationPlanSpec{
			Name:                src.Spec.ApplicationPlans[s].Name,
			AppsRequireApproval: src.Spec.ApplicationPlans[s].AppsRequireApproval,
			TrialPeriod:         src.Spec.ApplicationPlans[s].TrialPeriod,
			SetupFee:            src.Spec.ApplicationPlans[s].SetupFee,
			CostMonth:           src.Spec.ApplicationPlans[s].CostMonth,
			PricingRules:        nil,
			Limits:              nil,
			Published:           src.Spec.ApplicationPlans[s].Published,
		}

		for i := range src.Spec.ApplicationPlans[s].PricingRules {
			plan.PricingRules = append(plan.PricingRules, v1beta2.PricingRuleSpec{
				From: src.Spec.ApplicationPlans[s].PricingRules[i].From,
				To:   src.Spec.ApplicationPlans[s].PricingRules[i].To,
				MetricMethodRef: v1beta2.MetricMethodRefSpec{
					SystemName:        src.Spec.ApplicationPlans[s].PricingRules[i].MetricMethodRef.SystemName,
					BackendSystemName: src.Spec.ApplicationPlans[s].PricingRules[i].MetricMethodRef.BackendSystemName,
				},
				PricePerUnit: src.Spec.ApplicationPlans[s].PricingRules[i].PricePerUnit,
			})
		}

		dst.Spec.ApplicationPlans[s] = plan
	}

	dst.Spec.ProviderAccountRef = src.Spec.ProviderAccountRef

	for i := range src.Spec.Policies {
		dst.Spec.Policies = append(dst.Spec.Policies, v1beta2.PolicyConfig{
			Name:          src.Spec.Policies[i].Name,
			Version:       src.Spec.Policies[i].Version,
			Configuration: v1beta2.Configuration{Value: src.Spec.Policies[i].Configuration},
			Enabled:       src.Spec.Policies[i].Enabled,
		})
	}

	// Set the status
	dst.Status.ID = src.Status.ID
	dst.Status.State = src.Status.State
	dst.Status.ProviderAccountHost = src.Status.ProviderAccountHost
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Conditions = src.Status.Conditions
	return nil
}

// ConvertFrom converts from the Hub version (v1beta2) to this version.
func (dst *Product) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.Product)

	// Set the ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Set the spec
	dst.Spec.Name = src.Spec.Name
	dst.Spec.SystemName = src.Spec.SystemName
	dst.Spec.Description = src.Spec.Description
	if src.Spec.Deployment != nil {
		dst.Spec.Deployment = &ProductDeploymentSpec{}
		if src.Spec.Deployment.ApicastHosted != nil && src.Spec.Deployment.ApicastHosted.Authentication != nil {
			dst.Spec.Deployment.ApicastHosted = &ApicastHostedSpec{
				Authentication: &AuthenticationSpec{},
			}
			if src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication != nil {
				dst.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Key = src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Key
				dst.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.CredentialsLoc = src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.CredentialsLoc
				if src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Security != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Security = &SecuritySpec{
						HostHeader:  src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Security.HostHeader,
						SecretToken: src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.Security.SecretToken,
					}
				}
				if src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse = &GatewayResponseSpec{
						ErrorStatusAuthFailed:      src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorStatusAuthFailed,
						ErrorHeadersAuthFailed:     src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorHeadersAuthFailed,
						ErrorAuthFailed:            src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorAuthFailed,
						ErrorStatusAuthMissing:     src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorStatusAuthMissing,
						ErrorHeadersAuthMissing:    src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorHeadersAuthMissing,
						ErrorAuthMissing:           src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorAuthMissing,
						ErrorStatusNoMatch:         src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorStatusNoMatch,
						ErrorHeadersNoMatch:        src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorHeadersNoMatch,
						ErrorNoMatch:               src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorNoMatch,
						ErrorStatusLimitsExceeded:  src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorStatusLimitsExceeded,
						ErrorHeadersLimitsExceeded: src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorHeadersLimitsExceeded,
						ErrorLimitsExceeded:        src.Spec.Deployment.ApicastHosted.Authentication.UserKeyAuthentication.GatewayResponse.ErrorLimitsExceeded,
					}
				}

			}
			if src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication != nil {
				dst.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication = &AppKeyAppIDAuthenticationSpec{
					AppID:          src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.AppID,
					AppKey:         src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.AppKey,
					CredentialsLoc: src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.CredentialsLoc,
				}
				if src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.Security != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.Security = &SecuritySpec{
						HostHeader:  src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.Security.HostHeader,
						SecretToken: src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.Security.SecretToken,
					}
				}
				if src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse = &GatewayResponseSpec{
						ErrorStatusAuthFailed:      src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorStatusAuthFailed,
						ErrorHeadersAuthFailed:     src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorHeadersAuthFailed,
						ErrorAuthFailed:            src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorAuthFailed,
						ErrorStatusAuthMissing:     src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorStatusAuthMissing,
						ErrorHeadersAuthMissing:    src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorHeadersAuthMissing,
						ErrorAuthMissing:           src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorAuthMissing,
						ErrorStatusNoMatch:         src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorStatusNoMatch,
						ErrorHeadersNoMatch:        src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorHeadersNoMatch,
						ErrorNoMatch:               src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorNoMatch,
						ErrorStatusLimitsExceeded:  src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorStatusLimitsExceeded,
						ErrorHeadersLimitsExceeded: src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorHeadersLimitsExceeded,
						ErrorLimitsExceeded:        src.Spec.Deployment.ApicastHosted.Authentication.AppKeyAppIDAuthentication.GatewayResponse.ErrorLimitsExceeded,
					}
				}

			}
			if src.Spec.Deployment.ApicastHosted.Authentication.OIDC != nil {
				dst.Spec.Deployment.ApicastHosted.Authentication.OIDC = &OIDCSpec{
					IssuerType:               src.Spec.Deployment.ApicastHosted.Authentication.OIDC.IssuerType,
					IssuerEndpoint:           src.Spec.Deployment.ApicastHosted.Authentication.OIDC.IssuerEndpoint,
					JwtClaimWithClientID:     src.Spec.Deployment.ApicastHosted.Authentication.OIDC.JwtClaimWithClientID,
					JwtClaimWithClientIDType: src.Spec.Deployment.ApicastHosted.Authentication.OIDC.JwtClaimWithClientIDType,
					CredentialsLoc:           src.Spec.Deployment.ApicastHosted.Authentication.OIDC.CredentialsLoc,
				}
				if src.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow = &OIDCAuthenticationFlowSpec{
						StandardFlowEnabled:       src.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow.StandardFlowEnabled,
						ImplicitFlowEnabled:       src.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow.ImplicitFlowEnabled,
						ServiceAccountsEnabled:    src.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow.ServiceAccountsEnabled,
						DirectAccessGrantsEnabled: src.Spec.Deployment.ApicastHosted.Authentication.OIDC.AuthenticationFlow.DirectAccessGrantsEnabled,
					}

				}
				if src.Spec.Deployment.ApicastHosted.Authentication.OIDC.Security != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.OIDC.Security = &SecuritySpec{
						HostHeader:  src.Spec.Deployment.ApicastHosted.Authentication.OIDC.Security.HostHeader,
						SecretToken: src.Spec.Deployment.ApicastHosted.Authentication.OIDC.Security.SecretToken,
					}
				}
				if src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse != nil {
					dst.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse = &GatewayResponseSpec{
						ErrorStatusAuthFailed:      src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorStatusAuthFailed,
						ErrorHeadersAuthFailed:     src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorHeadersAuthFailed,
						ErrorAuthFailed:            src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorAuthFailed,
						ErrorStatusAuthMissing:     src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorStatusAuthMissing,
						ErrorHeadersAuthMissing:    src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorHeadersAuthMissing,
						ErrorAuthMissing:           src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorAuthMissing,
						ErrorStatusNoMatch:         src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorStatusNoMatch,
						ErrorHeadersNoMatch:        src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorHeadersNoMatch,
						ErrorNoMatch:               src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorNoMatch,
						ErrorStatusLimitsExceeded:  src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorStatusLimitsExceeded,
						ErrorHeadersLimitsExceeded: src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorHeadersLimitsExceeded,
						ErrorLimitsExceeded:        src.Spec.Deployment.ApicastHosted.Authentication.OIDC.GatewayResponse.ErrorLimitsExceeded,
					}
				}
			}
		}
	}

	dst.Spec.MappingRules = []MappingRuleSpec{}
	for i := range src.Spec.MappingRules {
		dst.Spec.MappingRules = append(dst.Spec.MappingRules, MappingRuleSpec{
			HTTPMethod:      src.Spec.MappingRules[i].HTTPMethod,
			Pattern:         src.Spec.MappingRules[i].Pattern,
			MetricMethodRef: src.Spec.MappingRules[i].MetricMethodRef,
			Increment:       src.Spec.MappingRules[i].Increment,
			Last:            src.Spec.MappingRules[i].Last,
		})
	}

	dst.Spec.BackendUsages = map[string]BackendUsageSpec{}
	for s := range src.Spec.BackendUsages {
		dst.Spec.BackendUsages[s] = BackendUsageSpec{
			Path: src.Spec.BackendUsages[s].Path,
		}
	}

	dst.Spec.Metrics = map[string]MetricSpec{}
	for s := range src.Spec.Metrics {
		dst.Spec.Metrics[s] = MetricSpec{
			Name:        src.Spec.Metrics[s].Name,
			Unit:        src.Spec.Metrics[s].Unit,
			Description: src.Spec.Metrics[s].Description,
		}
	}

	dst.Spec.Methods = map[string]MethodSpec{}
	for s := range src.Spec.Methods {
		dst.Spec.Methods[s] = MethodSpec{
			Name:        src.Spec.Methods[s].Name,
			Description: src.Spec.Methods[s].Description,
		}
	}

	dst.Spec.ApplicationPlans = map[string]ApplicationPlanSpec{}
	for s := range src.Spec.ApplicationPlans {
		plan := ApplicationPlanSpec{
			Name:                src.Spec.ApplicationPlans[s].Name,
			AppsRequireApproval: src.Spec.ApplicationPlans[s].AppsRequireApproval,
			TrialPeriod:         src.Spec.ApplicationPlans[s].TrialPeriod,
			SetupFee:            src.Spec.ApplicationPlans[s].SetupFee,
			CostMonth:           src.Spec.ApplicationPlans[s].CostMonth,
			PricingRules:        nil,
			Limits:              nil,
			Published:           src.Spec.ApplicationPlans[s].Published,
		}

		for i := range src.Spec.ApplicationPlans[s].PricingRules {
			plan.PricingRules = append(plan.PricingRules, PricingRuleSpec{
				From: src.Spec.ApplicationPlans[s].PricingRules[i].From,
				To:   src.Spec.ApplicationPlans[s].PricingRules[i].To,
				MetricMethodRef: MetricMethodRefSpec{
					SystemName:        src.Spec.ApplicationPlans[s].PricingRules[i].MetricMethodRef.SystemName,
					BackendSystemName: src.Spec.ApplicationPlans[s].PricingRules[i].MetricMethodRef.BackendSystemName,
				},
				PricePerUnit: src.Spec.ApplicationPlans[s].PricingRules[i].PricePerUnit,
			})
		}

		dst.Spec.ApplicationPlans[s] = plan
	}

	dst.Spec.ProviderAccountRef = src.Spec.ProviderAccountRef

	for i := range src.Spec.Policies {
		dst.Spec.Policies = append(dst.Spec.Policies, PolicyConfig{
			Name:          src.Spec.Policies[i].Name,
			Version:       src.Spec.Policies[i].Version,
			Configuration: src.Spec.Policies[i].Configuration.Value,
			Enabled:       src.Spec.Policies[i].Enabled,
		})
	}

	// Set the status
	dst.Status.ID = src.Status.ID
	dst.Status.State = src.Status.State
	dst.Status.ProviderAccountHost = src.Status.ProviderAccountHost
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Conditions = src.Status.Conditions
	return nil
}

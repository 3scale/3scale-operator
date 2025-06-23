package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// LastSlashRegexp matches the last slash
var LastSlashRegexp = regexp.MustCompile(`/$`)

type OpenAPIProductReconciler struct {
	*reconcilers.BaseReconciler
	openapiCR       *capabilitiesv1beta1.OpenAPI
	openapiObj      *openapi3.T
	providerAccount *controllerhelper.ProviderAccount
	logger          logr.Logger
}

func NewOpenAPIProductReconciler(b *reconcilers.BaseReconciler,
	openapiCR *capabilitiesv1beta1.OpenAPI,
	openapiObj *openapi3.T,
	providerAccount *controllerhelper.ProviderAccount,
	logger logr.Logger,
) *OpenAPIProductReconciler {
	return &OpenAPIProductReconciler{
		BaseReconciler:  b,
		openapiCR:       openapiCR,
		openapiObj:      openapiObj,
		providerAccount: providerAccount,
		logger:          logger,
	}
}

func (p *OpenAPIProductReconciler) Logger() logr.Logger {
	return p.logger
}

func (p *OpenAPIProductReconciler) Reconcile() (*capabilitiesv1beta1.Product, error) {
	desired, err := p.desired()
	if err != nil {
		return nil, err
	}

	if p.Logger().V(1).Enabled() {
		jsonData, err := json.MarshalIndent(desired, "", "  ")
		if err != nil {
			return nil, err
		}
		p.Logger().V(1).Info(string(jsonData))
	}

	return nil, p.ReconcileResource(&capabilitiesv1beta1.Product{}, desired, p.productMutator)
}

func (p *OpenAPIProductReconciler) desired() (*capabilitiesv1beta1.Product, error) {
	fieldErrors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	openapiRefFldPath := specFldPath.Child("openapiRef")

	// product obj name
	objName := p.desiredObjName()

	// DNS Subdomain Names
	// If the name would be part of some label, validation would be DNS Label Names (validation.IsDNS1123Label)
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
	errStrings := validation.IsDNS1123Subdomain(objName)
	if len(errStrings) > 0 {
		fieldErrors = append(fieldErrors, field.Invalid(openapiRefFldPath, p.openapiCR.Spec.OpenAPIRef, strings.Join(errStrings, ",")))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	// product name
	name := p.openapiObj.Info.Title

	// product system name
	systemName := p.desiredSystemName()

	// product description
	description := fmt.Sprintf(p.openapiObj.Info.Description)

	insecureSkipVerify := controllerhelper.GetInsecureSkipVerifyAnnotation(p.openapiCR.GetAnnotations())

	product := &capabilitiesv1beta1.Product{
		TypeMeta: metav1.TypeMeta{
			Kind:       capabilitiesv1beta1.ProductKind,
			APIVersion: capabilitiesv1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      objName,
			Namespace: p.openapiCR.Namespace,
			Annotations: map[string]string{
				"insecure_skip_verify": strconv.FormatBool(insecureSkipVerify),
			},
		},
		Spec: capabilitiesv1beta1.ProductSpec{
			Name:               name,
			SystemName:         systemName,
			Description:        description,
			ProviderAccountRef: p.openapiCR.Spec.ProviderAccountRef,
		},
	}

	// Deployment
	product.Spec.Deployment = p.desiredDeployment()

	// Methods
	product.Spec.Methods = p.desiredMethods()

	// Mapping rules
	mappingRules, err := p.desiredMappingRules()
	if err != nil {
		return nil, err
	}
	product.Spec.MappingRules = mappingRules

	// Metrics
	metrics, err := p.desiredMetrics()
	if err != nil {
		return nil, err
	}
	if metrics != nil && len(metrics) > 0 {
		product.Spec.Metrics = metrics
	}

	// Policies
	policies, err := p.desiredPolicies()
	if err != nil {
		return nil, err
	}
	if policies != nil && len(policies) > 0 {
		product.Spec.Policies = policies
	}

	// Application plans
	applicationPlans, err := p.desiredApplicationPlans()
	if err != nil {
		return nil, err
	}
	if applicationPlans != nil && len(applicationPlans) > 0 {
		product.Spec.ApplicationPlans = applicationPlans
	}

	// backend usages
	// current implementation assumes same system name for backend and product
	backendSystemName := p.desiredSystemName()
	product.Spec.BackendUsages = map[string]capabilitiesv1beta1.BackendUsageSpec{
		backendSystemName: {
			Path: "/",
		},
	}

	product.SetDefaults(p.Logger())

	// internal validation
	validationErrors := product.Validate()
	if len(validationErrors) > 0 {
		return nil, errors.New(validationErrors.ToAggregate().Error())
	}

	err = p.SetControllerOwnerReference(p.openapiCR, product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (p *OpenAPIProductReconciler) productMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*capabilitiesv1beta1.Product)
	if !ok {
		return false, fmt.Errorf("%T is not a *capabilitiesv1beta1.Product", existingObj)
	}
	desired, ok := desiredObj.(*capabilitiesv1beta1.Product)
	if !ok {
		return false, fmt.Errorf("%T is not a *capabilitiesv1beta1.Product", desiredObj)
	}

	// Metadata labels and annotations
	updated := helper.EnsureObjectMeta(existing, desired)

	// OwnerRefenrence
	updatedTmp, err := p.EnsureOwnerReference(p.openapiCR, existing)
	if err != nil {
		return false, err
	}
	updated = updated || updatedTmp

	// Maybe too rough compare method?
	// What if product controller adds or modifies something?
	// the openapi controller will be reconciliating.
	// maybe compare only "managed" fields
	if !reflect.DeepEqual(existing.Spec, desired.Spec) {
		diff := cmp.Diff(existing.Spec, desired.Spec)
		p.Logger().Info(fmt.Sprintf("%s spec has changed: %s", common.ObjectInfo(desired), diff))
		existing.Spec = desired.Spec
		updated = true
	}

	return updated, nil
}

func (p *OpenAPIProductReconciler) desiredSystemName() string {
	// Same as backend system name
	// Duplicated implementation. Refactor
	if p.openapiCR.Spec.ProductSystemName != nil {
		return *p.openapiCR.Spec.ProductSystemName
	}

	return helper.SystemNameFromOpenAPITitle(p.openapiObj)
}

func (p *OpenAPIProductReconciler) desiredObjName() string {
	// DNS1123 Label compliant name. Due to UIDs are 36 characters of length this
	// means that the maximum prefix lenght that can be provided is of 26
	// characters. If the generated name is not DNS1123 compliant an error is
	// returned
	// Maybe truncate?
	return fmt.Sprintf("%s-%s", helper.K8sNameFromOpenAPITitle(p.openapiObj), string(p.openapiCR.UID))
}

func (p *OpenAPIProductReconciler) desiredDeployment() *capabilitiesv1beta1.ProductDeploymentSpec {
	deployment := &capabilitiesv1beta1.ProductDeploymentSpec{}

	if p.openapiCR.Spec.ProductionPublicBaseURL != nil || p.openapiCR.Spec.StagingPublicBaseURL != nil {
		// Self managed deployment
		deployment.ApicastSelfManaged = &capabilitiesv1beta1.ApicastSelfManagedSpec{
			StagingPublicBaseURL:    p.openapiCR.Spec.StagingPublicBaseURL,
			ProductionPublicBaseURL: p.openapiCR.Spec.ProductionPublicBaseURL,
			Authentication:          p.desiredAuthentication(),
		}
	} else {
		// Hosted deployment
		deployment.ApicastHosted = &capabilitiesv1beta1.ApicastHostedSpec{
			Authentication: p.desiredAuthentication(),
		}
	}

	return deployment
}

func (p *OpenAPIProductReconciler) desiredAuthentication() *capabilitiesv1beta1.AuthenticationSpec {
	globalSecRequirements := helper.OpenAPIGlobalSecurityRequirements(p.openapiObj)
	if len(globalSecRequirements) == 0 {
		// if no security requirements are found, default to UserKey auth
		return p.desiredUserKeyAuthentication(nil)
	}

	// Only the first one is used
	secRequirementExtended := globalSecRequirements[0]

	var authenticationSpec *capabilitiesv1beta1.AuthenticationSpec

	switch secRequirementExtended.Value.Type {
	case "apiKey":
		authenticationSpec = p.desiredUserKeyAuthentication(secRequirementExtended)
	case "oauth2":
		authenticationSpec = p.desiredOIDCAuthentication(secRequirementExtended)
	case "openIdConnect":
		authenticationSpec = p.desiredOIDCAuthentication(secRequirementExtended)
	}

	if authenticationSpec == nil {
		return p.desiredUserKeyAuthentication(nil)
	}

	return authenticationSpec
}

func (p *OpenAPIProductReconciler) desiredUserKeyAuthentication(secReq *helper.ExtendedSecurityRequirement) *capabilitiesv1beta1.AuthenticationSpec {
	authSpec := &capabilitiesv1beta1.AuthenticationSpec{
		UserKeyAuthentication: &capabilitiesv1beta1.UserKeyAuthenticationSpec{
			Security: p.desiredPrivateAPISecurity(),
		},
	}

	if secReq != nil {
		authSpec.UserKeyAuthentication.Key = &secReq.Value.Name
		authSpec.UserKeyAuthentication.CredentialsLoc = p.parseUserKeyCredentialsLoc(secReq.Value.In)
	}

	return authSpec
}

func (p *OpenAPIProductReconciler) parseUserKeyCredentialsLoc(inField string) *string {
	tmpQuery := "query"
	tmpHeaders := "headers"
	switch inField {
	case "query":
		return &tmpQuery
	case "header":
		return &tmpHeaders
	default:
		return nil
	}
}

func (p *OpenAPIProductReconciler) desiredMethods() map[string]capabilitiesv1beta1.MethodSpec {
	methods := make(map[string]capabilitiesv1beta1.MethodSpec)
	for path, pathItem := range p.openapiObj.Paths {
		for opVerb, operation := range pathItem.Operations() {
			methodSystemName := helper.MethodSystemNameFromOpenAPIOperation(path, opVerb, operation)
			methods[methodSystemName] = capabilitiesv1beta1.MethodSpec{
				Name:        helper.MethodNameFromOpenAPIOperation(path, opVerb, operation),
				Description: operation.Description,
			}
		}
	}
	return methods
}

func (p *OpenAPIProductReconciler) desiredMappingRules() ([]capabilitiesv1beta1.MappingRuleSpec, error) {
	mappingRules := make([]capabilitiesv1beta1.MappingRuleSpec, 0)
	for path, pathItem := range p.openapiObj.Paths {
		desiredPattern, err := p.desiredMappingRulesPattern(path)
		if err != nil {
			return nil, err
		}

		for opVerb, operation := range pathItem.Operations() {
			mappingRule := capabilitiesv1beta1.MappingRuleSpec{
				HTTPMethod:      strings.ToUpper(opVerb),
				Pattern:         desiredPattern,
				MetricMethodRef: helper.MethodSystemNameFromOpenAPIOperation(path, opVerb, operation),
				Increment:       1,
			}

			// Extract OAS operation extension
			operationExtension, err := helper.NewOasOperationExtension(operation)
			if err != nil {
				return nil, err
			}
			if operationExtension != nil {
				if operationExtension.MappingRule.MetricMethodRef != "" {
					mappingRule.MetricMethodRef = operationExtension.MappingRule.MetricMethodRef
				}
				if operationExtension.MappingRule.Increment != 0 {
					mappingRule.Increment = operationExtension.MappingRule.Increment
				}
				if operationExtension.MappingRule.Last != nil {
					mappingRule.Last = operationExtension.MappingRule.Last
				}
			}

			mappingRules = append(mappingRules, mappingRule)
		}
	}
	return mappingRules, nil
}

func (p *OpenAPIProductReconciler) desiredMappingRulesPattern(path string) (string, error) {
	publicBasePath, err := p.desiredPublicBasePath()
	if err != nil {
		return "", err
	}

	// remove the last slash of the publicBasePath
	publicBasePathSanitized := LastSlashRegexp.ReplaceAllString(publicBasePath, "")

	//  According OAS 3.0: path MUST begin with a slash
	pattern := fmt.Sprintf("%s%s", publicBasePathSanitized, path)

	if p.openapiCR.Spec.PrefixMatching == nil || !*p.openapiCR.Spec.PrefixMatching {
		pattern = fmt.Sprintf("%s$", pattern)
	}

	return pattern, nil
}

func (p *OpenAPIProductReconciler) desiredMetrics() (map[string]capabilitiesv1beta1.MetricSpec, error) {
	metrics := make(map[string]capabilitiesv1beta1.MetricSpec)

	// Extract OAS product extension
	rootProductExtension, err := helper.NewOasRootProductExtension(p.openapiObj)
	if err != nil {
		return nil, err
	}

	if rootProductExtension != nil && rootProductExtension.Metrics != nil {
		// Loop through metrics in extension and create Metrics
		for metricKey, metricObj := range rootProductExtension.Metrics {
			metric := capabilitiesv1beta1.MetricSpec{
				Name:        metricObj.Name,
				Unit:        metricObj.Unit,
				Description: metricObj.Description,
			}
			metrics[metricKey] = metric
		}
	}
	return metrics, nil
}

func (p *OpenAPIProductReconciler) desiredApplicationPlans() (map[string]capabilitiesv1beta1.ApplicationPlanSpec, error) {
	applicationPlans := make(map[string]capabilitiesv1beta1.ApplicationPlanSpec)

	// Extract OAS product extension
	rootProductExtension, err := helper.NewOasRootProductExtension(p.openapiObj)
	if err != nil {
		return nil, err
	}

	if rootProductExtension != nil && rootProductExtension.ApplicationPlans != nil {
		// Loop through application plans in extension and create ApplicationPlans
		for appPlanKey, appPlanObj := range rootProductExtension.ApplicationPlans {
			appPlan := capabilitiesv1beta1.ApplicationPlanSpec{
				Name:                appPlanObj.Name,
				AppsRequireApproval: appPlanObj.AppsRequireApproval,
				TrialPeriod:         appPlanObj.TrialPeriod,
				SetupFee:            appPlanObj.SetupFee,
				CostMonth:           appPlanObj.CostMonth,
				PricingRules:        appPlanObj.PricingRules,
				Limits:              appPlanObj.Limits,
				Published:           appPlanObj.Published,
			}
			applicationPlans[appPlanKey] = appPlan
		}
	}

	return applicationPlans, nil
}

func (p *OpenAPIProductReconciler) desiredPolicies() ([]capabilitiesv1beta1.PolicyConfig, error) {
	var policyConfigs []capabilitiesv1beta1.PolicyConfig

	// Extract OAS product extension
	rootProductExtension, err := helper.NewOasRootProductExtension(p.openapiObj)
	if err != nil {
		return nil, err
	}

	if rootProductExtension != nil && rootProductExtension.Policies != nil {
		// Loop through policies in extension and create PolicyConfigs
		for _, policy := range rootProductExtension.Policies {
			policyConfig := capabilitiesv1beta1.PolicyConfig{
				Name:             policy.Name,
				Version:          policy.Version,
				Enabled:          policy.Enabled,
				Configuration:    policy.Configuration,
				ConfigurationRef: policy.ConfigurationRef,
			}
			policyConfigs = append(policyConfigs, policyConfig)
		}
	}

	return policyConfigs, nil
}

func (p *OpenAPIProductReconciler) desiredPublicBasePath() (string, error) {
	// TODO Override public base path optional param

	basePath, err := helper.BasePathFromOpenAPI(p.openapiObj)
	if err != nil {
		fieldErrors := field.ErrorList{}
		specFldPath := field.NewPath("spec")
		openapiRefFldPath := specFldPath.Child("openapiRef")
		fieldErrors = append(fieldErrors, field.Invalid(openapiRefFldPath, p.openapiCR.Spec.OpenAPIRef, err.Error()))
		return "", &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	if basePath == "" {
		return "/", nil
	}

	return basePath, nil
}

func (p *OpenAPIProductReconciler) desiredPrivateAPISecurity() *capabilitiesv1beta1.SecuritySpec {
	if p.openapiCR.Spec.PrivateAPIHostHeader == nil && p.openapiCR.Spec.PrivateAPISecretToken == nil {
		return nil
	}

	privateAPISec := &capabilitiesv1beta1.SecuritySpec{}

	if p.openapiCR.Spec.PrivateAPIHostHeader != nil {
		privateAPISec.HostHeader = p.openapiCR.Spec.PrivateAPIHostHeader
	}

	if p.openapiCR.Spec.PrivateAPISecretToken != nil {
		privateAPISec.SecretToken = p.openapiCR.Spec.PrivateAPISecretToken
	}

	return privateAPISec
}

func (p *OpenAPIProductReconciler) desiredOIDCAuthentication(secReq *helper.ExtendedSecurityRequirement) *capabilitiesv1beta1.AuthenticationSpec {
	if p.openapiCR.Spec.OIDC == nil {
		return nil
	}
	tmpHeaders := "headers"

	authSpec := &capabilitiesv1beta1.AuthenticationSpec{
		OIDC: &capabilitiesv1beta1.OIDCSpec{
			IssuerType:        p.openapiCR.Spec.OIDC.IssuerType,
			IssuerEndpoint:    p.openapiCR.Spec.OIDC.IssuerEndpoint,
			IssuerEndpointRef: p.openapiCR.Spec.OIDC.IssuerEndpointRef,
			Security:          p.desiredPrivateAPISecurity(),
			AuthenticationFlow: &capabilitiesv1beta1.OIDCAuthenticationFlowSpec{
				StandardFlowEnabled:       false,
				ImplicitFlowEnabled:       false,
				DirectAccessGrantsEnabled: false,
				ServiceAccountsEnabled:    false,
			},
			JwtClaimWithClientID:     p.openapiCR.Spec.OIDC.JwtClaimWithClientID,
			JwtClaimWithClientIDType: p.openapiCR.Spec.OIDC.JwtClaimWithClientIDType,
			CredentialsLoc:           &tmpHeaders,
			GatewayResponse:          p.openapiCR.Spec.OIDC.GatewayResponse,
		},
	}

	if secReq.Value.Type == "openIdConnect" {
		p.setOIDCAuthenticationParams(authSpec)
	} else { // oauth2
		p.setOauth2AuthenticationParams(authSpec, secReq)
	}

	return authSpec
}

func (p *OpenAPIProductReconciler) setOIDCAuthenticationParams(authSpec *capabilitiesv1beta1.AuthenticationSpec) {
	if p.openapiCR.Spec.OIDC.AuthenticationFlow != nil && authSpec != nil && authSpec.OIDC != nil && authSpec.OIDC.AuthenticationFlow != nil {
		authSpec.OIDC.AuthenticationFlow.StandardFlowEnabled = p.openapiCR.Spec.OIDC.AuthenticationFlow.StandardFlowEnabled
		authSpec.OIDC.AuthenticationFlow.ImplicitFlowEnabled = p.openapiCR.Spec.OIDC.AuthenticationFlow.ImplicitFlowEnabled
		authSpec.OIDC.AuthenticationFlow.DirectAccessGrantsEnabled = p.openapiCR.Spec.OIDC.AuthenticationFlow.DirectAccessGrantsEnabled
		authSpec.OIDC.AuthenticationFlow.ServiceAccountsEnabled = p.openapiCR.Spec.OIDC.AuthenticationFlow.ServiceAccountsEnabled
	}
}

func (p *OpenAPIProductReconciler) setOauth2AuthenticationParams(authSpec *capabilitiesv1beta1.AuthenticationSpec, secReq *helper.ExtendedSecurityRequirement) {
	if authSpec != nil && authSpec.OIDC != nil && authSpec.OIDC.AuthenticationFlow != nil {
		if secReq.Value.Flows.AuthorizationCode != nil {
			authSpec.OIDC.AuthenticationFlow.StandardFlowEnabled = true
		}
		if secReq.Value.Flows.Implicit != nil {
			authSpec.OIDC.AuthenticationFlow.ImplicitFlowEnabled = true
		}
		if secReq.Value.Flows.Password != nil {
			authSpec.OIDC.AuthenticationFlow.DirectAccessGrantsEnabled = true
		}
		if secReq.Value.Flows.ClientCredentials != nil {
			authSpec.OIDC.AuthenticationFlow.ServiceAccountsEnabled = true
		}
	}
}

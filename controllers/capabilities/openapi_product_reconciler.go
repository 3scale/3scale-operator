package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/google/go-cmp/cmp"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var (
	// LastSlashRegexp matches the last slash
	LastSlashRegexp = regexp.MustCompile(`/$`)
)

type OpenAPIProductReconciler struct {
	*reconcilers.BaseReconciler
	openapiCR       *capabilitiesv1beta1.OpenAPI
	openapiObj      *openapi3.Swagger
	providerAccount *controllerhelper.ProviderAccount
	logger          logr.Logger
}

func NewOpenAPIProductReconciler(b *reconcilers.BaseReconciler,
	openapiCR *capabilitiesv1beta1.OpenAPI,
	openapiObj *openapi3.Swagger,
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

	product := &capabilitiesv1beta1.Product{
		TypeMeta: metav1.TypeMeta{
			Kind:       capabilitiesv1beta1.ProductKind,
			APIVersion: capabilitiesv1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      objName,
			Namespace: p.openapiCR.Namespace,
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

	// backend usages
	// current implementation assumes same system name for backend and product
	backendSystemName := p.desiredSystemName()
	product.Spec.BackendUsages = map[string]capabilitiesv1beta1.BackendUsageSpec{
		backendSystemName: capabilitiesv1beta1.BackendUsageSpec{
			Path: "/",
		},
	}

	product.SetDefaults(p.Logger())

	// internal validation
	validationErrors := product.Validate()
	if len(validationErrors) > 0 {
		return nil, errors.New(validationErrors.ToAggregate().Error())
	}

	err = p.SetOwnerReference(p.openapiCR, product)
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
	// TODO types "oauth2", "openIdConnect"
	case "apiKey":
		authenticationSpec = p.desiredUserKeyAuthentication(secRequirementExtended)
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
			mappingRules = append(mappingRules, capabilitiesv1beta1.MappingRuleSpec{
				HTTPMethod:      strings.ToUpper(opVerb),
				Pattern:         desiredPattern,
				MetricMethodRef: helper.MethodSystemNameFromOpenAPIOperation(path, opVerb, operation),
				Increment:       1,
			})
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

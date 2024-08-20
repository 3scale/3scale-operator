/*
Copyright 2020 Red Hat.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"net/url"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	"github.com/getkin/kin-openapi/openapi3"
)

const (
	oasSecretLabelSelectorKey   = "apimanager.apps.3scale.net/watched-by"
	oasSecretLabelSelectorValue = "openapi"
	openAPISecretRefLabelKey    = "apimanager.apps.3scale.net/oas-source-secret-uid"
)

// OpenAPIReconciler reconciles a OpenAPI object
type OpenAPIReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that OpenAPIReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &OpenAPIReconciler{}

// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=openapis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=openapis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=openapis/finalizers,verbs=get;list;watch;create;update;patch;delete

func (r *OpenAPIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Logger().WithValues("openapi", req.NamespacedName)
	reqLogger.Info("Reconcile OpenAPI", "Operator version", version.Version)

	// Fetch the OpenAPI instance
	openapiCR := &capabilitiesv1beta1.OpenAPI{}
	err := r.Client().Get(context.TODO(), req.NamespacedName, openapiCR)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("resource not found. Ignoring since object must have been deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(openapiCR, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Ignore deleted OpenAPI, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if openapiCR.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	if openapiCR.SetDefaults(reqLogger) {
		err := r.Client().Update(r.Context(), openapiCR)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Failed setting openapi defaults: %w", err)
		}

		reqLogger.Info("resource defaults updated. Requeueing.")
		return ctrl.Result{Requeue: true}, nil
	}

	statusReconciler, reconcileStatus, reconcileErr := r.reconcileSpec(openapiCR)
	statusResult, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		if reconcileErr != nil {
			return ctrl.Result{}, fmt.Errorf("Failed to reconcile openapi: %v. Failed to update openapi status: %w", reconcileErr, statusUpdateErr)
		}

		return ctrl.Result{}, fmt.Errorf("Failed to update openapi status: %w", statusUpdateErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if reconcileErr != nil {
		if helper.IsInvalidSpecError(reconcileErr) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(openapiCR, corev1.EventTypeWarning, "Invalid OpenAPI Spec", "%v", reconcileErr)
			return ctrl.Result{}, nil
		}

		reqLogger.Error(reconcileErr, "Failed to reconcile")
		r.EventRecorder().Eventf(openapiCR, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
		return ctrl.Result{}, reconcileErr
	}

	return reconcileStatus, nil
}

func (r *OpenAPIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	secretToOpenAPIEventMapper := &SecretToOpenAPIEventMapper{
		Context:   r.Context(),
		K8sClient: r.Client(),
		Logger:    r.Logger().WithName("secretToOpenAPIEventMapper"),
	}

	oasSecretLabelSelectorPredicate, err := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchLabels: map[string]string{
			oasSecretLabelSelectorKey: oasSecretLabelSelectorValue,
		},
	})
	if err != nil {
		return nil
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.OpenAPI{}).
		Owns(&capabilitiesv1beta1.Product{}).
		Owns(&capabilitiesv1beta1.Backend{}).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(secretToOpenAPIEventMapper.Map), builder.WatchesOption(builder.WithPredicates(oasSecretLabelSelectorPredicate))).
		Complete(r)
}

func (r *OpenAPIReconciler) reconcileSpec(openapiCR *capabilitiesv1beta1.OpenAPI) (*OpenAPIStatusReconciler, ctrl.Result, error) {
	logger := r.Logger().WithValues("openapi", openapiCR.Name)

	err := r.validateSpec(openapiCR)
	if err != nil {
		statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, "", err, false)
		return statusReconciler, ctrl.Result{}, err
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), openapiCR.Namespace, openapiCR.Spec.ProviderAccountRef, logger)
	if err != nil {
		statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, "", err, false)
		return statusReconciler, ctrl.Result{}, err
	}

	// Sets ownerReference on OpenAPI CR if tenantCR exists so that if tenant is deleted, OpenAPI CR is deleted as well
	// Retrieve ownersReference of tenant CR that owns the Backend CR
	tenantCR, err := controllerhelper.RetrieveTenantCR(providerAccount, r.Client(), r.Logger(), openapiCR.Namespace)
	if err != nil {
		statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, "", err, false)
		return statusReconciler, ctrl.Result{}, err
	}

	// If tenant CR is found, set it's ownersReference as ownerReference in the OpenAPI CR
	if tenantCR != nil {
		updated, err := r.EnsureOwnerReference(tenantCR, openapiCR)
		if err != nil {
			statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, "", err, false)
			return statusReconciler, ctrl.Result{}, err
		}

		if updated {
			err := r.Client().Update(r.Context(), openapiCR)
			if err != nil {
				statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, "", err, false)
				return statusReconciler, ctrl.Result{}, err
			}
			statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, providerAccount.AdminURLStr, err, false)
			return statusReconciler, ctrl.Result{Requeue: true}, err
		}
	}

	openapiObj, err := r.readOpenAPI(openapiCR)
	if err != nil {
		statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, providerAccount.AdminURLStr, err, false)
		return statusReconciler, ctrl.Result{}, err
	}

	err = r.validateOpenAPIAs3scaleProduct(openapiCR, openapiObj)
	if err != nil {
		statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, providerAccount.AdminURLStr, err, false)
		return statusReconciler, ctrl.Result{}, err
	}

	err = r.validateOIDCSettingsInCR(openapiCR, openapiObj)
	if err != nil {
		statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, "", err, false)
		return statusReconciler, ctrl.Result{}, err
	}

	err = r.validateOASExtensions(openapiObj)
	if err != nil {
		statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, providerAccount.AdminURLStr, err, false)
		return statusReconciler, ctrl.Result{}, err
	}

	backendReconciler := NewOpenAPIBackendReconciler(r.BaseReconciler, openapiCR, openapiObj, providerAccount, logger)
	_, err = backendReconciler.Reconcile()
	if err != nil {
		statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, providerAccount.AdminURLStr, err, false)
		return statusReconciler, ctrl.Result{}, err
	}

	productReconciler := NewOpenAPIProductReconciler(r.BaseReconciler, openapiCR, openapiObj, providerAccount, logger)
	_, err = productReconciler.Reconcile()
	if err != nil {
		statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, providerAccount.AdminURLStr, err, false)
		return statusReconciler, ctrl.Result{}, err
	}

	// No need to check for backend sync state.
	// The product has the backends linked as backend usage.
	// The product will not be in sync until the backend usage items are sync'ed.
	// The product controller makes sure the backend usage's items are valid Backend CRs and are sync'ed.
	productSynced, err := r.checkProductSynced(openapiCR)
	if err != nil {
		statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, providerAccount.AdminURLStr, err, false)
		return statusReconciler, ctrl.Result{}, err
	}

	statusReconciler := NewOpenAPIStatusReconciler(r.BaseReconciler, openapiCR, providerAccount.AdminURLStr, err, productSynced)

	// If the product is successfully synced AND the OpenAPI CR is using URL ref, then requeue after 5 minutes
	// We have to requeue like this in case there were updates to the URL source because we can't watch the URL directly
	if productSynced && openapiCR.Spec.OpenAPIRef.SecretRef == nil {
		return statusReconciler, ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Minute}, err
	}

	return statusReconciler, ctrl.Result{Requeue: !productSynced}, err
}

func (r *OpenAPIReconciler) validateSpec(resource *capabilitiesv1beta1.OpenAPI) error {
	errors := field.ErrorList{}
	errors = append(errors, resource.Validate()...)

	if len(errors) == 0 {
		return nil
	}

	return &helper.SpecFieldError{
		ErrorType:      helper.InvalidError,
		FieldErrorList: errors,
	}
}

func (r *OpenAPIReconciler) checkProductSynced(resource *capabilitiesv1beta1.OpenAPI) (bool, error) {
	if resource.Status.ProductResourceName == nil {
		// product resource name not available to check
		return false, nil
	}

	// Fetch the Product instance
	product := &capabilitiesv1beta1.Product{}
	objectKey := client.ObjectKey{Name: resource.Status.ProductResourceName.Name, Namespace: resource.Namespace}
	err := r.Client().Get(r.Context(), objectKey, product)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		// Error reading the object - requeue the request.
		return false, err
	}

	return product.Status.Conditions.IsTrueFor(capabilitiesv1beta1.ProductSyncedConditionType), nil
}

func (r *OpenAPIReconciler) readOpenAPI(resource *capabilitiesv1beta1.OpenAPI) (*openapi3.T, error) {
	// OpenAPIRef is oneOf by CRD openapiV3 validation
	if resource.Spec.OpenAPIRef.SecretRef != nil {
		// Label the OAS source secret and OpenAPI so the secret can be watched by the openapi_controller
		err := r.labelOpenAPISecretAndCR(resource)
		if err != nil {
			return nil, err
		}

		return r.readOpenAPISecret(resource)
	}

	// Must be URL
	return r.readOpenAPIFromURL(resource)
}

func (r *OpenAPIReconciler) labelOpenAPISecretAndCR(openAPICR *capabilitiesv1beta1.OpenAPI) error {
	fieldErrors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	openapiRefFldPath := specFldPath.Child("openapiRef")
	secretRefFldPath := openapiRefFldPath.Child("secretRef")

	oasSourceSecret := &corev1.Secret{}
	objectKey := types.NamespacedName{Name: openAPICR.Spec.OpenAPIRef.SecretRef.Name, Namespace: openAPICR.Spec.OpenAPIRef.SecretRef.Namespace}

	// Read the OAS source secret
	if err := r.Client().Get(r.Context(), objectKey, oasSourceSecret); err != nil {
		if errors.IsNotFound(err) {
			fieldErrors = append(fieldErrors, field.Invalid(secretRefFldPath, openAPICR.Spec.OpenAPIRef.SecretRef, "Secret not found"))
			return &helper.SpecFieldError{
				ErrorType:      helper.InvalidError,
				FieldErrorList: fieldErrors,
			}
		}
		// Unexpected error
		return err
	}

	// Add label to OAS source secret so it can be watched
	if oasSourceSecret.ObjectMeta.Labels == nil {
		oasSourceSecret.ObjectMeta.Labels = map[string]string{}
	}
	oasSourceSecret.ObjectMeta.Labels[oasSecretLabelSelectorKey] = oasSecretLabelSelectorValue
	if err := r.Client().Update(r.Context(), oasSourceSecret); err != nil {
		return err
	}

	// Re-fetch the OpenAPI CR in case it's been modified
	objectKey = types.NamespacedName{Name: openAPICR.Name, Namespace: openAPICR.Namespace}
	if err := r.Client().Get(r.Context(), objectKey, openAPICR); err != nil {
		return err
	}

	// Add label to OpenAPI CR with source secret's UID
	if openAPICR.ObjectMeta.Labels == nil {
		openAPICR.ObjectMeta.Labels = map[string]string{}
	}
	openAPICR.ObjectMeta.Labels[openAPISecretRefLabelKey] = string(oasSourceSecret.GetUID())
	if err := r.Client().Update(r.Context(), openAPICR); err != nil {
		return err
	}

	return nil
}

func (r *OpenAPIReconciler) readOpenAPISecret(resource *capabilitiesv1beta1.OpenAPI) (*openapi3.T, error) {
	fieldErrors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	openapiRefFldPath := specFldPath.Child("openapiRef")
	secretRefFldPath := openapiRefFldPath.Child("secretRef")

	objectKey := types.NamespacedName{Name: resource.Spec.OpenAPIRef.SecretRef.Name, Namespace: resource.Spec.OpenAPIRef.SecretRef.Namespace}
	openapiSecretObj := &corev1.Secret{}

	// Read secret
	if err := r.Client().Get(r.Context(), objectKey, openapiSecretObj); err != nil {
		if errors.IsNotFound(err) {
			fieldErrors = append(fieldErrors, field.Invalid(secretRefFldPath, resource.Spec.OpenAPIRef.SecretRef, "Secret not found"))
			return nil, &helper.SpecFieldError{
				ErrorType:      helper.InvalidError,
				FieldErrorList: fieldErrors,
			}
		}

		// unexpected error
		return nil, err
	}

	if len(openapiSecretObj.Data) != 1 {
		fieldErrors = append(fieldErrors, field.Invalid(secretRefFldPath, resource.Spec.OpenAPIRef.SecretRef, "Secret was empty or contains too many fields. Only one is required."))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	// Get key value
	dataByteArray := func(secret *corev1.Secret) []byte {
		for _, v := range secret.Data {
			return v
		}
		return nil
	}(openapiSecretObj)

	openapiObj, err := openapi3.NewLoader().LoadFromData(dataByteArray)
	if err != nil {
		fieldErrors = append(fieldErrors, field.Invalid(secretRefFldPath, resource.Spec.OpenAPIRef.SecretRef, err.Error()))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	err = openapiObj.Validate(r.Context())
	if err != nil {
		fieldErrors = append(fieldErrors, field.Invalid(secretRefFldPath, resource.Spec.OpenAPIRef.SecretRef, err.Error()))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	return openapiObj, nil
}

func (r *OpenAPIReconciler) validateOpenAPIAs3scaleProduct(openapiCR *capabilitiesv1beta1.OpenAPI, openapiObj *openapi3.T) error {
	fieldErrors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	openapiRefFldPath := specFldPath.Child("openapiRef")

	// Multiple sec requirements
	globalSecRequirements := helper.OpenAPIGlobalSecurityRequirements(openapiObj)
	if len(globalSecRequirements) > 1 {
		fieldErrors = append(fieldErrors, field.Invalid(openapiRefFldPath, openapiCR.Spec.OpenAPIRef, "Invalid OAS: multiple security requirements"))
		return &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	if len(globalSecRequirements) == 1 {
		// Validate supported types
		switch globalSecRequirements[0].Value.Type {
		case "apiKey":
			break
		case "oauth2":
			break
		case "openIdConnect":
			break
		default:
			fieldErrors = append(fieldErrors, field.Invalid(openapiRefFldPath, openapiCR.Spec.OpenAPIRef, fmt.Sprintf("Unexpected security schema type: %s", globalSecRequirements[0].Value.Type)))
			return &helper.SpecFieldError{
				ErrorType:      helper.InvalidError,
				FieldErrorList: fieldErrors,
			}
		}
	}

	return nil
}

func (r *OpenAPIReconciler) validateOASExtensions(openapiObj *openapi3.T) error {
	extensionErrors := field.ErrorList{}
	productExtensionPath := field.NewPath("x-3scale-product")
	metricsExtensionPath := productExtensionPath.Child("metrics")
	policiesExtensionPath := productExtensionPath.Child("policies")

	// Validate OAS root product extension
	rootProductExtension, err := helper.NewOasRootProductExtension(openapiObj)
	if err != nil {
		return err
	}

	// Validate metrics
	if rootProductExtension != nil && rootProductExtension.Metrics != nil {
		// Loop through policies in extension and create PolicyConfig objects
		for metricKey, metric := range rootProductExtension.Metrics {
			if metric.Name == "" {
				extensionErrors = append(extensionErrors, field.Required(metricsExtensionPath, fmt.Sprintf("metric %s is missing a friendlyName", metricKey)))
			}
			if metric.Unit == "" {
				extensionErrors = append(extensionErrors, field.Required(metricsExtensionPath, fmt.Sprintf("metric %s is missing a unit", metricKey)))
			}
		}

	}

	// Validate policies
	if rootProductExtension != nil && rootProductExtension.Policies != nil {
		// Loop through policies in extension and create PolicyConfig objects
		for _, policy := range rootProductExtension.Policies {
			if policy.Name == "" {
				extensionErrors = append(extensionErrors, field.Required(policiesExtensionPath, "one or more policies are missing a name"))
			}
			if policy.Version == "" {
				extensionErrors = append(extensionErrors, field.Required(policiesExtensionPath, fmt.Sprintf("policy %s is missing a version", policy.Name)))
			}
		}
	}

	if len(extensionErrors) == 0 {
		return nil
	}

	return &helper.SpecFieldError{
		ErrorType:      helper.InvalidError,
		FieldErrorList: extensionErrors,
	}
}

func (r *OpenAPIReconciler) readOpenAPIFromURL(resource *capabilitiesv1beta1.OpenAPI) (*openapi3.T, error) {
	fieldErrors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	openapiRefFldPath := specFldPath.Child("openapiRef")
	urlRefFldPath := openapiRefFldPath.Child("url")

	openAPIURL, err := url.Parse(*resource.Spec.OpenAPIRef.URL)
	if err != nil {
		fieldErrors = append(fieldErrors, field.Invalid(urlRefFldPath, resource.Spec.OpenAPIRef.URL, err.Error()))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	// We need to add a cache-busting query parameter to the URL to ensure we are getting the latest version of the OAS source
	// The openapi3 library will otherwise cache the previous version of the OAS source by default
	openAPIURL.RawQuery = fmt.Sprintf("t=%d", time.Now().Unix())

	openapiObj, err := openapi3.NewLoader().LoadFromURI(openAPIURL)
	if err != nil {
		fieldErrors = append(fieldErrors, field.Invalid(urlRefFldPath, resource.Spec.OpenAPIRef.URL, err.Error()))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	err = openapiObj.Validate(r.Context())
	if err != nil {
		fieldErrors = append(fieldErrors, field.Invalid(urlRefFldPath, resource.Spec.OpenAPIRef.URL, err.Error()))
		return nil, &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	return openapiObj, nil
}

func (r *OpenAPIReconciler) validateOIDCSettingsInCR(openapiCR *capabilitiesv1beta1.OpenAPI, openapiObj *openapi3.T) error {
	logger := r.Logger().WithValues("openapi", openapiCR.Name)
	fieldErrors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	openapiOidcFldPath := specFldPath.Child("oidc")
	openapiIssuerEndpointFldPath := openapiOidcFldPath.Child("IssuerEndpoint")
	openapiIssuerEndpointRefFldPath := openapiOidcFldPath.Child("IssuerEndpointRef")

	if openapiCR.Spec.OIDC != nil &&
		(openapiCR.Spec.OIDC.IssuerEndpoint == "" && openapiCR.Spec.OIDC.IssuerEndpointRef == nil) {
		fieldErrors = append(fieldErrors, field.Invalid(openapiIssuerEndpointFldPath, openapiCR.Spec.OIDC.IssuerEndpoint, "OIDC IssuerEndpoint definition is missing in CR."))
		fieldErrors = append(fieldErrors, field.Invalid(openapiIssuerEndpointRefFldPath, openapiCR.Spec.OIDC.IssuerEndpointRef, "OIDC IssuerEndpointRef definition is missing in CR. "+
			"No IssuerEndpoint nor IssuerEndpointRef found in OIDC spec in CR, one of them must be set."))
		return &helper.SpecFieldError{
			ErrorType:      helper.InvalidError,
			FieldErrorList: fieldErrors,
		}
	}

	globalSecRequirements := helper.OpenAPIGlobalSecurityRequirements(openapiObj)
	if len(globalSecRequirements) == 0 && openapiCR.Spec.OIDC != nil {
		logger.Info("OIDC definitions in CR will be ignored, as no security requirements are found. Default to UserKey authentication")
		r.EventRecorder().Eventf(openapiCR, corev1.EventTypeWarning, "No security requirements are found in OAS", "%v", "OIDC definitions in CR will be ignored, as no security requirements are found. Default to UserKey authentication")
	}

	if len(globalSecRequirements) == 1 {
		// when the referenced OpenAPI spec's sec scheme is openIdConnect or oauth2, the spec.oidc must not be nil or empty
		if globalSecRequirements[0].Value.Type == "openIdConnect" || globalSecRequirements[0].Value.Type == "oauth2" {
			if openapiCR.Spec.OIDC == nil {
				fieldErrors = append(fieldErrors, field.Invalid(openapiOidcFldPath, openapiCR.Spec.OIDC, "Missing "+
					"OIDC definitions in CR. The referenced OpenAPI spec's sec scheme is openIdConnect or oauth2, the spec.oidc must not be nil or empty"))
				return &helper.SpecFieldError{
					ErrorType:      helper.InvalidError,
					FieldErrorList: fieldErrors,
				}
			}
		}
		// when OAS securitySchemes type is oauth2, and openapiCR spec is OIDC, then CR OIDC Authentication Flows parameters will be ignored,
		// and Product authentication flows will be set to match oauth2 flows in OAS
		if openapiCR.Spec.OIDC != nil && globalSecRequirements[0].Value.Type == "oauth2" {
			logger.Info("OIDC authentication flows in CR will be ignored and Product OIDC authentication flows will be set to match oauth2 flows in OAS since the SecuritySchemes type in OAS is \"oauth2\" (for OIDC it should be \"openIdConnect\")")
			r.EventRecorder().Eventf(openapiCR, corev1.EventTypeWarning, "OIDC authentication flows in CR will be ignored and Product OIDC authentication flows will be set to match oauth2 flows in OAS since the SecuritySchemes type in OAS is \"oauth2\" (for OIDC it should be \"openIdConnect\")", "%v", "Product OIDC authentication flows parameters will be set to match oauth2 flows as following (OIDC ~ OAuth2): StandardFlowEnabled ~ AuthorizationCode, ImplicitFlowEnabled ~ Implicit, DirectAccessGrantsEnabled ~ Password, ServiceAccountsEnabled ~ ClientCredentials")
		}
	}

	return nil
}

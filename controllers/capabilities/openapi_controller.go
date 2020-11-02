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
	"net/url"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	"github.com/getkin/kin-openapi/openapi3"
)

// OpenAPIReconciler reconciles a OpenAPI object
type OpenAPIReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that OpenAPIReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &OpenAPIReconciler{}

// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=openapis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=openapis/status,verbs=get;update;patch

func (r *OpenAPIReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.OpenAPI{}).
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

func (r *OpenAPIReconciler) readOpenAPI(resource *capabilitiesv1beta1.OpenAPI) (*openapi3.Swagger, error) {
	// OpenAPIRef is oneOf by CRD openapiV3 validation
	if resource.Spec.OpenAPIRef.SecretRef != nil {
		return r.readOpenAPISecret(resource)
	}

	// Must be URL
	return r.readOpenAPIFromURL(resource)
}

func (r *OpenAPIReconciler) readOpenAPISecret(resource *capabilitiesv1beta1.OpenAPI) (*openapi3.Swagger, error) {
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

	openapiObj, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(dataByteArray)
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

func (r *OpenAPIReconciler) validateOpenAPIAs3scaleProduct(openapiCR *capabilitiesv1beta1.OpenAPI, openapiObj *openapi3.Swagger) error {
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

func (r *OpenAPIReconciler) readOpenAPIFromURL(resource *capabilitiesv1beta1.OpenAPI) (*openapi3.Swagger, error) {
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

	openapiObj, err := openapi3.NewSwaggerLoader().LoadSwaggerFromURI(openAPIURL)
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

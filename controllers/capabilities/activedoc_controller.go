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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
)

// ActiveDocReconciler reconciles a ActiveDoc object
type ActiveDocReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that BackendReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &ActiveDocReconciler{}

// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=activedocs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=activedocs/status,verbs=get;update;patch

func (r *ActiveDocReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Logger().WithValues("activedoc", req.NamespacedName)
	reqLogger.Info("Reconcile ActiveDoc", "Operator version", version.Version)

	// Fetch the Backend instance
	activeDocCR := &capabilitiesv1beta1.ActiveDoc{}
	err := r.Client().Get(context.TODO(), req.NamespacedName, activeDocCR)
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
		jsonData, err := json.MarshalIndent(activeDocCR, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Ignore deleted resource, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if activeDocCR.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	if activeDocCR.SetDefaults(reqLogger) {
		err := r.Client().Update(r.Context(), activeDocCR)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed setting activedoc defaults: %w", err)
		}

		reqLogger.Info("resource defaults updated. Requeueing.")
		return ctrl.Result{Requeue: true}, nil
	}

	statusReconciler, reconcileErr := r.reconcileSpec(activeDocCR, reqLogger)
	statusResult, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		if reconcileErr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to reconcile activedoc: %v. Failed to update activedoc status: %w", reconcileErr, statusUpdateErr)
		}

		return ctrl.Result{}, fmt.Errorf("failed to update activedoc status: %w", statusUpdateErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if reconcileErr != nil {
		if helper.IsInvalidSpecError(reconcileErr) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(activeDocCR, corev1.EventTypeWarning, "Invalid ActiveDoc Spec", "%v", reconcileErr)
			return ctrl.Result{}, nil
		}

		if helper.IsOrphanSpecError(reconcileErr) {
			// On Orphan spec error, retry
			reqLogger.Info("ERROR", "spec orphan error", reconcileErr)
			return ctrl.Result{Requeue: true}, nil
		}

		reqLogger.Error(reconcileErr, "Failed to reconcile")
		r.EventRecorder().Eventf(activeDocCR, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
		return ctrl.Result{}, reconcileErr
	}

	return ctrl.Result{}, nil
}

func (r *ActiveDocReconciler) reconcileSpec(activeDocCR *capabilitiesv1beta1.ActiveDoc, logger logr.Logger) (*ActiveDocStatusReconciler, error) {
	err := r.validateSpec(activeDocCR)
	if err != nil {
		statusReconciler := NewActiveDocStatusReconciler(r.BaseReconciler, activeDocCR, "", nil, err)
		return statusReconciler, err
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), activeDocCR.Namespace, activeDocCR.Spec.ProviderAccountRef, logger)
	if err != nil {
		statusReconciler := NewActiveDocStatusReconciler(r.BaseReconciler, activeDocCR, "", nil, err)
		return statusReconciler, err
	}

	err = r.checkExternalRefs(activeDocCR, providerAccount.AdminURLStr, logger)
	if err != nil {
		statusReconciler := NewActiveDocStatusReconciler(r.BaseReconciler, activeDocCR, providerAccount.AdminURLStr, nil, err)
		return statusReconciler, err
	}

	insecureSkipVerify := controllerhelper.GetInsecureSkipVerifyAnnotation(activeDocCR.GetAnnotations())
	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount, insecureSkipVerify)
	if err != nil {
		statusReconciler := NewActiveDocStatusReconciler(r.BaseReconciler, activeDocCR, providerAccount.AdminURLStr, nil, err)
		return statusReconciler, err
	}

	reconciler := NewActiveDocThreescaleReconciler(r.BaseReconciler, activeDocCR, threescaleAPIClient, providerAccount.AdminURLStr, logger)
	activeDocObj, err := reconciler.Reconcile()

	statusReconciler := NewActiveDocStatusReconciler(r.BaseReconciler, activeDocCR, providerAccount.AdminURLStr, activeDocObj, err)
	return statusReconciler, err
}

func (r *ActiveDocReconciler) validateSpec(activeDocCR *capabilitiesv1beta1.ActiveDoc) error {
	errors := field.ErrorList{}
	errors = append(errors, activeDocCR.Validate()...)

	if len(errors) == 0 {
		return nil
	}

	return &helper.SpecFieldError{
		ErrorType:      helper.InvalidError,
		FieldErrorList: errors,
	}
}

func (r *ActiveDocReconciler) checkExternalRefs(resource *capabilitiesv1beta1.ActiveDoc, providerAccountHost string, logger logr.Logger) error {
	errors := field.ErrorList{}

	// Check product referenced by the ActiveDoc Spec is valid
	if resource.Spec.ProductSystemName != nil {
		productList, err := controllerhelper.ProductList(resource.Namespace, r.Client(), providerAccountHost, logger)
		if err != nil {
			return fmt.Errorf("ActiveDocReconciler.checkExternalRefs: %w", err)
		}

		specFldPath := field.NewPath("spec")
		productFldPath := specFldPath.Child("productSystemName")
		idx := controllerhelper.FindProductBySystemName(productList, *resource.Spec.ProductSystemName)
		if idx < 0 {
			errors = append(errors, field.Invalid(productFldPath, resource.Spec.ProductSystemName, "not a valid reference"))
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return &helper.SpecFieldError{
		ErrorType:      helper.OrphanError,
		FieldErrorList: errors,
	}
}

func (r *ActiveDocReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.ActiveDoc{}).
		Complete(r)
}

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

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// BackendReconciler reconciles a Backend object
type BackendReconciler struct {
	*reconcilers.BaseReconciler
}

const backendFinalizer = "backend.capabilities.3scale.net/finalizer"

// blank assignment to verify that BackendReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &BackendReconciler{}

// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=backends,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=backends/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=backends/finalizers,verbs=get;list;watch;create;update;patch;delete

func (r *BackendReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Logger().WithValues("backend", req.NamespacedName)
	reqLogger.Info("Reconcile Backend", "Operator version", version.Version)

	// Fetch the Backend instance
	backend := &capabilitiesv1beta1.Backend{}
	err := r.Client().Get(context.TODO(), req.NamespacedName, backend)
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
		jsonData, err := json.MarshalIndent(backend, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Ignore deleted Backends, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if backend.GetDeletionTimestamp() != nil && controllerutil.ContainsFinalizer(backend, backendFinalizer) {
		err = controllerhelper.ReconcileFinalizers(backend, r.Client(), backendFinalizer)
		if err != nil {
			return ctrl.Result{}, err
		}

		err = r.removeBackend(backend)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Ignore deleted Backends, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if backend.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(backend, backendFinalizer) {
		err = controllerhelper.ReconcileFinalizers(backend, r.Client(), backendFinalizer)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if backend.Spec.ProviderAccountRef != nil {
		providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), backend.GetNamespace(), backend.Spec.ProviderAccountRef, r.Logger())
		if err != nil {
			return ctrl.Result{}, err
		}

		// Retrieve ownersReference of tenant CR that owns the Backend CR
		tenantCR, err := controllerhelper.RetrieveTenantCR(providerAccount, r.Client())
		if err != nil {
			return ctrl.Result{}, err
		}

		// If tenant CR is found, set it's ownersReference as ownerReference in the BackendCR CR
		if tenantCR != nil {
			err := controllerhelper.SetOwnersReference(backend, r.Client(), tenantCR)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	if backend.SetDefaults(reqLogger) {
		err := r.Client().Update(r.Context(), backend)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Failed setting backend defaults: %w", err)
		}

		reqLogger.Info("resource defaults updated. Requeueing.")
		return ctrl.Result{Requeue: true}, nil
	}

	statusReconciler, reconcileErr := r.reconcile(backend)
	statusResult, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		if reconcileErr != nil {
			return ctrl.Result{}, fmt.Errorf("Failed to sync backend: %v. Failed to update backend status: %w", reconcileErr, statusUpdateErr)
		}

		return ctrl.Result{}, fmt.Errorf("Failed to update backend status: %w", statusUpdateErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if reconcileErr != nil {
		if helper.IsInvalidSpecError(reconcileErr) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(backend, corev1.EventTypeWarning, "Invalid Backend Spec", "%v", reconcileErr)
			return ctrl.Result{}, nil
		}

		reqLogger.Error(reconcileErr, "Failed to reconcile")
		r.EventRecorder().Eventf(backend, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
		return ctrl.Result{}, reconcileErr
	}

	reqLogger.Info("END", "error", reconcileErr)
	return ctrl.Result{}, nil
}

func (r *BackendReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.Backend{}).
		Complete(r)
}

func (r *BackendReconciler) reconcile(backendResource *capabilitiesv1beta1.Backend) (*BackendStatusReconciler, error) {
	logger := r.Logger().WithValues("backend", backendResource.Name)

	err := r.validateSpec(backendResource)
	if err != nil {
		statusReconciler := NewBackendStatusReconciler(r.BaseReconciler, backendResource, nil, "", err)
		return statusReconciler, err
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), backendResource.Namespace, backendResource.Spec.ProviderAccountRef, logger)
	if err != nil {
		statusReconciler := NewBackendStatusReconciler(r.BaseReconciler, backendResource, nil, "", err)
		return statusReconciler, err
	}

	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		statusReconciler := NewBackendStatusReconciler(r.BaseReconciler, backendResource, nil, providerAccount.AdminURLStr, err)
		return statusReconciler, err
	}

	backendRemoteIndex, err := controllerhelper.NewBackendAPIRemoteIndex(threescaleAPIClient, logger)
	if err != nil {
		statusReconciler := NewBackendStatusReconciler(r.BaseReconciler, backendResource, nil, providerAccount.AdminURLStr, err)
		return statusReconciler, err
	}

	reconciler := NewThreescaleReconciler(r.BaseReconciler, backendResource, threescaleAPIClient, backendRemoteIndex, providerAccount)
	backendAPIEntity, err := reconciler.Reconcile()
	statusReconciler := NewBackendStatusReconciler(r.BaseReconciler, backendResource, backendAPIEntity, providerAccount.AdminURLStr, err)
	return statusReconciler, err
}

func (r *BackendReconciler) validateSpec(backendResource *capabilitiesv1beta1.Backend) error {
	errors := field.ErrorList{}
	// internal validation
	errors = append(errors, backendResource.Validate()...)

	if len(errors) == 0 {
		return nil
	}

	return &helper.SpecFieldError{
		ErrorType:      helper.InvalidError,
		FieldErrorList: errors,
	}
}

func (r *BackendReconciler) removeBackend(backend *capabilitiesv1beta1.Backend) error {
	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), backend.Namespace, backend.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		return err
	}

	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		return err
	}

	// Retrieve tenant's products in 3scale that belong to a particular tenant
	// This is required to loop through the tenant products to retrieve their IDs and then, pull backendUsages for each
	productsIn3scale, err := threescaleAPIClient.ListProducts()
	if err != nil {
		return err
	}

	// Retrieve all product CRs
	// This is required to pull CRs that have the backendAPI ID mentioned and to remove the backendUsage mentions
	productCRsList := &capabilitiesv1beta1.ProductList{}
	err = r.Client().List(context.TODO(), productCRsList)
	if err != nil {
		return err
	}

	for _, productIn3scale := range productsIn3scale.Products {
		// for each product in 3scale retrieve product backend usages
		// Required for pulling backendUsages for each product that belongs to a tenant and to make 3scale API call to remove the backendUsage
		backendUsages, err := threescaleAPIClient.ListBackendapiUsages(productIn3scale.Element.ID)
		if err != nil {
			return err
		}

		for _, backendUsage := range backendUsages {
			if backendUsage.Element.BackendAPIID == *backend.Status.ID {
				// retrieve only the CRs that have the given backend usage listed
				crWithBackendUsageList := r.retriveCRsByBackendUsage(productCRsList, backend.Spec.SystemName)

				// for each of returned CR's - remove the backend by the backend.Spec.SystemName and update the product
				for _, productCR := range crWithBackendUsageList {
					delete(productCR.Spec.BackendUsages, backend.Spec.SystemName)
					// update product CR
					err = r.Client().Update(context.TODO(), &productCR)
					if err != nil {
						return err
					}
				}

				// delete backendAPI usage regardless of whether productCR is present or not
				threescaleAPIClient.DeleteBackendapiUsage(productIn3scale.Element.ID, backendUsage.Element.ID)
				if err != nil {
					return err
				}
			}
		}
	}

	// Remove backendAPI
	err = threescaleAPIClient.DeleteBackendApi(*backend.Status.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *BackendReconciler) retriveCRsByBackendUsage(productsCRsList *capabilitiesv1beta1.ProductList, systemName string) (productsList []capabilitiesv1beta1.Product) {
	for _, productCR := range productsCRsList.Items {
		if _, ok := productCR.Spec.BackendUsages[systemName]; ok {
			productsList = append(productsList, productCR)
		}
	}

	return productsList
}

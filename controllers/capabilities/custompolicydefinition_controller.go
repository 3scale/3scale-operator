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
	"github.com/3scale/3scale-operator/apis/common/version"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/go-logr/logr"
)

// CustomPolicyDefinitionReconciler reconciles a CustomPolicyDefinition object
type CustomPolicyDefinitionReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that PolicyReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &CustomPolicyDefinitionReconciler{}

// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=custompolicydefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=custompolicydefinitions/status,verbs=get;update;patch

func (r *CustomPolicyDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Logger().WithValues("custompolicydefinition", req.NamespacedName)
	reqLogger.Info("Reconcile CustomPolicyDefinition", "Operator version", version.Version)

	// Fetch the instance
	customPolicyDefinitionCR := &capabilitiesv1beta1.CustomPolicyDefinition{}
	err := r.Client().Get(context.TODO(), req.NamespacedName, customPolicyDefinitionCR)
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
		jsonData, err := json.MarshalIndent(customPolicyDefinitionCR, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Ignore deleted resource, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if customPolicyDefinitionCR.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	statusReconciler, reconcileErr := r.reconcileSpec(customPolicyDefinitionCR, reqLogger)
	statusResult, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		if reconcileErr != nil {
			return ctrl.Result{}, fmt.Errorf("Failed to reconcile custompolicydefinition: %v. Failed to update custompolicydefinition status: %w", reconcileErr, statusUpdateErr)
		}

		return ctrl.Result{}, fmt.Errorf("Failed to update custompolicydefinition status: %w", statusUpdateErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if reconcileErr != nil {
		if helper.IsInvalidSpecError(reconcileErr) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(customPolicyDefinitionCR, corev1.EventTypeWarning, "Invalid CustomPolicyDefinition Spec", "%v", reconcileErr)
			return ctrl.Result{}, nil
		}

		reqLogger.Error(reconcileErr, "Failed to reconcile")
		r.EventRecorder().Eventf(customPolicyDefinitionCR, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
		return ctrl.Result{}, reconcileErr
	}

	return ctrl.Result{}, nil
}

func (r *CustomPolicyDefinitionReconciler) reconcileSpec(customPolicyDefinitionCR *capabilitiesv1beta1.CustomPolicyDefinition, logger logr.Logger) (*CustomPolicyDefinitionStatusReconciler, error) {
	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), customPolicyDefinitionCR.Namespace, customPolicyDefinitionCR.Spec.ProviderAccountRef, logger)
	if err != nil {
		statusReconciler := NewCustomPolicyDefinitionStatusReconciler(r.BaseReconciler, customPolicyDefinitionCR, "", nil, err)
		return statusReconciler, err
	}

	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		statusReconciler := NewCustomPolicyDefinitionStatusReconciler(r.BaseReconciler, customPolicyDefinitionCR, providerAccount.AdminURLStr, nil, err)
		return statusReconciler, err
	}

	reconciler := NewCustomPolicyDefinitionThreescaleReconciler(r.BaseReconciler, customPolicyDefinitionCR, threescaleAPIClient, providerAccount.AdminURLStr, logger)
	customPolicyObj, err := reconciler.Reconcile()

	statusReconciler := NewCustomPolicyDefinitionStatusReconciler(r.BaseReconciler, customPolicyDefinitionCR, providerAccount.AdminURLStr, customPolicyObj, err)
	return statusReconciler, err
}

func (r *CustomPolicyDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.CustomPolicyDefinition{}).
		Complete(r)
}

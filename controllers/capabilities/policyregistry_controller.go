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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	"github.com/go-logr/logr"
)

// PolicyRegistryReconciler reconciles a PolicyRegistry object
type PolicyRegistryReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that PolicyRegistryReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &PolicyRegistryReconciler{}

// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=policyregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=policyregistries/status,verbs=get;update;patch

func (r *PolicyRegistryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Logger().WithValues("policy", req.NamespacedName)
	reqLogger.Info("Reconcile Policy", "Operator version", version.Version)

	// Fetch the instance
	policyRegistryCR := &capabilitiesv1beta1.PolicyRegistry{}
	err := r.Client().Get(context.TODO(), req.NamespacedName, policyRegistryCR)
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
		jsonData, err := json.MarshalIndent(policyRegistryCR, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Ignore deleted resource, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if policyRegistryCR.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	statusReconciler, reconcileErr := r.reconcileSpec(policyRegistryCR, reqLogger)
	statusResult, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		if reconcileErr != nil {
			return ctrl.Result{}, fmt.Errorf("Failed to reconcile policy: %v. Failed to update policy status: %w", reconcileErr, statusUpdateErr)
		}

		return ctrl.Result{}, fmt.Errorf("Failed to update policy status: %w", statusUpdateErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if reconcileErr != nil {
		if helper.IsInvalidSpecError(reconcileErr) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(policyRegistryCR, corev1.EventTypeWarning, "Invalid Policy Spec", "%v", reconcileErr)
			return ctrl.Result{}, nil
		}

		reqLogger.Error(reconcileErr, "Failed to reconcile")
		r.EventRecorder().Eventf(policyRegistryCR, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
		return ctrl.Result{}, reconcileErr
	}

	return ctrl.Result{}, nil
}

func (r *PolicyRegistryReconciler) reconcileSpec(policyRegistryCR *capabilitiesv1beta1.PolicyRegistry, logger logr.Logger) (*PolicyRegistryStatusReconciler, error) {
	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), policyRegistryCR.Namespace, policyRegistryCR.Spec.ProviderAccountRef, logger)
	if err != nil {
		statusReconciler := NewPolicyRegistryStatusReconciler(r.BaseReconciler, policyRegistryCR, "", nil, err)
		return statusReconciler, err
	}

	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		statusReconciler := NewPolicyRegistryStatusReconciler(r.BaseReconciler, policyRegistryCR, providerAccount.AdminURLStr, nil, err)
		return statusReconciler, err
	}

	reconciler := NewPolicyRegistryThreescaleReconciler(r.BaseReconciler, policyRegistryCR, threescaleAPIClient, providerAccount.AdminURLStr, logger)
	policyObj, err := reconciler.Reconcile()

	statusReconciler := NewPolicyRegistryStatusReconciler(r.BaseReconciler, policyRegistryCR, providerAccount.AdminURLStr, policyObj, err)
	return statusReconciler, err
}

func (r *PolicyRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.PolicyRegistry{}).
		Complete(r)
}

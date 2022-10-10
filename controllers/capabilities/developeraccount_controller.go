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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// DeveloperAccountReconciler reconciles a DeveloperAccount object
type DeveloperAccountReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that DeveloperAccountReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &DeveloperAccountReconciler{}

// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=developeraccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=developeraccounts/status,verbs=get;update;patch

func (r *DeveloperAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Logger().WithValues("developeraccount", req.NamespacedName)
	reqLogger.Info("Reconcile DeveloperAccount", "Operator version", version.Version)

	// Fetch the instance
	developerAccountCR := &capabilitiesv1beta1.DeveloperAccount{}
	err := r.Client().Get(context.TODO(), req.NamespacedName, developerAccountCR)
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
		jsonData, err := json.MarshalIndent(developerAccountCR, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Ignore deleted resource, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if developerAccountCR.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	statusReconciler, reconcileErr := r.reconcileSpec(developerAccountCR, reqLogger)
	statusResult, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		if reconcileErr != nil {
			return ctrl.Result{}, fmt.Errorf("Failed to reconcile developer account: %v. Failed to update status: %w", reconcileErr, statusUpdateErr)
		}

		return ctrl.Result{}, fmt.Errorf("Failed to update developers account status: %w", statusUpdateErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if reconcileErr != nil {
		if helper.IsInvalidSpecError(reconcileErr) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(developerAccountCR, corev1.EventTypeWarning, "Invalid developer account spec", "%v", reconcileErr)
			return ctrl.Result{}, nil
		}

		if helper.IsWaitError(reconcileErr) {
			// On wait error, retry
			reqLogger.Info("retrying", "reason", reconcileErr)
			return ctrl.Result{Requeue: true}, nil
		}

		reqLogger.Error(reconcileErr, "Failed to reconcile")
		r.EventRecorder().Eventf(developerAccountCR, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
		return ctrl.Result{}, reconcileErr
	}

	return ctrl.Result{}, nil
}

func (r *DeveloperAccountReconciler) reconcileSpec(accountCR *capabilitiesv1beta1.DeveloperAccount, logger logr.Logger) (*DeveloperAccountStatusReconciler, error) {
	err := r.validateSpec(accountCR)
	if err != nil {
		statusReconciler := NewDeveloperAccountStatusReconciler(r.BaseReconciler, accountCR, "", nil, err)
		return statusReconciler, err
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), accountCR.Namespace, accountCR.Spec.ProviderAccountRef, logger)
	if err != nil {
		statusReconciler := NewDeveloperAccountStatusReconciler(r.BaseReconciler, accountCR, "", nil, err)
		return statusReconciler, err
	}

	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		statusReconciler := NewDeveloperAccountStatusReconciler(r.BaseReconciler, accountCR, providerAccount.AdminURLStr, nil, err)
		return statusReconciler, err
	}

	reconciler := NewDeveloperAccountThreescaleReconciler(r.BaseReconciler, accountCR, threescaleAPIClient, providerAccount.AdminURLStr, logger)
	accountObj, err := reconciler.Reconcile()

	statusReconciler := NewDeveloperAccountStatusReconciler(r.BaseReconciler, accountCR, providerAccount.AdminURLStr, accountObj, err)
	return statusReconciler, err
}

func (r *DeveloperAccountReconciler) validateSpec(resource *capabilitiesv1beta1.DeveloperAccount) error {
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

func (r *DeveloperAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.DeveloperAccount{}).
		Complete(r)
}

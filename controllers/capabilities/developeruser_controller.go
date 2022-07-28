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

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const developerUserFinalizer = "developeruser.capabilities.3scale.net/finalizer"

// DeveloperUserReconciler reconciles a DeveloperUser object
type DeveloperUserReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that DeveloperUserReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &DeveloperUserReconciler{}

// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=developerusers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=developerusers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=developerusers/finalizers,verbs=get;list;watch;create;update;patch;delete

func (r *DeveloperUserReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Logger().WithValues("developeruser", req.NamespacedName)
	reqLogger.Info("Reconcile DeveloperUser", "Operator version", version.Version)

	// Fetch the instance
	developerUserCR := &capabilitiesv1beta1.DeveloperUser{}
	err := r.Client().Get(context.TODO(), req.NamespacedName, developerUserCR)
	if err != nil {
		if apierrors.IsNotFound(err) {
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
		jsonData, err := json.MarshalIndent(developerUserCR, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// DeveloperUser has been marked for deletion
	if developerUserCR.GetDeletionTimestamp() != nil && controllerutil.ContainsFinalizer(developerUserCR, developerUserFinalizer) {
		err = r.removeDeveloperUserFrom3scale(developerUserCR)
		if err != nil {
			r.EventRecorder().Eventf(developerUserCR, corev1.EventTypeWarning, "Failed to delete developer user", "%v", err)

			// Update status with err
			statusResult, statusUpdateErr := NewDeveloperUserStatusReconciler(r.BaseReconciler, developerUserCR, nil, "", nil, err).Reconcile()
			if statusUpdateErr != nil {
				return ctrl.Result{}, fmt.Errorf("Failed to update developers user status: %w", statusUpdateErr)
			}

			if statusResult.Requeue {
				return statusResult, nil
			}

			return ctrl.Result{}, err
		}

		controllerutil.RemoveFinalizer(developerUserCR, developerUserFinalizer)
		err = r.UpdateResource(developerUserCR)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Ignore deleted resource, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if developerUserCR.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(developerUserCR, developerUserFinalizer) {
		controllerutil.AddFinalizer(developerUserCR, developerUserFinalizer)
		err = r.UpdateResource(developerUserCR)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Retrieve ownersReference of DeveloperAccount CR that owns the DeveloperUser CR
	developerAccountCR, err := r.retrieveDevelopAccountCR(developerUserCR)
	if err != nil {
		return ctrl.Result{}, err
	}

	// If DeveloperAccount CR is found, set it's ownersReference as ownerReference in the DeveloperUser CR
	if developerAccountCR != nil {
		updated, err := r.EnsureOwnerReference(developerAccountCR, developerUserCR)
		if err != nil {
			return ctrl.Result{}, err
		}

		if updated {
			err := r.Client().Update(r.Context(), developerUserCR)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	}

	statusReconciler, reconcileErr := r.reconcileSpec(developerUserCR, reqLogger)
	statusResult, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		if reconcileErr != nil {
			return ctrl.Result{}, fmt.Errorf("Failed to reconcile developer user: %v. Failed to update status: %w", reconcileErr, statusUpdateErr)
		}

		return ctrl.Result{}, fmt.Errorf("Failed to update developers user status: %w", statusUpdateErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if reconcileErr != nil {
		if helper.IsInvalidSpecError(reconcileErr) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(developerUserCR, corev1.EventTypeWarning, "Invalid developer user spec", "%v", reconcileErr)
			return ctrl.Result{}, nil
		}

		if helper.IsOrphanSpecError(reconcileErr) {
			// On Orphan spec error, retry
			reqLogger.Info("orphan", "message", reconcileErr)
			return ctrl.Result{Requeue: true}, nil
		}

		reqLogger.Error(reconcileErr, "Failed to reconcile")
		r.EventRecorder().Eventf(developerUserCR, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
		return ctrl.Result{}, reconcileErr
	}

	return ctrl.Result{}, nil
}

func (r *DeveloperUserReconciler) reconcileSpec(userCR *capabilitiesv1beta1.DeveloperUser, logger logr.Logger) (*DeveloperUserStatusReconciler, error) {
	err := r.validateSpec(userCR)
	if err != nil {
		statusReconciler := NewDeveloperUserStatusReconciler(r.BaseReconciler, userCR, nil, "", nil, err)
		return statusReconciler, err
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), userCR.Namespace, userCR.Spec.ProviderAccountRef, logger)
	if err != nil {
		statusReconciler := NewDeveloperUserStatusReconciler(r.BaseReconciler, userCR, nil, "", nil, err)
		return statusReconciler, err
	}

	parentAccountCR, err := r.findParentAccount(userCR, providerAccount, logger)
	if err != nil {
		statusReconciler := NewDeveloperUserStatusReconciler(r.BaseReconciler, userCR, nil, providerAccount.AdminURLStr, nil, err)
		return statusReconciler, err
	}

	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		statusReconciler := NewDeveloperUserStatusReconciler(r.BaseReconciler, userCR, parentAccountCR, providerAccount.AdminURLStr, nil, err)
		return statusReconciler, err
	}

	reconciler := NewDeveloperUserThreescaleReconciler(r.BaseReconciler, userCR, parentAccountCR, threescaleAPIClient, providerAccount.AdminURLStr, logger)
	userObj, err := reconciler.Reconcile()

	statusReconciler := NewDeveloperUserStatusReconciler(r.BaseReconciler, userCR, parentAccountCR, providerAccount.AdminURLStr, userObj, err)
	return statusReconciler, err
}

func (r *DeveloperUserReconciler) validateSpec(resource *capabilitiesv1beta1.DeveloperUser) error {
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

func (r *DeveloperUserReconciler) findParentAccount(userCR *capabilitiesv1beta1.DeveloperUser, userProviderAccount *controllerhelper.ProviderAccount, logger logr.Logger) (*capabilitiesv1beta1.DeveloperAccount, error) {
	parentAccountFldPath := field.NewPath("spec").Child("developerAccountRef")

	devAccountCR := &capabilitiesv1beta1.DeveloperAccount{}
	devAccountKey := types.NamespacedName{Name: userCR.Spec.DeveloperAccountRef.Name, Namespace: userCR.Namespace}
	if err := r.Client().Get(r.Context(), devAccountKey, devAccountCR); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, &helper.SpecFieldError{
				ErrorType: helper.OrphanError,
				FieldErrorList: field.ErrorList{
					field.Invalid(parentAccountFldPath, userCR.Spec.DeveloperAccountRef, "parent account resource not found"),
				},
			}
		}

		return nil, err
	}

	// Check it belongs to the same providerAccount
	parentProviderAccount, err := controllerhelper.LookupProviderAccount(r.Client(), userCR.Namespace, devAccountCR.Spec.ProviderAccountRef, logger)
	if err != nil {
		return nil, err
	}

	// Filter by provider account
	if userProviderAccount.AdminURLStr != parentProviderAccount.AdminURLStr {
		return nil, &helper.SpecFieldError{
			ErrorType: helper.OrphanError,
			FieldErrorList: field.ErrorList{
				field.Invalid(parentAccountFldPath, userCR.Spec.DeveloperAccountRef, "parent account resource does not belong to the same provider account"),
			},
		}
	}

	if !devAccountCR.Status.IsReady() {
		return nil, &helper.SpecFieldError{
			ErrorType: helper.OrphanError,
			FieldErrorList: field.ErrorList{
				field.Invalid(parentAccountFldPath, userCR.Spec.DeveloperAccountRef, "parent account resource not ready"),
			},
		}
	}

	return devAccountCR, nil
}

func (r *DeveloperUserReconciler) removeDeveloperUserFrom3scale(developerUser *capabilitiesv1beta1.DeveloperUser) error {
	logger := r.Logger().WithValues("developerUser", client.ObjectKey{Name: developerUser.Name, Namespace: developerUser.Namespace})

	// Attempt to remove developerUser only if developerUser.Status.ID is present
	if developerUser.Status.ID == nil {
		logger.Info("could not remove developerUser because ID is missing in status")
		return nil
	}

	if developerUser.Status.AccountID == nil {
		logger.Info("could not remove developerUser because Account ID is missing in status")
		return nil
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), developerUser.Namespace, developerUser.Spec.ProviderAccountRef, logger)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("developer user not deleted from 3scale, provider account not found")
			return nil
		}
		return err
	}

	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		return err
	}

	err = threescaleAPIClient.DeleteDeveloperUser(*developerUser.Status.AccountID, *developerUser.Status.ID)
	if err != nil && !threescaleapi.IsNotFound(err) {
		return err
	}

	return nil
}

func (r *DeveloperUserReconciler) retrieveDevelopAccountCR(developerUser *capabilitiesv1beta1.DeveloperUser) (*capabilitiesv1beta1.DeveloperAccount, error) {
	developerAccount := &capabilitiesv1beta1.DeveloperAccount{}

	if err := r.Client().Get(context.TODO(), client.ObjectKey{Namespace: developerUser.Namespace, Name: developerUser.Spec.DeveloperAccountRef.Name}, developerAccount); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}

		return nil, nil
	}

	return developerAccount, nil
}

func (r *DeveloperUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.DeveloperUser{}).
		Complete(r)
}

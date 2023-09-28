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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	// tenant deletion state
	scheduledForDeletionState = "scheduled_for_deletion"

	// tenant finalizer
	tenantFinalizer = "tenant.capabilities.3scale.net/finalizer"

	// Secret field name with Tenant's admin user password
	TenantAdminPasswordSecretField = "admin_password"

	// Tenant's credentials secret field name for access token
	TenantAccessTokenSecretField = "token"

	// Tenant's credentials secret field name for admin domain url
	TenantAdminDomainKeySecretField = "adminURL"

	// Tenant ID annotation matches the tenant.status.ID
	tenantIdAnnotation = "tenantID"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that TenantReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &TenantReconciler{}

// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=tenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=tenants/finalizers,verbs=get;list;watch;create;update;patch;delete

func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Logger().WithValues("tenant", req.NamespacedName)
	reqLogger.Info("Reconcile Tenant", "Operator version", version.Version)

	// Fetch the Tenant instance
	tenantR := &capabilitiesv1alpha1.Tenant{}
	err := r.Client().Get(context.TODO(), req.NamespacedName, tenantR)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Tenant resource not found")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Ignore deleted resources, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if tenantR.GetDeletionTimestamp() != nil && !controllerutil.ContainsFinalizer(tenantR, tenantFinalizer) {
		return ctrl.Result{}, nil
	}

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(tenantR, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// reconcileErr can be an error of different types - generic k8s err, isInvalid, waitError, isOrphaned
	// different errors are handled differently when it comes to requeueing the reconciliation loop
	statusReconciler, reconcileErr := r.specReconciler(tenantR, reqLogger)

	// statusUpdateError can be an error when making the update call to the resource
	res, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update tenantCR: %w", statusUpdateErr)
	}

	// Requeue if either object or status were updated
	if res.Requeue {
		return reconcile.Result{}, nil
	}

	if reconcileErr != nil {
		// On validation error, do not retry since there's something wrong with the spec
		if helper.IsInvalidSpecError(reconcileErr) {
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(tenantR, corev1.EventTypeWarning, "Invalid Spec", "%v", reconcileErr)
			return ctrl.Result{}, nil
		}

		// On wait error, retry
		if helper.IsWaitError(reconcileErr) {
			reqLogger.Info("ERROR", "wait error", reconcileErr)
			return ctrl.Result{Requeue: true}, nil
		}

		// On orphan error, do not retry
		if helper.IsOrphanSpecError(reconcileErr) {
			reqLogger.Info("ERROR", "orphan error", reconcileErr)
			return ctrl.Result{Requeue: false}, nil
		}

		// On all other errors, retry
		reqLogger.Error(reconcileErr, "failed to reconcile")
		r.EventRecorder().Eventf(tenantR, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
		return ctrl.Result{}, reconcileErr
	}

	reqLogger.Info("Tenant reconciled successfully")
	return ctrl.Result{}, nil
}

func (r *TenantReconciler) specReconciler(tenantCR *capabilitiesv1alpha1.Tenant, logger logr.Logger) (*TenantStatusReconciler, error) {
	// Save a copy of the original tenant CR prior to changing it for status reconciler
	originalTenantCR := tenantCR.DeepCopy()

	// Setup porta client
	portaClient, err := r.setupPortaClient(tenantCR, originalTenantCR, logger)
	if err != nil {
		statusReconciler := NewTenantStatusReconciler(r.BaseReconciler, tenantCR, originalTenantCR, err)
		return statusReconciler, err
	}

	// Process tenant deletion
	tenantInDeletion, err := r.reconcileDeletion(tenantCR, portaClient, logger)
	if err != nil || tenantInDeletion {
		statusReconciler := NewTenantStatusReconciler(r.BaseReconciler, tenantCR, originalTenantCR, err)
		return statusReconciler, err
	}

	// Adding annotations and finalizers
	changedMetadata := r.reconcileMetadata(tenantCR)

	// Bring tenant secret ref and namespace to lower if setup incorrectly
	changedDefaults := tenantCR.SetDefaults()

	if changedMetadata || changedDefaults {
		statusReconciler := NewTenantStatusReconciler(r.BaseReconciler, tenantCR, originalTenantCR, err)
		return statusReconciler, err
	}

	// Main logic for reconciling the tenant CR against 3scale
	internalReconciler := NewTenantInternalReconciler(r.BaseReconciler, tenantCR, portaClient, logger)
	err = internalReconciler.Run()
	if err != nil {
		// Can be multiple error types depending on where it failed
		statusReconciler := NewTenantStatusReconciler(r.BaseReconciler, tenantCR, originalTenantCR, err)
		return statusReconciler, err
	}

	statusReconciler := NewTenantStatusReconciler(r.BaseReconciler, tenantCR, originalTenantCR, err)
	return statusReconciler, err
}

func (r *TenantReconciler) reconcileMetadata(tenantCR *capabilitiesv1alpha1.Tenant) bool {
	changed := false
	// If the tenant.Status.TenantID is found and the annotation is not found - create
	// If the tenant.Status.TenantID is found and the annotation is found but, the value of annotation is different to the status.TenantID - update
	tenantId := tenantCR.Status.TenantId
	if value, found := tenantCR.ObjectMeta.Annotations[tenantIdAnnotation]; tenantId != 0 && !found || tenantId != 0 && found && value != strconv.FormatInt(tenantCR.Status.TenantId, 10) {
		if tenantCR.ObjectMeta.Annotations == nil {
			tenantCR.ObjectMeta.Annotations = make(map[string]string)
		}
		tenantCR.ObjectMeta.Annotations[tenantIdAnnotation] = strconv.FormatInt(tenantCR.Status.TenantId, 10)
		changed = true
	}

	if !controllerutil.ContainsFinalizer(tenantCR, tenantFinalizer) {
		controllerutil.AddFinalizer(tenantCR, tenantFinalizer)
		changed = true
	}

	return changed
}

func (r *TenantReconciler) fetchMasterCredentials(tenantR *capabilitiesv1alpha1.Tenant) (string, error) {
	masterCredentialsSecret := &corev1.Secret{}

	err := r.Client().Get(context.TODO(), tenantR.MasterSecretKey(), masterCredentialsSecret)

	if err != nil {
		return "", err
	}

	masterAccessTokenByteArray, ok := masterCredentialsSecret.Data[component.SystemSecretSystemSeedMasterAccessTokenFieldName]
	if !ok {
		return "", &helper.WaitError{
			Err: fmt.Errorf("key not found in master secret (%s) key: %s", tenantR.MasterSecretKey(), component.SystemSecretSystemSeedMasterAccessTokenFieldName),
		}
	}

	return bytes.NewBuffer(masterAccessTokenByteArray).String(), nil
}

func (r *TenantReconciler) setupPortaClient(tenantCR, originalTenantCR *capabilitiesv1alpha1.Tenant, logger logr.Logger) (*threescaleapi.ThreeScaleClient, error) {
	masterAccessToken, err := r.fetchMasterCredentials(tenantCR)
	if err != nil {
		logger.Error(err, "Error fetching master credentials secret")
		return nil, err
	}

	insecureSkipVerify := controllerhelper.GetInsecureSkipVerifyAnnotation(tenantCR.GetAnnotations())
	portaClient, err := controllerhelper.PortaClientFromURLString(tenantCR.Spec.SystemMasterUrl, masterAccessToken, insecureSkipVerify)
	if err != nil {
		return nil, err
	}

	return portaClient, nil
}

func (r *TenantReconciler) reconcileDeletion(tenantCR *capabilitiesv1alpha1.Tenant, portaClient *threescaleapi.ThreeScaleClient, logger logr.Logger) (bool, error) {
	if tenantCR.GetDeletionTimestamp() != nil && controllerutil.ContainsFinalizer(tenantCR, tenantFinalizer) {
		existingTenant, err := controllerhelper.FetchTenant(tenantCR.Status.TenantId, portaClient)
		if err != nil {
			// Error fetching tenant from 3scale
			return true, err
		}

		// delete tenantCR if tenant is present in 3scale
		if existingTenant != nil {
			// do not attempt to delete tenant that is already scheduled for deletion
			if existingTenant.Signup.Account.State != scheduledForDeletionState {
				err := portaClient.DeleteTenant(tenantCR.Status.TenantId)
				if err != nil {
					r.EventRecorder().Eventf(tenantCR, corev1.EventTypeWarning, "Failed to delete tenant", "%v", err)
					// Error when deleting the tenant from 3scale
					return true, err
				}
			} else {
				logger.Info("Removing tenant CR - tenant is already scheduled for deletion", "tenantID", existingTenant.Signup.Account.ID)
				parentFldPath := field.NewPath("status").Child("TenantID")
				err := &helper.SpecFieldError{
					ErrorType: helper.OrphanError,
					FieldErrorList: field.ErrorList{
						field.Invalid(parentFldPath, tenantCR.Status.TenantId, "tenant account already marked for deletion, please delete the tenant CR manually"),
					},
				}
				return true, err
			}
		}

		controllerutil.RemoveFinalizer(tenantCR, tenantFinalizer)
		return true, nil
	}

	return false, nil
}

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1alpha1.Tenant{}).
		Complete(r)
}

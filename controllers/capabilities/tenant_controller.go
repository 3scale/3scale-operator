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
	tenantCR := &capabilitiesv1alpha1.Tenant{}
	err := r.Client().Get(context.TODO(), req.NamespacedName, tenantCR)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Tenant resource not found")
			return ctrl.Result{}, nil
		}
		// On error, always requeue the request
		return ctrl.Result{}, err
	}

	originalTenantCR := tenantCR.DeepCopy()

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(tenantCR, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Setup porta client
	portaClient, err := r.setupPortaClient(tenantCR, reqLogger)
	if err != nil {
		_, statusReconcilerError := r.reconcileStatus(tenantCR, originalTenantCR, err)
		if statusReconcilerError != nil {
			return helper.ReconcileErrorHandler(err, reqLogger), nil
		}

		return helper.ReconcileErrorHandler(err, reqLogger), nil
	}

	// Reconcile whether or not the tenant should be removed
	if tenantCR.GetDeletionTimestamp() != nil && controllerutil.ContainsFinalizer(tenantCR, tenantFinalizer) {
		existingTenant, err := controllerhelper.FetchTenant(tenantCR.Status.TenantId, portaClient)
		if err != nil {
			return ctrl.Result{}, err
		}

		// delete tenantCR if tenant is present in 3scale
		if existingTenant != nil {
			// do not attempt to delete tenant that is already scheduled for deletion
			if existingTenant.Signup.Account.State != scheduledForDeletionState {
				err := portaClient.DeleteTenant(tenantCR.Status.TenantId)
				if err != nil {
					r.EventRecorder().Eventf(tenantCR, corev1.EventTypeWarning, "Failed to delete tenant", "%v", err)
					return ctrl.Result{}, err
				}
			} else {
				reqLogger.Info("Removing tenant CR - tenant is already scheduled for deletion", "tenantID", existingTenant.Signup.Account.ID)
			}
		}

		controllerutil.RemoveFinalizer(tenantCR, tenantFinalizer)
		err = r.UpdateResource(tenantCR)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Ignore deleted resources, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if tenantCR.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	// Update/Check metadata and defaults
	metadataUpdated := r.reconcileMetadata(tenantCR)
	defaultsUpdated := tenantCR.SetDefaults()
	if metadataUpdated || defaultsUpdated {
		err := r.UpdateResource(tenantCR)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Do not retrigger - automatic retrigger done by udpating the tenantCR
		return ctrl.Result{}, nil
	}

	// Validate and update spec if required
	internalReconciler := NewTenantThreescaleReconciler(r.BaseReconciler, tenantCR, portaClient, reqLogger)
	specReconcileErr := internalReconciler.Run()
	statusIsEqual, statusReconcilerError := r.reconcileStatus(tenantCR, originalTenantCR, specReconcileErr)
	if statusReconcilerError != nil {
		return helper.ReconcileErrorHandler(statusReconcilerError, reqLogger), nil
	}

	if specReconcileErr != nil {
		return helper.ReconcileErrorHandler(specReconcileErr, reqLogger), nil
	}

	// If error did not occur and the status was updated, quit the reoncile loop since another reconcile is incoming
	if !statusIsEqual {
		return ctrl.Result{}, nil
	}

	reqLogger.Info("Tenant reconciled successfully")
	return ctrl.Result{}, nil
}

func (r *TenantReconciler) reconcileStatus(updatedTenantCR, originalTenantCR *capabilitiesv1alpha1.Tenant, reconcileError error) (bool, error) {
	statusReconciler := NewTenantStatusReconciler(r.BaseReconciler, updatedTenantCR, originalTenantCR, reconcileError)
	statusEqual, err := statusReconciler.Reconcile()
	if err != nil {
		return statusEqual, err
	}

	return statusEqual, nil
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

func (r *TenantReconciler) setupPortaClient(tenantCR *capabilitiesv1alpha1.Tenant, logger logr.Logger) (*threescaleapi.ThreeScaleClient, error) {
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

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1alpha1.Tenant{}).
		Complete(r)
}

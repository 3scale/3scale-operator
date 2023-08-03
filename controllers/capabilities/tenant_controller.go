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

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	"github.com/3scale/3scale-operator/apis/common/version"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

// tenant deletion state
const scheduledForDeletionState = "scheduled_for_deletion"

// tenant finalizer
const tenantFinalizer = "tenant.capabilities.3scale.net/finalizer"

// Secret field name with Tenant's admin user password
const TenantAdminPasswordSecretField = "admin_password"

// Tenant's credentials secret field name for access token
const TenantAccessTokenSecretField = "token"

// Tenant's credentials secret field name for admin domain url
const TenantAdminDomainKeySecretField = "adminURL"

// Tenant ID annotation matches the tenant.status.ID
const tenantIdAnnotation = "tenantID"

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
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Tenant resource not found")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(tenantR, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	masterAccessToken, err := r.fetchMasterCredentials(tenantR)
	if err != nil {
		reqLogger.Error(err, "Error fetching master credentials secret")
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	portaClient, err := controllerhelper.PortaClientFromURLString(tenantR.Spec.SystemMasterUrl, masterAccessToken)
	if err != nil {
		reqLogger.Error(err, "Error creating porta client object")
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Tenant has been marked for deletion
	if tenantR.GetDeletionTimestamp() != nil && controllerutil.ContainsFinalizer(tenantR, tenantFinalizer) {
		existingTenant, err := controllerhelper.FetchTenant(tenantR.Status.TenantId, portaClient)
		if err != nil {
			return ctrl.Result{}, err
		}

		// delete tenantCR if tenant is present in 3scale
		if existingTenant != nil {
			// do not attempt to delete tenant that is already scheduled for deletion
			if existingTenant.Signup.Account.State != scheduledForDeletionState {
				err := portaClient.DeleteTenant(tenantR.Status.TenantId)
				if err != nil {
					r.EventRecorder().Eventf(tenantR, corev1.EventTypeWarning, "Failed to delete tenant", "%v", err)
					return ctrl.Result{}, err
				}
			} else {
				reqLogger.Info("Removing tenant CR - tenant is already scheduled for deletion", "tenantID", existingTenant.Signup.Account.ID)
			}
		}

		controllerutil.RemoveFinalizer(tenantR, tenantFinalizer)
		err = r.UpdateResource(tenantR)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Ignore deleted resources, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if tenantR.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	// If the tenant.Status.TenantID is found and the annotation is not found - create
	// If the tenant.Status.TenantID is found and the annotation is found but, the value of annotation is different to the status.TenantID - update
	tenantIdAnnotationFound := true
	tenantId := tenantR.Status.TenantId
	if value, found := tenantR.ObjectMeta.Annotations[tenantIdAnnotation]; tenantId != 0 && !found || tenantId != 0 && found && value != strconv.FormatInt(tenantR.Status.TenantId, 10) {
		if tenantR.ObjectMeta.Annotations == nil {
			tenantR.ObjectMeta.Annotations = make(map[string]string)
		}

		tenantR.ObjectMeta.Annotations[tenantIdAnnotation] = strconv.FormatInt(tenantR.Status.TenantId, 10)
		tenantIdAnnotationFound = false
	}

	tenantFinalizerFound := true
	if !controllerutil.ContainsFinalizer(tenantR, tenantFinalizer) {
		controllerutil.AddFinalizer(tenantR, tenantFinalizer)
		tenantFinalizerFound = false
	}

	if !tenantFinalizerFound || !tenantIdAnnotationFound {
		err = r.UpdateResource(tenantR)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	changed := tenantR.SetDefaults()
	if changed {
		err = r.UpdateResource(tenantR)
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.Info("Tenant resource updated with defaults")
		// Expect for re-trigger
		return ctrl.Result{}, nil
	}

	internalReconciler := NewTenantInternalReconciler(r.BaseReconciler, tenantR, portaClient, reqLogger)
	res, err := internalReconciler.Run()
	if err != nil {
		reqLogger.Error(err, "Error in tenant reconciliation")
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if res.Requeue {
		return res, nil
	}

	reqLogger.Info("Tenant reconciled successfully")
	return ctrl.Result{}, nil
}

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1alpha1.Tenant{}).
		Complete(r)
}

func (r *TenantReconciler) fetchMasterCredentials(tenantR *capabilitiesv1alpha1.Tenant) (string, error) {
	masterCredentialsSecret := &v1.Secret{}

	err := r.Client().Get(context.TODO(), tenantR.MasterSecretKey(), masterCredentialsSecret)

	if err != nil {
		return "", err
	}

	masterAccessTokenByteArray, ok := masterCredentialsSecret.Data[component.SystemSecretSystemSeedMasterAccessTokenFieldName]
	if !ok {
		return "", fmt.Errorf("Key not found in master secret (%s) key: %s",
			tenantR.MasterSecretKey(),
			component.SystemSecretSystemSeedMasterAccessTokenFieldName)
	}

	return bytes.NewBuffer(masterAccessTokenByteArray).String(), nil
}

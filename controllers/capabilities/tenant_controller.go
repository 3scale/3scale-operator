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
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
)

// Secret field name with Tenant's admin user password
const TenantAdminPasswordSecretField = "admin_password"

// Tenant's credentials secret field name for access token
const TenantProviderKeySecretField = "token"

// Tenant's credentials secret field name for admin domain url
const TenantAdminDomainKeySecretField = "adminURL"

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=tenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=tenants/finalizers,verbs=get;list;watch;create;update;patch;delete

func (r *TenantReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log.WithValues("tenant", req.NamespacedName)

	// Fetch the Tenant instance
	tenantR := &capabilitiesv1alpha1.Tenant{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, tenantR)
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

	changed := tenantR.SetDefaults()
	if changed {
		err = r.Client.Update(context.TODO(), tenantR)
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.Info("Tenant resource updated with defaults")
		// Expect for re-trigger
		return ctrl.Result{}, nil
	}

	masterAccessToken, err := r.FetchMasterCredentials(r.Client, tenantR)
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

	internalReconciler := NewTenantInternalReconciler(r.Client, tenantR, portaClient, reqLogger)
	err = internalReconciler.Run()
	if err != nil {
		reqLogger.Error(err, "Error in tenant reconciliation")
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	reqLogger.Info("Tenant reconciled successfully")
	return ctrl.Result{}, nil
}

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1alpha1.Tenant{}).
		Complete(r)
}

// FetchMasterCredentials get secret using k8s client
func (r *TenantReconciler) FetchMasterCredentials(k8sClient client.Client, tenantR *capabilitiesv1alpha1.Tenant) (string, error) {
	masterCredentialsSecret := &v1.Secret{}

	err := k8sClient.Get(context.TODO(),
		types.NamespacedName{
			Name: tenantR.Spec.MasterCredentialsRef.Name,
			// Master credential secret MUST be on same namespace as tenant CR
			Namespace: tenantR.Spec.MasterCredentialsRef.Namespace,
		},
		masterCredentialsSecret)

	if err != nil {
		return "", err
	}

	masterAccessTokenByteArray, ok := masterCredentialsSecret.Data[component.SystemSecretSystemSeedMasterAccessTokenFieldName]
	if !ok {
		return "", fmt.Errorf("Key not found in master secret (ns: %s, name: %s) key: %s",
			tenantR.Spec.MasterCredentialsRef.Namespace, tenantR.Spec.MasterCredentialsRef.Name,
			component.SystemSecretSystemSeedMasterAccessTokenFieldName)
	}

	return bytes.NewBuffer(masterAccessTokenByteArray).String(), nil
}

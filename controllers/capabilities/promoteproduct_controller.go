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
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/version"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

// PromoteProductReconciler reconciles a PromoteProduct object
type PromoteProductReconciler struct {
	*reconcilers.BaseReconciler
}

// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=promoteproducts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=promoteproducts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=promoteproducts/finalizers,verbs=get;list;watch;create;update;patch;delete

func (r *PromoteProductReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Logger().WithValues("promoteproduct", req.NamespacedName)
	reqLogger.Info("Reconcile Product", "Operator version", version.Version)

	// your logic here
	promoteProduct := &capabilitiesv1beta1.PromoteProduct{}
	err := r.Client().Get(r.Context(), req.NamespacedName, promoteProduct)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("resource not found. Ignoring since object must have been deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request
		return ctrl.Result{}, err
	}

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(promoteProduct, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	status, err := r.promoteProductReconciler(promoteProduct, reqLogger)
	reqLogger.WithValues("status", status)

	return ctrl.Result{}, nil
}

/** promoteProduct reconciler acts on the following fields in promoteProduct.spec
- promote
- productId
- promoteVersion
- promoteEnvironment
**/
func (r *PromoteProductReconciler) promoteProductReconciler(promoteProduct *capabilitiesv1beta1.PromoteProduct, reqLogger logr.Logger) (capabilitiesv1beta1.PromoteProductStatus, error) {

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), promoteProduct.GetNamespace(), promoteProduct.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		status := r.promoteProductStatus(nil, promoteProduct, err, reqLogger)
		return status, err
	}

	// connect to the 3scale porta client
	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		status := r.promoteProductStatus(threescaleAPIClient, promoteProduct, err, reqLogger)
		return status, err
	}

	if promoteProduct.Spec.ProductId == "" {
		reqLogger.Info("empty productId selected, expected string number productId")
		status := r.promoteProductStatus(threescaleAPIClient, promoteProduct, err, reqLogger)
		return status, err
	}

	// check if promotion is enabled in the CR
	if promoteProduct.Spec.Promote && promoteProduct.Spec.ProductId != "" {
		productIdInt64, err := strconv.ParseInt(promoteProduct.Spec.ProductId, 10, 64)
		if err != nil {
			status := r.promoteProductStatus(threescaleAPIClient, promoteProduct, err, reqLogger)
			return status, err
		}

		if promoteProduct.Spec.PromoteEnvironment == "staging" {
			_, err = threescaleAPIClient.DeployProductProxy(productIdInt64)
			if err != nil {
				status := r.promoteProductStatus(threescaleAPIClient, promoteProduct, err, reqLogger)
				return status, err
			}
			promoteProduct.Spec.Promote = false
			err = r.Client().Update(r.Context(), promoteProduct)
		}
		if promoteProduct.Spec.PromoteEnvironment == "production" {
			if promoteProduct.Spec.PromoteVersion == "" {
				reqLogger.Info("empty configuration version selected, expected a number string for promoteVersion")
				status := r.promoteProductStatus(threescaleAPIClient, promoteProduct, err, reqLogger)
				return status, err
			}
			if promoteProduct.Spec.PromoteVersion != "" {
				_, err = threescaleAPIClient.PromoteProxyConfig(promoteProduct.Spec.ProductId, "sandbox", promoteProduct.Spec.PromoteVersion, promoteProduct.Spec.PromoteEnvironment)
				if err != nil {
					status := r.promoteProductStatus(threescaleAPIClient, promoteProduct, err, reqLogger)
					return status, err
				}
			}
			promoteProduct.Spec.Promote = false
			err = r.Client().Update(r.Context(), promoteProduct)
		}

		if promoteProduct.Spec.PromoteEnvironment != "staging" && promoteProduct.Spec.PromoteEnvironment != "production" {
			reqLogger.Info("invalid environment selected, expected 'staging' or 'production' for promoteEnvironment")
			status := r.promoteProductStatus(threescaleAPIClient, promoteProduct, err, reqLogger)
			return status, err
		}
	}
	status := r.promoteProductStatus(threescaleAPIClient, promoteProduct, err, reqLogger)
	return status, nil
}

//  promoteProductStatus generates the status for existing promoteProduct CR
func (r *PromoteProductReconciler) promoteProductStatus(threescaleAPIClient *client.ThreeScaleClient, promoteProduct *capabilitiesv1beta1.PromoteProduct, err error, reqLogger logr.Logger) capabilitiesv1beta1.PromoteProductStatus {
	promoteProduct.Status.ProductId = promoteProduct.Spec.ProductId
	promoteProduct.Status.LatestPromoteEnvironment = promoteProduct.Spec.PromoteEnvironment
	if threescaleAPIClient != nil {
		getLatestProxyConfig, err := threescaleAPIClient.GetLatestProxyConfig(promoteProduct.Spec.ProductId, promoteProduct.Spec.PromoteEnvironment)
		if err != nil {
			reqLogger.Info("Failed to get latest proxy config ", err)
		}
		promoteProduct.Status.LatestPromoteVersion = getLatestProxyConfig.ProxyConfig.Version
	}
	err = r.Client().Status().Update(r.Context(), promoteProduct)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("failed to update the status in promoteProduct", promoteProduct.Name)
		}

	}

	return promoteProduct.Status
}

func (r *PromoteProductReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.PromoteProduct{}).
		Complete(r)
}

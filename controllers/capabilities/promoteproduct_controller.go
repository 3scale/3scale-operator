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

// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=promoteproducts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=promoteproducts/status,verbs=get;update;patch

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

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), promoteProduct.GetNamespace(), promoteProduct.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		return ctrl.Result{}, err
	}

	// connect to the 3scale porta client
	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		return ctrl.Result{}, err
	}

	if promoteProduct.Spec.ProductId == "" {
		reqLogger.Info("empty productId selected, expected string number productId")
		return ctrl.Result{}, nil
	}
	// check if promotion is enabled in the CR
	if promoteProduct.Spec.Promote && promoteProduct.Spec.ProductId != "" {
		// convert productId to int64
		productIdInt64, err := strconv.ParseInt(promoteProduct.Spec.ProductId, 10, 64)
		if err != nil {
			return ctrl.Result{}, err
		}

		if promoteProduct.Spec.PromoteEnvironment == "staging" {
			_, err = threescaleAPIClient.DeployProductProxy(productIdInt64)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		if promoteProduct.Spec.PromoteEnvironment == "production" {
			if promoteProduct.Spec.PromoteVersion == "" {
				reqLogger.Info("empty configuration version selected, expected a number string for promoteVersion")
				return ctrl.Result{}, nil
			}
			if promoteProduct.Spec.PromoteVersion != "" {
				_, err = threescaleAPIClient.PromoteProxyConfig(promoteProduct.Spec.ProductId, "sandbox", promoteProduct.Spec.PromoteVersion, promoteProduct.Spec.PromoteEnvironment)
				if err != nil {
					return ctrl.Result{}, err
				}
			}
		}
		if promoteProduct.Spec.PromoteEnvironment != "staging" && promoteProduct.Spec.PromoteEnvironment != "production" {
			reqLogger.Info("invalid environment selected, expected 'staging' or 'production' for promoteEnvironment")
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *PromoteProductReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.PromoteProduct{}).
		Complete(r)
}

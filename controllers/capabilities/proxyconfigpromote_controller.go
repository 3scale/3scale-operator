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
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"
)

// ProxyConfigPromoteReconciler reconciles a ProxyConfigPromote object
type ProxyConfigPromoteReconciler struct {
	*reconcilers.BaseReconciler
}

// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=proxyconfigpromotes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=proxyconfigpromotes/status,verbs=get;update;patch

func (r *ProxyConfigPromoteReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Logger().WithValues("proxyconfigpromote", req.NamespacedName)
	reqLogger.Info("Reconcile Proxy Config", "Operator version", version.Version)

	proxyConfigPromote := &capabilitiesv1beta1.ProxyConfigPromote{}
	err := r.Client().Get(r.Context(), req.NamespacedName, proxyConfigPromote)
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
		jsonData, err := json.MarshalIndent(proxyConfigPromote, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}
	//if proxyConfigPromote.Status.State != "Completed" {
	if !proxyConfigPromote.Status.Conditions.IsTrueFor("Ready") {
		statusReconciler, reconcileErr := r.proxyConfigPromoteReconciler(proxyConfigPromote, reqLogger, req)
		statusResult, statusUpdateErr := statusReconciler.Reconcile()
		if statusUpdateErr != nil {
			if reconcileErr != nil {
				return ctrl.Result{}, fmt.Errorf("Failed to reconcile activedoc: %v. Failed to update activedoc status: %w", reconcileErr, statusUpdateErr)
			}

			return ctrl.Result{}, fmt.Errorf("Failed to update activedoc status: %w", statusUpdateErr)
		}

		if statusResult.Requeue {
			return statusResult, nil
		}
	}
	if proxyConfigPromote.Spec.DeleteCR != nil && proxyConfigPromote.Status.Conditions.IsTrueFor("Ready") {
		err := r.DeleteResource(proxyConfigPromote)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ProxyConfigPromoteReconciler) proxyConfigPromoteReconciler(proxyConfigPromote *capabilitiesv1beta1.ProxyConfigPromote, reqLogger logr.Logger, req ctrl.Request) (*ProxyConfigPromoteStatusReconciler, error) {

	var latestStagingVersion int
	var latestProductionVersion int
	//get product
	product := &capabilitiesv1beta1.Product{}
	projectMeta := types.NamespacedName{
		Name:      proxyConfigPromote.Spec.ProductCRName,
		Namespace: req.Namespace,
	}

	err := r.Client().Get(r.Context(), projectMeta, product)
	if err != nil {
		if errors.IsNotFound(err) {
			statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", "", 0, 0, err)
			reqLogger.Info("resource not found. Ignoring since object must have been deleted")
			return statusReconciler, nil
		}
		statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", "", 0, 0, err)
		return statusReconciler, err
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), proxyConfigPromote.GetNamespace(), product.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", "", 0, 0, err)
		return statusReconciler, err
	}

	// connect to the 3scale porta client
	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", "", 0, 0, err)
		return statusReconciler, err
	}
	productList, err := threescaleAPIClient.ListProducts()
	productID := FindServiceBySystemName(*productList, product.Spec.SystemName)

	if productID > 0 {
		productIDStr := strconv.Itoa(int(productID))

		if proxyConfigPromote.Spec.Production == nil {
			// check the existing config to get the lastUpdate time
			oldConfig, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "sandbox")
			newConfig, err := threescaleAPIClient.DeployProductProxy(*product.Status.ID)
			if err != nil {
				statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", productIDStr, 0, 0, err)
				return statusReconciler, err
			}
			// if the UpdateAt strings are the same return a failed status as no product config changes
			if oldConfig.ProxyConfig.Content.Proxy.UpdatedAt == newConfig.Element.UpdatedAt {
				err := fmt.Errorf("No update to product config,  returning failed status")
				statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", productIDStr, 0, 0, err)
				return statusReconciler, err
			}

			err = r.Client().Update(r.Context(), proxyConfigPromote)
			stageElement, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "sandbox")
			if err != nil {
				statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", productIDStr, 0, 0, err)
				return statusReconciler, err
			}
			latestStagingVersion = stageElement.ProxyConfig.Version

			productionElement, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "production")
			if err != nil {
				if !threescaleapi.IsNotFound(err) {
					statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", productIDStr, 0, latestStagingVersion, err)
					return statusReconciler, err
				}
			}
			latestProductionVersion = productionElement.ProxyConfig.Version

			statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Completed", productIDStr, latestProductionVersion, latestStagingVersion, err)
			return statusReconciler, nil
		}
		if *proxyConfigPromote.Spec.Production == true {
			_, err := threescaleAPIClient.DeployProductProxy(*product.Status.ID)
			if err != nil {
				reqLogger.Info("Error", "Config version already exists in stage, skipping promotion to stage ", err)
			}

			stageElement, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "sandbox")
			if err != nil {
				reqLogger.Info("Error while finding sandbox version")
				statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", productIDStr, 0, 0, err)
				return statusReconciler, err
			}
			latestStagingVersion = stageElement.ProxyConfig.Version

			productionElement, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "production")
			if err != nil {
				reqLogger.Info("Error while finding production version")
				if !threescaleapi.IsNotFound(err) {
					//status := r.proxyConfigPromoteStatus(productIDStr, proxyConfigPromote, err, reqLogger, "Failed")
					statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", productIDStr, 0, latestStagingVersion, err)
					return statusReconciler, err
				}
			}
			latestProductionVersion = productionElement.ProxyConfig.Version

			_, err = threescaleAPIClient.PromoteProxyConfig(productIDStr, "sandbox", strconv.Itoa(stageElement.ProxyConfig.Version), "production")
			if err != nil {
				statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", productIDStr, latestProductionVersion, latestStagingVersion, err)
				return statusReconciler, err
			}

			err = r.Client().Update(r.Context(), proxyConfigPromote)
		}
		statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Completed", productIDStr, latestProductionVersion, latestStagingVersion, err)

		return statusReconciler, nil
	}
	err = fmt.Errorf("ProxyPromoteConfig Failed to find Product ID")
	//status := r.proxyConfigPromoteStatus("", proxyConfigPromote, err, reqLogger, "Invalid")
	statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Invalid", "", 0, 0, err)
	return statusReconciler, err
}


func (r *ProxyConfigPromoteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.ProxyConfigPromote{}).
		Complete(r)
}

func FindServiceBySystemName(list threescaleapi.ProductList, systemName string) int64 {
	for idx := range list.Products {
		if list.Products[idx].Element.SystemName == systemName {
			return list.Products[idx].Element.ID
		}
	}
	return -1
}


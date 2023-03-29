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
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	capabilitiesv1beta2 "github.com/3scale/3scale-operator/apis/capabilities/v1beta2"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ProxyConfigPromoteReconciler reconciles a ProxyConfigPromote object
type ProxyConfigPromoteReconciler struct {
	*reconcilers.BaseReconciler
}

// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=proxyconfigpromotes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=proxyconfigpromotes/status,verbs=get;update;patch

func (r *ProxyConfigPromoteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
	// get product
	product := &capabilitiesv1beta2.Product{}
	projectMeta := types.NamespacedName{
		Name:      proxyConfigPromote.Spec.ProductCRName,
		Namespace: req.Namespace,
	}

	err = r.Client().Get(r.Context(), projectMeta, product)
	if err != nil {
		if errors.IsNotFound(err) {
			statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", "product not found", 0, 0, err)
			statusReconciler.Reconcile()
			reqLogger.Info("product not found. Ignoring since object must have been deleted")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	// get providerAccountRef from product
	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), proxyConfigPromote.GetNamespace(), product.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		return ctrl.Result{}, err
	}

	// connect to the 3scale porta client
	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		return ctrl.Result{}, err
	}

	//if proxyConfigPromote.Status.State != "Completed" {
	if !proxyConfigPromote.Status.Conditions.IsTrueFor(capabilitiesv1beta1.ProxyPromoteConfigReadyConditionType) {
		statusReconciler, reconcileErr := r.proxyConfigPromoteReconciler(proxyConfigPromote, reqLogger, threescaleAPIClient, product)
		statusResult, statusUpdateErr := statusReconciler.Reconcile()
		if statusUpdateErr != nil {
			if reconcileErr != nil {
				return ctrl.Result{}, fmt.Errorf("Failed to reconcile proxyConfigPromote: %v. Failed to update proxyConfigPromote status: %w", reconcileErr, statusUpdateErr)
			}

			return ctrl.Result{}, fmt.Errorf("Failed to update proxyConfigPromote status: %w", statusUpdateErr)
		}

		if statusResult.Requeue {
			return statusResult, nil
		}
	}
	if (proxyConfigPromote.Spec.DeleteCR != nil && *proxyConfigPromote.Spec.DeleteCR) && proxyConfigPromote.Status.Conditions.IsTrueFor(capabilitiesv1beta1.ProxyPromoteConfigReadyConditionType) {
		err := r.DeleteResource(proxyConfigPromote)
		if err != nil && !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ProxyConfigPromoteReconciler) proxyConfigPromoteReconciler(proxyConfigPromote *capabilitiesv1beta1.ProxyConfigPromote, reqLogger logr.Logger, threescaleAPIClient *threescaleapi.ThreeScaleClient, product *capabilitiesv1beta2.Product) (*ProxyConfigPromoteStatusReconciler, error) {

	var latestStagingVersion int
	var latestProductionVersion int
	//get product

	if product.Status.Conditions.IsTrueFor(capabilitiesv1beta2.ProductSyncedConditionType) {
		productID := product.Status.ID
		productIDInt64 := *productID
		productIDStr := strconv.Itoa(int(productIDInt64))

		if proxyConfigPromote.Spec.Production == nil {
			// check the existing config to get the lastUpdate time
			_, err := threescaleAPIClient.DeployProductProxy(*product.Status.ID)
			if err != nil {
				statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", productIDStr, 0, 0, err)
				return statusReconciler, err
			}

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
		if *proxyConfigPromote.Spec.Production {
			_, err := threescaleAPIClient.DeployProductProxy(*product.Status.ID)
			if err != nil {
				reqLogger.Info("Error", "Config version already exists in stage, skipping promotion to stage ", err)
				statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", productIDStr, 0, 0, err)
				return statusReconciler, err
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
			} else {
				latestProductionVersion = latestStagingVersion
			}

		}
		statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Completed", productIDStr, latestProductionVersion, latestStagingVersion, nil)
		return statusReconciler, nil
	} else {
		err := fmt.Errorf("Proudct CR is not ready")
		statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "Failed", "", 0, 0, err)
		return statusReconciler, err
	}
}

func (r *ProxyConfigPromoteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.ProxyConfigPromote{}).
		Complete(r)
}

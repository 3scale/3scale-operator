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
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
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

	status, err := r.proxyConfigPromoteReconciler(proxyConfigPromote, reqLogger, req)
	reqLogger.WithValues("status", status)

	return ctrl.Result{}, nil
}

func (r *ProxyConfigPromoteReconciler) proxyConfigPromoteReconciler(proxyConfigPromote *capabilitiesv1beta1.ProxyConfigPromote, reqLogger logr.Logger, req ctrl.Request) (capabilitiesv1beta1.ProxyConfigPromoteStatus, error) {

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), proxyConfigPromote.GetNamespace(), proxyConfigPromote.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		status := r.proxyConfigPromoteStatus("", proxyConfigPromote, err, reqLogger, "Failed")
		return status, err
	}

	// connect to the 3scale porta client
	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		status := r.proxyConfigPromoteStatus("", proxyConfigPromote, err, reqLogger, "Failed")
		return status, err
	}

	// Find product ID
	productList, err := controllerhelper.ProductList(req.Namespace, r.Client(), providerAccount.AdminURLStr, reqLogger)
	productID := controllerhelper.FindProductBySystemName(productList, proxyConfigPromote.Spec.ProductName)

	productIDStr := strconv.Itoa(productID)
	if productIDStr == "" {
		reqLogger.Info("product ID failed to convert to string")
		status := r.proxyConfigPromoteStatus("", proxyConfigPromote, err, reqLogger, "Failed")
		return status, err
	}

	if productID == -1 {
		reqLogger.Info("name doesnt corresponding to a valid product")
		status := r.proxyConfigPromoteStatus("", proxyConfigPromote, err, reqLogger, "Failed")
		return status, err
	}

	// check if promotion is enabled in the CR
	if productID != -1 {
		productIdInt64 := int64(productID)

		if proxyConfigPromote.Spec.Production == false {
			_, err = threescaleAPIClient.DeployProductProxy(productIdInt64)
			if err != nil {
				status := r.proxyConfigPromoteStatus(productIDStr, proxyConfigPromote, err, reqLogger, "Failed")
				return status, err
			}
			err = r.Client().Update(r.Context(), proxyConfigPromote)
		}
		if proxyConfigPromote.Spec.Production == true {
			if proxyConfigPromote.Spec.PromoteVersion == "" {
				reqLogger.Info("empty configuration version selected, expected a number string for promoteVersion")
				status := r.proxyConfigPromoteStatus(productIDStr, proxyConfigPromote, err, reqLogger, "Failed")
				return status, err
			}
			if proxyConfigPromote.Spec.PromoteVersion != "" {
				_, err = threescaleAPIClient.PromoteProxyConfig(productIDStr, "sandbox", proxyConfigPromote.Spec.PromoteVersion, "production")
				if err != nil {
					status := r.proxyConfigPromoteStatus(productIDStr, proxyConfigPromote, err, reqLogger, "Failed")
					return status, err
				}
			}
			err = r.Client().Update(r.Context(), proxyConfigPromote)
		}
	}
	status := r.proxyConfigPromoteStatus(productIDStr, proxyConfigPromote, err, reqLogger, "Completed")
	return status, nil
}

func (r *ProxyConfigPromoteReconciler) proxyConfigPromoteStatus(productID string, promoteProxyConfig *capabilitiesv1beta1.ProxyConfigPromote, err error, reqLogger logr.Logger, state string) capabilitiesv1beta1.ProxyConfigPromoteStatus {
	promoteProxyConfig.Status.ProductId = productID

	if promoteProxyConfig.Spec.Production == true {
		promoteProxyConfig.Status.PromoteEnvironment = "production"
	} else {
		promoteProxyConfig.Status.PromoteEnvironment = "staging"
	}

	promoteProxyConfig.Status.State = state
	err = r.Client().Status().Update(r.Context(), promoteProxyConfig)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("failed to update the status in promoteProduct", promoteProxyConfig.Name)
		}

	}

	return promoteProxyConfig.Status
}

func (r *ProxyConfigPromoteReconciler) addPromotionAnnotationToProduct() {

}

func (r *ProxyConfigPromoteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.ProxyConfigPromote{}).
		Complete(r)
}

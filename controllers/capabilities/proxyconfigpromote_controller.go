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
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"
	"strings"
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

	if proxyConfigPromote.Status.State != "Completed" {
		status, err := r.proxyConfigPromoteReconciler(proxyConfigPromote, reqLogger, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.WithValues("status", status)
	}

	//err = r.addAnnotations(proxyConfigPromote.Status.PromoteEnvironment, proxyConfigPromote.Name, req.NamespacedName, req.Namespace, reqLogger)
	//if err != nil {
	//	return ctrl.Result{}, err
	//}

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
	productList, err := threescaleAPIClient.ListProducts()
	productID := FindServiceBySystemName(*productList, proxyConfigPromote.Spec.SystemName)

	productIDStr := strconv.Itoa(int(productID))

	if productID == -1 {
		reqLogger.Info("name doesnt correspond to a valid product")
		status := r.proxyConfigPromoteStatus("", proxyConfigPromote, err, reqLogger, "Failed")
		return status, err
	}

	// check if promotion is enabled in the CR
	if productID != -1 {
		var flag bool

		if proxyConfigPromote.Spec.Production == false {
			_, err = threescaleAPIClient.DeployProductProxy(productID)
			if err != nil {
				status := r.proxyConfigPromoteStatus(productIDStr, proxyConfigPromote, err, reqLogger, "Failed")
				return status, err
			}
			err = r.Client().Update(r.Context(), proxyConfigPromote)
		}
		if proxyConfigPromote.Spec.Production == true {
			//change this
			_, err = threescaleAPIClient.DeployProductProxy(productID)
			if err != nil {
				reqLogger.Info("Config version already exists in stage, skipping promotion to stage ", err)
			}

			stageElement, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "sandbox")
			if err != nil {
				reqLogger.Info("Error while finding sandbox version")
				status := r.proxyConfigPromoteStatus(productIDStr, proxyConfigPromote, err, reqLogger, "Failed")
				return status, err
			}

			productionElement, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "production")
			if err != nil {
				reqLogger.Info("Error while finding production version")
				status := r.proxyConfigPromoteStatus(productIDStr, proxyConfigPromote, err, reqLogger, "Failed")
				if !strings.Contains(err.Error(), "Not found") {
					return status, err
				}
				flag = true
			}

			if stageElement.ProxyConfig.Version != productionElement.ProxyConfig.Version || flag {
				_, err = threescaleAPIClient.PromoteProxyConfig(productIDStr, "sandbox", strconv.Itoa(stageElement.ProxyConfig.Version), "production")
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

func FindServiceBySystemName(list threescaleapi.ProductList, systemName string) int64 {
	for idx := range list.Products {
		if list.Products[idx].Element.SystemName == systemName {
			return list.Products[idx].Element.ID
		}
	}
	return -1
}

//func (r *ProxyConfigPromoteReconciler) addAnnotations(env string, name string, namespace types.NamespacedName, strNamespace string, reqLogger logr.Logger) error {
//
//	product := &capabilitiesv1beta1.Product{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      name,
//			Namespace: strNamespace,
//		},
//	}
//
//	reqLogger.Info("this is our product", product)
//
//	_, err := controllerutil.CreateOrUpdate(r.Context(), r.Client(), product, func() error {
//		annotations := product.ObjectMeta.GetAnnotations()
//		if annotations == nil {
//			annotations = map[string]string{}
//		}
//		annotations["latest promotion"] = env
//		product.ObjectMeta.SetAnnotations(annotations)
//
//		return nil
//	})
//	if err != nil {
//		reqLogger.Info("Failed to update product annotations", err)
//		return err
//	}
//	return nil
//}

func (r *ProxyConfigPromoteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.ProxyConfigPromote{}).
		Complete(r)
}

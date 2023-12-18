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
	"strings"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
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
			// Return and don't requeue
			reqLogger.Info("resource not found. Ignoring since object must have been deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(proxyConfigPromote, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	product := &capabilitiesv1beta1.Product{}

	// Retrieve product CR, on failed retrieval update status and requeue
	err = r.Client().Get(r.Context(), types.NamespacedName{Name: proxyConfigPromote.Spec.ProductCRName, Namespace: proxyConfigPromote.Namespace}, product)
	if err != nil {
		// If the product CR is not found, update status and requeue
		if errors.IsNotFound(err) {
			statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "product not found", 0, 0, err)
			reqLogger.Info("product not found. Ignoring since object must have been deleted")
			statusResult, statusErr := statusReconciler.Reconcile()
			// Reconcile status first as the reconcilerError might need to be updated to the status section of the CR before requeueing
			if statusErr != nil {
				return ctrl.Result{}, statusErr
			}
			if statusResult.Requeue {
				reqLogger.Info("Reconciling status not finished. Requeueing.")
				return statusResult, nil
			}
		}

		// If API call error, return err
		return ctrl.Result{}, err
	}

	// Retrieve providerAccountRef
	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), proxyConfigPromote.GetNamespace(), product.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		return ctrl.Result{}, err
	}

	// connect to the 3scale porta client
	insecureSkipVerify := controllerhelper.GetInsecureSkipVerifyAnnotation(proxyConfigPromote.GetAnnotations())
	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount, insecureSkipVerify)
	if err != nil {
		return ctrl.Result{}, err
	}

	// reconcile spec of the proxyConfigPromote only if the CR isn't already marked as "Completed" since the CR is a one off update.
	if !proxyConfigPromote.Status.Conditions.IsTrueFor(capabilitiesv1beta1.ProxyPromoteConfigReadyConditionType) {
		statusReconciler, reconcileErr := r.proxyConfigPromoteReconciler(proxyConfigPromote, reqLogger, threescaleAPIClient, product)

		// If status reconciler is not nil, proceed with reconciling the status
		if statusReconciler != nil {
			statusResult, statusErr := statusReconciler.Reconcile()
			// Reconcile status first as the reconcilerError might need to be updated to the status section of the CR before requeueing
			if statusErr != nil {
				return ctrl.Result{}, statusErr
			}
			if statusResult.Requeue {
				reqLogger.Info("Reconciling status not finished. Requeueing.")
				return statusResult, nil
			}
		}

		// If reconcile error but no status update required, requeue.
		if reconcileErr != nil {
			return ctrl.Result{}, reconcileErr
		}
	}

	if (proxyConfigPromote.Spec.DeleteCR != nil && *proxyConfigPromote.Spec.DeleteCR) && proxyConfigPromote.Status.Conditions.IsTrueFor(capabilitiesv1beta1.ProxyPromoteConfigReadyConditionType) {
		err := r.DeleteResource(proxyConfigPromote)
		if err != nil && !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	reqLogger.Info("Successfully reconciled")
	return ctrl.Result{}, nil
}

func (r *ProxyConfigPromoteReconciler) proxyConfigPromoteReconciler(proxyConfigPromote *capabilitiesv1beta1.ProxyConfigPromote, reqLogger logr.Logger, threescaleAPIClient *threescaleapi.ThreeScaleClient, product *capabilitiesv1beta1.Product) (*ProxyConfigPromoteStatusReconciler, error) {
	var latestStagingVersion int
	var latestProductionVersion int
	var currentStagingVersion int

	// Only proceed with proxyConfigPromote if status of the product is marked as "Completed"
	if product.Status.Conditions.IsTrueFor(capabilitiesv1beta1.ProductSyncedConditionType) {
		productID := product.Status.ID
		productIDInt64 := *productID
		productIDStr := strconv.Itoa(int(productIDInt64))

		// If wanting to promote to Stage but not production.
		if proxyConfigPromote.Spec.Production == nil || !*proxyConfigPromote.Spec.Production {

			// Fetch current stage version before promotion
			currentStageElement, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "sandbox")
			if err != nil {
				// If error exists staging is empty
				if strings.Contains(err.Error(), "error calling 3scale system - reason: {\"status\":\"Not found\"} - code: 404") {
					// Log and treat as an empty staging environment
					reqLogger.Info("Staging environment is empty")
					latestStagingVersion = 0 // Setting default version for an empty staging environment
				} else {
					statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, productIDStr, 0, 0, err)
					return statusReconciler, err
				}
			} else {
				currentStagingVersion = currentStageElement.ProxyConfig.Version
			}

			// Promote to stage
			_, err = threescaleAPIClient.DeployProductProxy(*productID)
			if err != nil {
				statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, productIDStr, 0, 0, err)
				return statusReconciler, err
			}

			// Retrieve latest stage version
			stageElement, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "sandbox")
			if err != nil {
				statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, productIDStr, 0, 0, err)
				return statusReconciler, err
			}
			latestStagingVersion = stageElement.ProxyConfig.Version

			// Fetch production state, if product has not been promoted to production yet the state would be 0.
			productionElement, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "production")
			if err != nil {
				if !threescaleapi.IsNotFound(err) {
					statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, productIDStr, 0, latestStagingVersion, err)
					return statusReconciler, err
				}
			}
			latestProductionVersion = productionElement.ProxyConfig.Version
			
			// Compare the version before and after promotion
			if currentStagingVersion == latestStagingVersion {
				// If no changes have been applied, return an error
				err := fmt.Errorf("can't promote to staging as no product changes detected. Delete this proxyConfigPromote CR, then introduce changes to configuration, and then create a new proxyConfigPromote CR")
				statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, productIDStr, latestStagingVersion, 0, err)
				return statusReconciler, err
			}
		} else {
			// Spec.production not nil, if production value is true promote to production, in all other cases skip.
			if *proxyConfigPromote.Spec.Production {
				// Before promoting to Production we want to update latest changes to staging first
				_, err := threescaleAPIClient.DeployProductProxy(*product.Status.ID)
				if err != nil {
					reqLogger.Info("Error", "Config version already exists in stage, skipping promotion to stage ", err)
					statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, productIDStr, 0, 0, err)
					return statusReconciler, err
				}

				// Retrieving latest stage version
				stageElement, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "sandbox")
				if err != nil {
					reqLogger.Info("Error while finding sandbox version")
					statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, productIDStr, 0, 0, err)
					return statusReconciler, err
				}
				latestStagingVersion = stageElement.ProxyConfig.Version

				// Retrieving latest production version
				productionElement, err := threescaleAPIClient.GetLatestProxyConfig(productIDStr, "production")
				if err != nil {
					reqLogger.Info("Error while finding production version")
					if !threescaleapi.IsNotFound(err) {
						statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, productIDStr, 0, latestStagingVersion, err)
						return statusReconciler, err
					}
				}
				latestProductionVersion = productionElement.ProxyConfig.Version

				// Promoting staging latest to production
				_, err = threescaleAPIClient.PromoteProxyConfig(productIDStr, "sandbox", strconv.Itoa(stageElement.ProxyConfig.Version), "production")
				if err != nil {
					// The version can already be in the production meaning that it can't be updated again, the proxyPromote is not going to be deleted by the operator but instead, will notify the user of the issue
					statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, productIDStr, latestProductionVersion, latestStagingVersion, fmt.Errorf("can't promote to production as no product changes detected. Delete this proxyConfigPromote CR, then introduce changes to configuration, and then create a new proxyConfigPromote CR"))
					return statusReconciler, err
				} else {
					latestProductionVersion = latestStagingVersion
				}
			}
		}

		statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, productIDStr, latestProductionVersion, latestStagingVersion, nil)
		return statusReconciler, nil
	} else {
		// If product CR is not ready, update the status and requeue based on err.
		reqLogger.Info("product CR is not ready")
		err := fmt.Errorf("product CR is not ready")
		statusReconciler := NewProxyConfigPromoteStatusReconciler(r.BaseReconciler, proxyConfigPromote, "", 0, 0, err)
		return statusReconciler, err
	}
}

func (r *ProxyConfigPromoteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.ProxyConfigPromote{}).
		Complete(r)
}

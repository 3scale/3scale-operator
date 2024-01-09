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
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const billingFinalizer = "Billing.capabilities.3scale.net/finalizer"

// BillingReconciler reconciles a Billing object
type BillingReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that BillingReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &BillingReconciler{}

// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=Billings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=Billings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=Billings/finalizers,verbs=get;list;watch;create;update;patch;delete

func (r *BillingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Logger().WithValues("Billing", req.NamespacedName)
	reqLogger.Info("Reconcile Billing", "Operator version", version.Version)

	// Fetch the instance
	BillingCR := &capabilitiesv1beta1.Billing{}
	err := r.Client().Get(context.TODO(), req.NamespacedName, BillingCR)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("resource not found. Ignoring since object must have been deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(BillingCR, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Billing has been marked for deletion
	if BillingCR.GetDeletionTimestamp() != nil && controllerutil.ContainsFinalizer(BillingCR, billingFinalizer) {
		err = r.removeBillingFrom3scale(BillingCR)
		if err != nil {
			r.EventRecorder().Eventf(BillingCR, corev1.EventTypeWarning, "Failed to delete Billing", "%v", err)
			return ctrl.Result{}, err
		}

		controllerutil.RemoveFinalizer(BillingCR, billingFinalizer)
		err = r.UpdateResource(BillingCR)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Ignore deleted resource, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if BillingCR.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(BillingCR, billingFinalizer) {
		controllerutil.AddFinalizer(BillingCR, billingFinalizer)
		err = r.UpdateResource(BillingCR)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), BillingCR.GetNamespace(), BillingCR.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		return ctrl.Result{}, err
	}

	// Retrieve ownersReference of tenant CR that owns the Billing CR
	tenantCR, err := controllerhelper.RetrieveTenantCR(providerAccount, r.Client(), r.Logger(), BillingCR.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	// If tenant CR is found, set it's ownersReference as ownerReference in the BillingCR
	if tenantCR != nil {
		updated, err := r.EnsureOwnerReference(tenantCR, BillingCR)
		if err != nil {
			return ctrl.Result{}, err
		}

		if updated {
			err := r.Client().Update(r.Context(), BillingCR)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	}

	statusReconciler, reconcileErr := r.reconcileSpec(BillingCR, reqLogger)
	statusResult, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		if reconcileErr != nil {
			return ctrl.Result{}, fmt.Errorf("Failed to reconcile Billing: %v. Failed to update status: %w", reconcileErr, statusUpdateErr)
		}

		return ctrl.Result{}, fmt.Errorf("Failed to update Billing status: %w", statusUpdateErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if reconcileErr != nil {
		if helper.IsInvalidSpecError(reconcileErr) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(BillingCR, corev1.EventTypeWarning, "Invalid Billing spec", "%v", reconcileErr)
			return ctrl.Result{}, nil
		}

		if helper.IsWaitError(reconcileErr) {
			// On wait error, retry
			reqLogger.Info("retrying", "reason", reconcileErr)
			return ctrl.Result{Requeue: true}, nil
		}

		reqLogger.Error(reconcileErr, "Failed to reconcile")
		r.EventRecorder().Eventf(BillingCR, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
		return ctrl.Result{}, reconcileErr
	}

	return ctrl.Result{}, nil
}

func (r *BillingReconciler) reconcileSpec(billingCR *capabilitiesv1beta1.Billing, logger logr.Logger) (*BillingStatusReconciler, error) {
	err := r.validateSpec(billingCR)
	if err != nil {
		statusReconciler := NewBillingStatusReconciler(r.BaseReconciler, billingCR, "", err)
		return statusReconciler, err
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), billingCR.Namespace, billingCR.Spec.ProviderAccountRef, logger)
	if err != nil {
		statusReconciler := NewBillingStatusReconciler(r.BaseReconciler, billingCR, "", err)
		return statusReconciler, err
	}

	statusReconciler := NewBillingStatusReconciler(r.BaseReconciler, billingCR, providerAccount.AdminURLStr, err)
	return statusReconciler, err
}

func (r *BillingReconciler) validateSpec(resource *capabilitiesv1beta1.Billing) error {
	err := field.ErrorList{}
	err = append(err, resource.Validate()...)

	if len(err) == 0 {
		return nil
	}

	return &helper.SpecFieldError{
		ErrorType:      helper.InvalidError,
		FieldErrorList: err,
	}
}

func (r *BillingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.Billing{}).
		Complete(r)
}

func (r *BillingReconciler) removeBillingFrom3scale(BillingCR *capabilitiesv1beta1.Billing) error {
	logger := r.Logger().WithValues("Billing", client.ObjectKey{Name: BillingCR.Name, Namespace: BillingCR.Namespace})

	// Attempt to remove Billing only if BillingCR.Status.ID is present
	if BillingCR.Status.ID == nil {
		logger.Info("could not remove Billing because ID is missing in status")
		return nil
	}

	//providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), BillingCR.Namespace, BillingCR.Spec.ProviderAccountRef, r.Logger())
	//if err != nil {
	//	if apierrors.IsNotFound(err) {
	//		logger.Info("billing not deleted from 3scale, provider account not found")
	//		return nil
	//	}
	//	return err
	//}

	return nil
}

func (r *BillingReconciler) runBillingForTenant(BillingCR *capabilitiesv1beta1.Billing, TenantAccountID int64) error {
	//Triggers billing process for all developer accounts.
	//curl -X 'POST' \
	//'https://master.apps.vmo01.giq5.s1.devshift.org/master/api/providers/7/billing_jobs.xml' \
	//-H 'accept: */*' \
	//-H 'Content-Type: application/x-www-form-urlencoded' \
	//-d 'access_token=h3rttHvN&date=2024-01-07'

	return nil
}

func (r *BillingReconciler) runBillingForAllTenantsMonthly(threeScaleClient *threescaleapi.ThreeScaleClient, BillingCR *capabilitiesv1beta1.Billing) error {
	// Get list of tenants accounts and trigger billing request for each tenant
	//Triggers billing process for all developer accounts.
	//curl -X 'POST' \
	//'https://master.apps.vmo01.giq5.s1.devshift.org/master/api/providers/7/billing_jobs.xml' \
	//-H 'accept: */*' \
	//-H 'Content-Type: application/x-www-form-urlencoded' \
	//-d 'access_token=h3rttHvN&date=2024-01-07'

	if BillingCR.Spec.BillingConfig.BillDayOfMonth != time.Now().Day() {
		return nil
	}

	listDeveloperAccounts, err := threeScaleClient.ListDeveloperAccounts()
	if err != nil {
		return err
	}
	masterRouteUrl, err := r.getMasterRouteUrl(r.Client(), BillingCR)
	// devAccount == Tenant Account ? TODO
	for _, devAccount := range listDeveloperAccounts.Items {
		if devAccount.Element.MonthlyBillingEnabled != nil && *devAccount.Element.MonthlyBillingEnabled {
			accountId := devAccount.Element.ID
			err := r.billTenant(threeScaleClient, masterRouteUrl, accountId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *BillingReconciler) billTenant(threeScaleClient *threescaleapi.ThreeScaleClient, masterRouteUrl string, tenantAccountId *int64) error {
	//Triggers billing process for all developer accounts.
	//curl -X 'POST' \
	//'https://master.apps.vmo01.giq5.s1.devshift.org/master/api/providers/7/billing_jobs.xml' \
	//-H 'accept: */*' \
	//-H 'Content-Type: application/x-www-form-urlencoded' \
	//-d 'access_token=h3rttHvN&date=2024-01-07'
	billDate := time.Now().Format("YYYY-MM-DD")

	requestUrl := masterRouteUrl + "/master/api/providers/" + strconv.FormatInt(*tenantAccountId, 10) + "/billing_jobs.xml"
	accessToken := helper.EnvVarFromSecret("MASTER_ACCESS_TOKEN", component.SystemSecretSystemSeedSecretName, component.SystemSecretSystemSeedMasterAccessTokenFieldName).Value

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("date", billDate)
	body := strings.NewReader(values.Encode())

	req, err := http.NewRequest("POST", requestUrl, body)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (r *BillingReconciler) getMasterRouteUrl(cl client.Client, billingCR *capabilitiesv1beta1.Billing) (string, error) {
	//get Wildcard Domain from API manage CR
	listOps := []client.ListOption{client.InNamespace(billingCR.Namespace)}
	apimanagerList := &appsv1alpha1.APIManagerList{}
	err := cl.List(context.TODO(), apimanagerList, listOps...)
	if err != nil {
		return "", err
	}
	if len(apimanagerList.Items) > 0 {
		return "", nil
	}
	apimanager := apimanagerList.Items[0]
	masterRoute := "https://master." + apimanager.Spec.WildcardDomain
	return masterRoute, nil
}

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
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	*reconcilers.BaseReconciler
}

const (
	// applicationIdAnnotation matches the application.statu.ID
	applicationIdAnnotation = "applicationID"

	applicationFinalizer = "application.capabilities.3scale.net/finalizer"
)

// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,resources=applications/status,verbs=get;update;patch

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Logger().WithValues("application", req.NamespacedName)
	reqLogger.Info("Reconcile Application", "Operator version", version.Version)

	application := &capabilitiesv1beta1.Application{}
	err := r.Client().Get(context.TODO(), req.NamespacedName, application)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info(" Application resource not found. Ignoring since object must have been deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(application, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}
	// get Account
	accountResource := &capabilitiesv1beta1.DeveloperAccount{}
	projectMeta := types.NamespacedName{
		Name:      application.Spec.AccountCR.Name,
		Namespace: req.Namespace,
	}

	err = r.Client().Get(r.Context(), projectMeta, accountResource)
	if err != nil {
		if apierrors.IsNotFound(err) {
			reqLogger.Error(err, "error developer account not found ")
			statusReconciler := NewApplicationStatusReconciler(r.BaseReconciler, application, nil, "", err)
			statusResult, statusUpdateErr := statusReconciler.Reconcile()
			if statusUpdateErr != nil {
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("Failed to reconcile application: %v. Failed to update application status: %w", err, statusUpdateErr)
				}

				return ctrl.Result{}, fmt.Errorf("Failed to update applicatino status: %w", statusUpdateErr)
			}
			if statusResult.Requeue {
				return statusResult, nil
			}
		}
	}
	// get providerAccountRef from account
	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), accountResource.Namespace, accountResource.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		return ctrl.Result{}, err
	}

	insecureSkipVerify := controllerhelper.GetInsecureSkipVerifyAnnotation(application.GetAnnotations())
	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount, insecureSkipVerify)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Ignore deleted Applications, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if application.GetDeletionTimestamp() != nil && controllerutil.ContainsFinalizer(application, applicationFinalizer) {
		err = r.removeApplicationFrom3scale(application, req, *threescaleAPIClient)
		if err != nil {
			r.EventRecorder().Eventf(application, corev1.EventTypeWarning, "Failed to delete application", "%v", err)
			return ctrl.Result{}, err
		}

		controllerutil.RemoveFinalizer(application, applicationFinalizer)
		err = r.UpdateResource(application)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if application.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	metadataUpdated := r.reconcileMetadata(application)
	if metadataUpdated {
		err = r.UpdateResource(application)
		if err != nil {
			return ctrl.Result{}, err
		}

		// No need requeue because the reconcile will trigger automatically since updating the Application CR
		return ctrl.Result{}, nil
	}

	statusReconciler, reconcileErr := r.applicationReconciler(application, req, threescaleAPIClient, providerAccount.AdminURLStr, accountResource)
	statusResult, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		if reconcileErr != nil {
			return ctrl.Result{}, fmt.Errorf("Failed to sync application: %v. Failed to update application status: %w", reconcileErr, statusUpdateErr)
		}

		return ctrl.Result{}, fmt.Errorf("Failed to update application status: %w", statusUpdateErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if reconcileErr != nil {
		if helper.IsInvalidSpecError(reconcileErr) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(application, corev1.EventTypeWarning, "Invalid application Spec", "%v", reconcileErr)
			return ctrl.Result{}, nil
		}

		if helper.IsOrphanSpecError(reconcileErr) {
			// On Orphan spec error, retry
			reqLogger.Info("ERROR", "spec orphan error", reconcileErr)
			return ctrl.Result{Requeue: true}, nil
		}

		reqLogger.Error(reconcileErr, "Failed to reconcile")
		r.EventRecorder().Eventf(application, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
		return ctrl.Result{}, reconcileErr
	}

	reqLogger.Info("END", "error", reconcileErr)

	return ctrl.Result{}, nil
}

func (r *ApplicationReconciler) reconcileMetadata(application *capabilitiesv1beta1.Application) bool {
	changed := false

	// If the application.Status.ID is found and the annotation is not found - create
	// If the application.Status.ID is found and the annotation is found but, the value of annotation is different to the application.Status.ID - update
	var applicationId int64 = 0
	if application.Status.ID != nil {
		applicationId = *application.Status.ID
	}
	if value, found := application.ObjectMeta.Annotations[applicationIdAnnotation]; (applicationId != 0 && !found) || (applicationId != 0 && found && value != strconv.FormatInt(*application.Status.ID, 10)) {
		if application.ObjectMeta.Annotations == nil {
			application.ObjectMeta.Annotations = make(map[string]string)
		}
		application.ObjectMeta.Annotations[applicationIdAnnotation] = strconv.FormatInt(*application.Status.ID, 10)
		changed = true
	}

	if !controllerutil.ContainsFinalizer(application, applicationFinalizer) {
		controllerutil.AddFinalizer(application, applicationFinalizer)
		changed = true
	}

	return changed
}

func (r *ApplicationReconciler) applicationReconciler(applicationResource *capabilitiesv1beta1.Application, req ctrl.Request, threescaleAPIClient *threescaleapi.ThreeScaleClient, providerAccountAdminURLStr string, accountResource *capabilitiesv1beta1.DeveloperAccount) (*ApplicationStatusReconciler, error) {

	// get product
	productResource := &capabilitiesv1beta1.Product{}
	projectMeta := types.NamespacedName{
		Name:      applicationResource.Spec.ProductCR.Name,
		Namespace: req.Namespace,
	}

	err := r.Client().Get(r.Context(), projectMeta, productResource)
	if err != nil {
		if apierrors.IsNotFound(err) {
			statusReconciler := NewApplicationStatusReconciler(r.BaseReconciler, applicationResource, nil, "", err)
			return statusReconciler, err
		}
	}

	err = r.checkExternalResources(applicationResource, accountResource, productResource)
	if err != nil {
		statusReconciler := NewApplicationStatusReconciler(r.BaseReconciler, applicationResource, nil, "", err)
		return statusReconciler, err
	}

	reconciler := NewApplicationReconciler(r.BaseReconciler, applicationResource, accountResource, productResource, threescaleAPIClient)
	ApplicationEntity, err := reconciler.Reconcile()
	if err != nil {
		statusReconciler := NewApplicationStatusReconciler(r.BaseReconciler, applicationResource, nil, providerAccountAdminURLStr, err)
		return statusReconciler, err
	}
	statusReconciler := NewApplicationStatusReconciler(r.BaseReconciler, applicationResource, ApplicationEntity, providerAccountAdminURLStr, err)
	return statusReconciler, err
}

func (r *ApplicationReconciler) removeApplicationFrom3scale(application *capabilitiesv1beta1.Application, req ctrl.Request, threescaleAPIClient threescaleapi.ThreeScaleClient) error {
	logger := r.Logger().WithValues("application", client.ObjectKey{Name: application.Name, Namespace: application.Namespace})

	// get Account
	account := &capabilitiesv1beta1.DeveloperAccount{}
	projectMeta := types.NamespacedName{
		Name:      application.Spec.AccountCR.Name,
		Namespace: req.Namespace,
	}

	err := r.Client().Get(r.Context(), projectMeta, account)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Account resource not found. Ignoring since object must have been deleted")
		}
	}

	// Attempt to remove application only if application.Status.ID is present
	if application.Status.ID == nil {
		logger.Info("could not remove application because ID is missing in status")
		return nil
	}

	if account.Status.ID == nil {
		logger.Info("could not remove application because ID is missing in the status of developer account")
		return fmt.Errorf("could not remove application because ID is missing in the satus of developer account %s", account.Name)
	}

	err = threescaleAPIClient.DeleteApplication(*account.Status.ID, *application.Status.ID)
	if err != nil && !threescaleapi.IsNotFound(err) {
		return err
	}

	return nil
}

func (r *ApplicationReconciler) checkExternalResources(applicationResource *capabilitiesv1beta1.Application, accountResource *capabilitiesv1beta1.DeveloperAccount, productResource *capabilitiesv1beta1.Product) error {
	errors := field.ErrorList{}

	specFldPath := field.NewPath("spec")
	accountFldPath := specFldPath.Child("accountCRName")
	productFldPath := specFldPath.Child("productCRName")

	if accountResource.Status.ID == nil {
		errors = append(errors, field.Invalid(accountFldPath, applicationResource.Spec.AccountCR, "accountCR name doesnt have a valid account reference"))
	}
	if productResource.Status.ID == nil {
		errors = append(errors, field.Invalid(productFldPath, applicationResource.Spec.ProductCR, "productCR name doesnt have a valid product reference"))
	}
	if accountResource.Status.Conditions.IsTrueFor(capabilitiesv1beta1.DeveloperAccountInvalidConditionType) {
		errors = append(errors, field.Invalid(accountFldPath, applicationResource.Spec.AccountCR, "account CR is in an invalid state"))
	}
	if productResource.Status.Conditions.IsTrueFor(capabilitiesv1beta1.ProductInvalidConditionType) {
		errors = append(errors, field.Invalid(productFldPath, applicationResource.Spec.ProductCR, "product CR is in an invalid state"))
	}

	if len(errors) == 0 && productResource.Status.ProviderAccountHost != accountResource.Status.ProviderAccountHost {
		errors = append(errors, field.Invalid(productFldPath, applicationResource.Spec.ProductCR, "product and account providerAccounts dont match"))
	}

	if len(errors) == 0 {
		return nil
	}

	return &helper.SpecFieldError{
		ErrorType:      helper.OrphanError,
		FieldErrorList: errors,
	}
}

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.Application{}).
		Complete(r)
}

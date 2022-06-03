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
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/handlers"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
)

// APIManagerReconciler reconciles a APIManager object
type APIManagerReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that APIManagerReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &APIManagerReconciler{}

func (r *APIManagerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	logger := r.BaseReconciler.Logger().WithValues("apimanager", req.NamespacedName)
	logger.Info("ReconcileAPIManager", "Operator version", version.Version, "3scale release", product.ThreescaleRelease)

	// your logic here

	instance, err := r.apiManagerInstance(req.NamespacedName)
	if err != nil {
		logger.Error(err, "Error fetching apimanager instance")
		return ctrl.Result{}, err
	}
	if instance == nil {
		logger.Info("resource not found. Ignoring since object must have been deleted")
		return ctrl.Result{}, nil
	}

	err = r.validateCR(instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	res, err := r.setAPIManagerDefaults(instance)
	if err != nil {
		logger.Error(err, "Error")
		return ctrl.Result{}, err
	}
	if res.Requeue {
		logger.Info("Defaults set for APIManager resource")
		return res, nil
	}

	if instance.Annotations[appsv1alpha1.OperatorVersionAnnotation] != version.Version {
		logger.Info(fmt.Sprintf("Upgrade %s -> %s", instance.Annotations[appsv1alpha1.OperatorVersionAnnotation], version.Version))
		// TODO add logic to check that only immediate consecutive installs
		// are possible?
		res, err := r.upgradeAPIManager(instance)
		if err != nil {
			logger.Error(err, "Error upgrading APIManager")
			return ctrl.Result{}, err
		}
		if res.Requeue {
			logger.Info("Upgrading not finished. Requeueing.")
			return res, nil
		}

		err = r.updateVersionAnnotations(instance)
		if err != nil {
			logger.Error(err, "Error updating annotations")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	specResult, specErr := r.reconcileAPIManagerLogic(instance)
	if specErr != nil && specResult.Requeue {
		logger.Info("Reconciling not finished. Requeueing.")
		return specResult, nil
	}

	// reconcile status regardless specErr
	statusResult, statusErr := r.reconcileAPIManagerStatus(instance)
	if statusErr != nil {
		return ctrl.Result{}, statusErr
	}

	if specErr != nil {
		return ctrl.Result{}, specErr
	}

	if statusResult.Requeue {
		logger.Info("Reconciling not finished. Requeueing.")
		return statusResult, nil
	}

	return ctrl.Result{}, nil
}

func (r *APIManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.APIManager{}).
		Owns(&appsv1.DeploymentConfig{}).
		Owns(&policyv1beta1.PodDisruptionBudget{}).
		Watches(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: &handlers.APIManagerRoutesEventMapper{
				K8sClient: r.Client(),
				Logger:    r.Logger().WithName("APIManagerRoutesHandler"),
			},
		}).
		Complete(r)
}

func (r *APIManagerReconciler) validateCR(cr *appsv1alpha1.APIManager) error {
	fieldError := field.ErrorList{}
	// internal validation
	fieldError = append(fieldError, cr.Validate()...)

	fieldError = append(fieldError, r.validateApicastTLSCertificates(cr)...)

	if len(fieldError) > 0 {
		return fieldError.ToAggregate()
	}

	return nil
}

func (r *APIManagerReconciler) updateVersionAnnotations(cr *appsv1alpha1.APIManager) error {
	if cr.Annotations == nil {
		cr.Annotations = map[string]string{}
	}
	cr.Annotations[appsv1alpha1.ThreescaleVersionAnnotation] = product.ThreescaleRelease
	cr.Annotations[appsv1alpha1.OperatorVersionAnnotation] = version.Version
	return r.Client().Update(context.TODO(), cr)
}

func (r *APIManagerReconciler) upgradeAPIManager(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	// The object to instantiate would change in every release of the operator
	// that upgrades the threescale version
	upgradeAPIManager := operator.NewUpgradeApiManager(r.BaseReconciler, cr)
	return upgradeAPIManager.Upgrade()
}

func (r *APIManagerReconciler) apiManagerInstance(namespacedName types.NamespacedName) (*appsv1alpha1.APIManager, error) {
	// Fetch the APIManager instance
	instance := &appsv1alpha1.APIManager{}

	err := r.Client().Get(context.TODO(), namespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return nil, nil
		}
		return nil, err
	}
	return instance, nil
}

func (r *APIManagerReconciler) setAPIManagerDefaults(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	updated := cr.UpdateExternalComponentsFromHighAvailability()

	defaultsUpdated, err := cr.SetDefaults()
	if err != nil {
		return ctrl.Result{}, err
	}
	updated = updated || defaultsUpdated

	if updated {
		err = r.Client().Update(context.TODO(), cr)
	}

	return ctrl.Result{Requeue: updated}, err
}

func (r *APIManagerReconciler) reconcileAPIManagerLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseAPIManagerLogicReconciler := operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr)
	imageReconciler := operator.NewAMPImagesReconciler(baseAPIManagerLogicReconciler)
	result, err := imageReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	dependencyReconciler := r.dependencyReconcilerForComponents(cr, baseAPIManagerLogicReconciler)
	result, err = dependencyReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	backendReconciler := operator.NewBackendReconciler(baseAPIManagerLogicReconciler)
	result, err = backendReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	memcachedReconciler := operator.NewMemcachedReconciler(baseAPIManagerLogicReconciler)
	result, err = memcachedReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	systemReconciler := operator.NewSystemReconciler(baseAPIManagerLogicReconciler)
	result, err = systemReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	zyncReconciler := operator.NewZyncReconciler(baseAPIManagerLogicReconciler)
	result, err = zyncReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	apicastReconciler := operator.NewApicastReconciler(baseAPIManagerLogicReconciler)
	result, err = apicastReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	genericMonitoringReconciler := operator.NewGenericMonitoringReconciler(baseAPIManagerLogicReconciler)
	result, err = genericMonitoringReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	return ctrl.Result{}, nil
}

func (r *APIManagerReconciler) reconcileAPIManagerStatus(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	statusReconciler := NewAPIManagerStatusReconciler(r.BaseReconciler, cr)
	res, err := statusReconciler.Reconcile()
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("Failed to update APIManager status: %w", err)
	}

	return res, nil
}

func (r *APIManagerReconciler) validateApicastTLSCertificates(cr *appsv1alpha1.APIManager) field.ErrorList {
	fieldErrors := field.ErrorList{}

	if cr.Spec.Apicast != nil && cr.Spec.Apicast.ProductionSpec != nil && cr.Spec.Apicast.ProductionSpec.HTTPSCertificateSecretRef != nil {
		secretPath := field.NewPath("spec").Child("apicast").Child("productionSpec").Child("httpsCertificateSecretRef")
		if cr.Spec.Apicast.ProductionSpec.HTTPSCertificateSecretRef.Name == "" {
			fieldErrors = append(fieldErrors, field.Required(secretPath.Child("name"), "secret name not provided"))
		} else {
			nn := types.NamespacedName{
				Name:      cr.Spec.Apicast.ProductionSpec.HTTPSCertificateSecretRef.Name,
				Namespace: cr.Namespace,
			}
			err := helper.ValidateTLSSecret(nn, r.Client())
			if err != nil {
				fieldErrors = append(fieldErrors, field.Invalid(secretPath, cr.Spec.Apicast.ProductionSpec.HTTPSCertificateSecretRef, err.Error()))
			}
		}
	}

	if cr.Spec.Apicast != nil && cr.Spec.Apicast.StagingSpec != nil && cr.Spec.Apicast.StagingSpec.HTTPSCertificateSecretRef != nil {
		secretPath := field.NewPath("spec").Child("apicast").Child("stagingSpec").Child("httpsCertificateSecretRef")
		if cr.Spec.Apicast.StagingSpec.HTTPSCertificateSecretRef.Name == "" {
			fieldErrors = append(fieldErrors, field.Required(secretPath.Child("name"), "secret name not provided"))
		} else {
			nn := types.NamespacedName{
				Name:      cr.Spec.Apicast.StagingSpec.HTTPSCertificateSecretRef.Name,
				Namespace: cr.Namespace,
			}
			err := helper.ValidateTLSSecret(nn, r.Client())
			if err != nil {
				fieldErrors = append(fieldErrors, field.Invalid(secretPath, cr.Spec.Apicast.StagingSpec.HTTPSCertificateSecretRef, err.Error()))
			}
		}
	}

	return fieldErrors
}

func (r *APIManagerReconciler) dependencyReconcilerForComponents(cr *appsv1alpha1.APIManager, baseAPIManagerLogicReconciler *operator.BaseAPIManagerLogicReconciler) operator.DependencyReconciler {
	// Helper type that contains the constructors for a dependency reconciler
	// whether it's external or internal
	type constructors struct {
		External operator.DependencyReconcilerConstructor
		Internal operator.DependencyReconcilerConstructor
	}

	// Helper function that instantiates a dependency reconciler depending
	// on whether it's external or internal
	selectReconciler := func(cs constructors, selectIsExternal func(*appsv1alpha1.ExternalComponentsSpec) bool) operator.DependencyReconciler {
		constructor := cs.Internal
		if selectIsExternal(cr.Spec.ExternalComponents) {
			constructor = cs.External
		}

		return constructor(baseAPIManagerLogicReconciler)
	}

	// Select whether to use PostgreSQL or MySQL for the System database
	var systemDatabaseReconcilerConstructor operator.DependencyReconcilerConstructor
	if cr.Spec.System.DatabaseSpec != nil && cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		systemDatabaseReconcilerConstructor = operator.CompositeDependencyReconcilerConstructor(
			operator.NewSystemPostgreSQLReconciler,
			operator.NewSystemPostgreSQLImageReconciler,
		)
	} else {
		systemDatabaseReconcilerConstructor = operator.CompositeDependencyReconcilerConstructor(
			operator.NewSystemMySQLReconciler,
			operator.NewSystemMySQLImageReconciler,
		)
	}

	systemDatabaseConstructors := constructors{
		External: operator.NewSystemExternalDatabaseReconciler,
		Internal: systemDatabaseReconcilerConstructor,
	}
	systemRedisConstructors := constructors{
		External: operator.NewSystemExternalRedisReconciler,
		Internal: operator.NewSystemRedisDependencyReconciler,
	}
	backendRedisConstructors := constructors{
		External: operator.NewBackendExternalRedisReconciler,
		Internal: operator.NewBackendRedisDependencyReconciler,
	}

	// Build the final reconciler composed by the chosen external/internal
	// combination
	result := []operator.DependencyReconciler{
		selectReconciler(systemDatabaseConstructors, appsv1alpha1.SystemDatabase),
		selectReconciler(systemRedisConstructors, appsv1alpha1.SystemRedis),
		selectReconciler(backendRedisConstructors, appsv1alpha1.BackendRedis),
	}

	return &operator.CompositeDependencyReconciler{
		Reconcilers: result,
	}
}

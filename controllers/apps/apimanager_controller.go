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
	k8sappsv1 "k8s.io/api/apps/v1"

	routev1 "github.com/openshift/api/route/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimachinerymetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
	SecretLabelSelector apimachinerymetav1.LabelSelector
	WatchedNamespace    string
}

// blank assignment to verify that APIManagerReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &APIManagerReconciler{}

// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagers/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,namespace=placeholder,resources=pods;services;services/finalizers;replicationcontrollers;endpoints;persistentvolumeclaims;events;configmaps;secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,namespace=placeholder,resources=deployments;daemonsets;replicasets;statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,namespace=placeholder,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,namespace=placeholder,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=image.openshift.io,namespace=placeholder,resources=imagestreams;imagestreams/layers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=image.openshift.io,namespace=placeholder,resources=imagestreamtags,verbs=get;list;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,namespace=placeholder,resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,namespace=placeholder,resources=routes/custom-host,verbs=create
// +kubebuilder:rbac:groups=route.openshift.io,namespace=placeholder,resources=routes/status,verbs=get
// +kubebuilder:rbac:groups=apps.openshift.io,namespace=placeholder,resources=deploymentconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,namespace=placeholder,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,namespace=placeholder,resources=podmonitors;servicemonitors;prometheusrules,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=integreatly.org,namespace=placeholder,resources=grafanadashboards,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch

func (r *APIManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
		// TODO: Validation errors should also be represented on warning conditions. To be done when HPA feature branch merges with master.
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
	secretToApimanagerEventMapper := &SecretToApimanagerEventMapper{
		K8sClient: r.Client(),
		Logger:    r.Logger().WithName("secretToApimanagerEventMapper"),
		Namespace: r.WatchedNamespace,
	}

	handlers := &handlers.APIManagerRoutesEventMapper{
		K8sClient: r.Client(),
		Logger:    r.Logger().WithName("APIManagerRoutesHandler"),
	}

	labelSelectorPredicate, err := predicate.LabelSelectorPredicate(r.SecretLabelSelector)
	if err != nil {
		return nil
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.APIManager{}).
		Watches(
			&source.Kind{Type: &v1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(secretToApimanagerEventMapper.Map),
			builder.WithPredicates(labelSelectorPredicate),
		).
		Owns(&k8sappsv1.Deployment{}).
		Watches(&source.Kind{Type: &routev1.Route{}}, handler.EnqueueRequestsFromMapFunc(handlers.Map)).
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

	systemSearchdReconciler := operator.NewSystemSearchdReconciler(baseAPIManagerLogicReconciler)
	result, err = systemSearchdReconciler.Reconcile()
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

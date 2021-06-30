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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/handlers"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// APIManagerReconciler reconciles a APIManager object
type APIManagerReconciler struct {
	*reconcilers.BaseReconciler
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
	changed, err := cr.SetDefaults() // TODO check where to put this
	if err != nil {
		return ctrl.Result{}, err
	}

	if changed {
		err = r.Client().Update(context.TODO(), cr)
	}

	return ctrl.Result{Requeue: changed}, err
}

func (r *APIManagerReconciler) reconcileAPIManagerLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseAPIManagerLogicReconciler := operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr)
	imageReconciler := operator.NewAMPImagesReconciler(baseAPIManagerLogicReconciler)
	result, err := imageReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	if !cr.IsExternalDatabaseEnabled() {
		redisReconciler := operator.NewRedisReconciler(baseAPIManagerLogicReconciler)
		result, err = redisReconciler.Reconcile()
		if err != nil || result.Requeue {
			return result, err
		}

		result, err = r.reconcileSystemDatabaseLogic(cr, baseAPIManagerLogicReconciler)
		if err != nil || result.Requeue {
			return result, err
		}
	} else {
		// External databases
		haReconciler := operator.NewHighAvailabilityReconciler(baseAPIManagerLogicReconciler)
		result, err = haReconciler.Reconcile()
		if err != nil || result.Requeue {
			return result, err
		}
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

func (r *APIManagerReconciler) reconcileSystemDatabaseLogic(cr *appsv1alpha1.APIManager, baseAPIManagerLogicReconciler *operator.BaseAPIManagerLogicReconciler) (reconcile.Result, error) {
	if cr.Spec.System.DatabaseSpec != nil && cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return r.reconcileSystemPostgreSQLLogic(cr, baseAPIManagerLogicReconciler)
	}

	// Defaults to MySQL
	return r.reconcileSystemMySQLLogic(cr, baseAPIManagerLogicReconciler)
}

func (r *APIManagerReconciler) reconcileSystemPostgreSQLLogic(cr *appsv1alpha1.APIManager, baseAPIManagerLogicReconciler *operator.BaseAPIManagerLogicReconciler) (reconcile.Result, error) {
	reconciler := operator.NewSystemPostgreSQLReconciler(baseAPIManagerLogicReconciler)
	result, err := reconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	imageReconciler := operator.NewSystemPostgreSQLImageReconciler(baseAPIManagerLogicReconciler)
	result, err = imageReconciler.Reconcile()
	return result, err
}

func (r *APIManagerReconciler) reconcileSystemMySQLLogic(cr *appsv1alpha1.APIManager, baseAPIManagerLogicReconciler *operator.BaseAPIManagerLogicReconciler) (reconcile.Result, error) {
	reconciler := operator.NewSystemMySQLReconciler(baseAPIManagerLogicReconciler)
	result, err := reconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	imageReconciler := operator.NewSystemMySQLImageReconciler(baseAPIManagerLogicReconciler)
	result, err = imageReconciler.Reconcile()
	return result, err
}

func (r *APIManagerReconciler) reconcileAPIManagerStatus(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	statusReconciler := NewAPIManagerStatusReconciler(r.BaseReconciler, cr)
	res, err := statusReconciler.Reconcile()
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("Failed to update APIManager status: %w", err)
	}

	return res, nil
}

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
	"reflect"
	"sort"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	"github.com/RHsyseng/operator-utils/pkg/olm"
	appsv1 "github.com/openshift/api/apps/v1"
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

// +kubebuilder:rbac:groups=apps.3scale.net,resources=apimanagers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.3scale.net,resources=apimanagers/status,verbs=get;update;patch

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

	result, err := r.reconcileAPIManagerLogic(instance)
	if err != nil {
		logger.Error(err, "Error during reconciliation")
		return result, err
	}
	if result.Requeue {
		logger.Info("Reconciling not finished. Requeueing.")
		return result, nil
	}

	statusResult, err := r.reconcileAPIManagerStatus(instance)
	if err != nil {
		logger.Error(err, "Error updating status")
		return ctrl.Result{}, err
	}

	if statusResult.Requeue {
		logger.V(1).Info("Reconciling status not finished. Requeueing.")
		return statusResult, nil
	}

	return ctrl.Result{}, nil
}

func (r *APIManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.APIManager{}).
		Owns(&appsv1.DeploymentConfig{}).
		Owns(&policyv1beta1.PodDisruptionBudget{}).
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
	imageReconciler := operator.NewAMPImagesReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
	result, err := imageReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	if !cr.IsExternalDatabaseEnabled() {
		redisReconciler := operator.NewRedisReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
		result, err = redisReconciler.Reconcile()
		if err != nil || result.Requeue {
			return result, err
		}

		result, err = r.reconcileSystemDatabaseLogic(cr)
		if err != nil || result.Requeue {
			return result, err
		}
	} else {
		// External databases
		haReconciler := operator.NewHighAvailabilityReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
		result, err = haReconciler.Reconcile()
		if err != nil || result.Requeue {
			return result, err
		}
	}

	backendReconciler := operator.NewBackendReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
	result, err = backendReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	memcachedReconciler := operator.NewMemcachedReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
	result, err = memcachedReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	systemReconciler := operator.NewSystemReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
	result, err = systemReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	zyncReconciler := operator.NewZyncReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
	result, err = zyncReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	apicastReconciler := operator.NewApicastReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
	result, err = apicastReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	genericMonitoringReconciler := operator.NewGenericMonitoringReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
	result, err = genericMonitoringReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	return ctrl.Result{}, nil
}

func (r *APIManagerReconciler) reconcileSystemDatabaseLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	if cr.Spec.System.DatabaseSpec != nil && cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return r.reconcileSystemPostgreSQLLogic(cr)
	}

	// Defaults to MySQL
	return r.reconcileSystemMySQLLogic(cr)
}

func (r *APIManagerReconciler) reconcileSystemPostgreSQLLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	reconciler := operator.NewSystemPostgreSQLReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
	result, err := reconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	imageReconciler := operator.NewSystemPostgreSQLImageReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
	result, err = imageReconciler.Reconcile()
	return result, err
}

func (r *APIManagerReconciler) reconcileSystemMySQLLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	reconciler := operator.NewSystemMySQLReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
	result, err := reconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	imageReconciler := operator.NewSystemMySQLImageReconciler(operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr))
	result, err = imageReconciler.Reconcile()
	return result, err
}

func (r *APIManagerReconciler) reconcileAPIManagerStatus(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	updated, err := r.setDeploymentStatus(cr)
	if err != nil {
		return ctrl.Result{}, err
	}

	if updated {
		err = r.Client().Status().Update(context.TODO(), cr)
		if err != nil {
			// Ignore conflicts, resource might just be outdated.
			if errors.IsConflict(err) {
				r.Logger().Info("Failed to update status: resource might just be outdated")
				return ctrl.Result{Requeue: true}, nil
			}

			return ctrl.Result{}, fmt.Errorf("Failed to update API Manager deployment status: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *APIManagerReconciler) setDeploymentStatus(instance *appsv1alpha1.APIManager) (bool, error) {
	updated := false

	listOps := []client.ListOption{
		client.InNamespace(instance.Namespace),
	}
	dcList := &appsv1.DeploymentConfigList{}
	err := r.Client().List(context.TODO(), dcList, listOps...)
	if err != nil {
		return false, fmt.Errorf("Failed to list deployment configs: %w", err)
	}
	var dcs []appsv1.DeploymentConfig
	for _, dc := range dcList.Items {
		for _, ownerRef := range dc.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				dcs = append(dcs, dc)
				break
			}
		}
	}
	sort.Slice(dcs, func(i, j int) bool { return dcs[i].Name < dcs[j].Name })

	deploymentStatus := olm.GetDeploymentConfigStatus(dcs)
	if !reflect.DeepEqual(instance.Status.Deployments, deploymentStatus) {
		r.Logger().Info("Deployment status will be updated")
		instance.Status.Deployments = deploymentStatus
		updated = true
	}

	return updated, nil
}

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

	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

// APIManagerBackupReconciler reconciles a APIManagerBackup object
type APIManagerBackupReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that ReconcileAPIManagerBackup implements reconcile.Reconciler
var _ reconcile.Reconciler = &APIManagerBackupReconciler{}

// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagerbackups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagerbackups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagerbackups/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,namespace=placeholder,resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups=batch,namespace=placeholder,resources=jobs,verbs=get;list;watch;create;update;patch;delete

func (r *APIManagerBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger().WithValues("apimanagerbackup", req.NamespacedName)
	logger.Info("Reconciling APIManagerBackup")

	// Fetch the APIManagerBackup instance
	instance, err := r.getAPIManagerBackupCR(req)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Logger().Info("APIManagerBackup not found")
			return ctrl.Result{}, nil
		}
		r.Logger().Error(err, "Error getting APIManagerBackup")
		return ctrl.Result{}, err
	}

	res, err := r.setAPIManagerBackupDefaults(instance)
	if err != nil {
		logger.Error(err, "Error")
		return ctrl.Result{}, err
	}
	if res.Requeue {
		logger.Info("Defaults set for APIManagerBackup resource")
		return res, nil
	}

	// TODO prepare / implement something related to version annotations or upgrade?

	apiManagerBackupLogicReconciler, err := r.apiManagerBackupLogicReconciler(instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	res, err = apiManagerBackupLogicReconciler.Reconcile()
	if err != nil {
		logger.Error(err, "Error during reconciliation")
		return res, err
	}
	if res.Requeue {
		logger.Info("Reconciling not finished. Requeueing.")
		return res, nil
	}

	logger.Info("Reconciliation finished")
	return ctrl.Result{}, nil
}

func (r *APIManagerBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.APIManagerBackup{}).
		Complete(r)
}

func (r *APIManagerBackupReconciler) getAPIManagerBackupCR(request reconcile.Request) (*appsv1alpha1.APIManagerBackup, error) {
	instance := appsv1alpha1.APIManagerBackup{}
	err := r.Client().Get(context.TODO(), request.NamespacedName, &instance)
	return &instance, err
}

func (r *APIManagerBackupReconciler) setAPIManagerBackupDefaults(cr *appsv1alpha1.APIManagerBackup) (reconcile.Result, error) {
	changed, err := cr.SetDefaults() // TODO check where to put this
	if err != nil {
		return reconcile.Result{}, err
	}

	if changed {
		err = r.Client().Update(context.TODO(), cr)
	}

	return reconcile.Result{Requeue: changed}, err
}

func (r *APIManagerBackupReconciler) apiManagerBackupLogicReconciler(cr *appsv1alpha1.APIManagerBackup) (*APIManagerBackupLogicReconciler, error) {
	return NewAPIManagerBackupLogicReconciler(r.BaseReconciler, cr)
}

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

	ctrl "sigs.k8s.io/controller-runtime"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/restore"
	corev1 "k8s.io/api/core/v1"
)

// APIManagerRestoreReconciler reconciles a APIManagerRestore object
type APIManagerRestoreReconciler struct {
	*reconcilers.BaseReconciler
}

// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagerrestores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagerrestores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagerrestores/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,namespace=placeholder,resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups=batch,namespace=placeholder,resources=jobs,verbs=get;list;watch;create;update;patch;delete

func (r *APIManagerRestoreReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Logger().WithValues("apimanagerrestore", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *APIManagerRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.APIManagerRestore{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

func (r *APIManagerRestoreReconciler) getAPIManagerRestoreCR(request ctrl.Request) (*appsv1alpha1.APIManagerRestore, error) {
	instance := appsv1alpha1.APIManagerRestore{}
	err := r.Client().Get(context.TODO(), request.NamespacedName, &instance)
	return &instance, err
}

func (r *APIManagerRestoreReconciler) setAPIManagerRestoreDefaults(cr *appsv1alpha1.APIManagerRestore) (ctrl.Result, error) {
	changed, err := cr.SetDefaults() // TODO check where to put this
	if err != nil {
		return ctrl.Result{}, err
	}

	if changed {
		err = r.Client().Update(context.TODO(), cr)
	}

	return ctrl.Result{Requeue: changed}, err
}

func (r *APIManagerRestoreReconciler) apiManagerRestoreLogicReconciler(cr *appsv1alpha1.APIManagerRestore) (*APIManagerRestoreLogicReconciler, error) {
	apiManagerRestoreOptionsProvider := restore.NewAPIManagerRestoreOptionsProvider(cr, r.BaseReconciler.Client())
	options, err := apiManagerRestoreOptionsProvider.Options()
	if err != nil {
		return nil, err
	}

	apiManagerRestore := restore.NewAPIManagerRestore(options)
	return NewAPIManagerRestoreLogicReconciler(r.BaseReconciler, cr, apiManagerRestore), nil

}

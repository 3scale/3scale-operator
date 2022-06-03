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

package v1beta1

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1beta1 "github.com/3scale/3scale-operator/apis/apps/v1beta1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
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

	return ctrl.Result{}, nil
}

func (r *APIManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1beta1.APIManager{}).
		Complete(r)
}

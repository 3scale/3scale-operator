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
	"strings"

	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	"github.com/go-logr/logr"
	consolev1 "github.com/openshift/api/console/v1"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// WebConsoleReconciler reconciles a WebConsole object
type WebConsoleReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that ReconcileWebConsole implements reconcile.Reconciler
var _ reconcile.Reconciler = &WebConsoleReconciler{}

// +kubebuilder:rbac:groups=console.openshift.io,resources=consolelinks,verbs=get;create;update;delete

func (r *WebConsoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger().WithValues("webconsole", req.NamespacedName)

	logger.Info("Reconciling ReconcileWebConsole", "Operator version", version.Version)

	kindExists, err := r.HasConsoleLink()
	if err != nil {
		return ctrl.Result{}, err
	}
	if !kindExists {
		logger.Info("Console link not supported in the cluster")
		return ctrl.Result{}, nil
	}

	err = r.reconcileMasterLink(req, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	logger.V(1).Info("END")

	return ctrl.Result{}, nil
}

func (r *WebConsoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&routev1.Route{}).
		Complete(r)
}

func (r *WebConsoleReconciler) reconcileMasterLink(request reconcile.Request, logger logr.Logger) error {
	if !strings.Contains(request.Name, "zync-3scale-master") {
		// Nothing to do
		return nil
	}

	route := &routev1.Route{}
	err := r.Client().Get(r.Context(), request.NamespacedName, route)
	if err != nil && !errors.IsNotFound(err) {
		// Error reading the object - requeue the request.
		return err
	}

	if errors.IsNotFound(err) {
		logger.V(1).Info("Master route not found", "name", request.Name)
		// cluster-scoped resource must not have a namespace-scoped owner
		// So consolelinks cannot have owners like apimanager or route object
		// delete consolelink if exists
		desired := &consolev1.ConsoleLink{
			ObjectMeta: metav1.ObjectMeta{
				Name: helper.GetMasterConsoleLinkName(request.Namespace),
			},
		}
		helper.TagObjectToDelete(desired)
		err := r.ReconcileResource(&consolev1.ConsoleLink{}, desired, reconcilers.CreateOnlyMutator)
		return err
	}

	logger.V(1).Info("Master route found", "name", request.Name)

	err = r.ReconcileResource(&consolev1.ConsoleLink{}, helper.GetMasterConsoleLink(route), helper.GenericConsoleLinkMutator)
	logger.V(1).Info("Reconcile master consolelink", "err", err)
	return err
}

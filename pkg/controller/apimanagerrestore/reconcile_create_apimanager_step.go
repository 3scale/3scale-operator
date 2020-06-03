package apimanagerrestore

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcileCreateAPIManagerStep struct {
	APIManagerRestoreBaseStep
}

func (r *ReconcileCreateAPIManagerStep) Execute() (reconcile.Result, error) {
	apimanager, err := r.apiManagerFromSharedBackupSecret()
	if err != nil {
		return reconcile.Result{}, err
	}

	existing := &appsv1alpha1.APIManager{}
	err = r.ReconcileResource(existing, apimanager, reconcilers.CreateOnlyMutator)
	return reconcile.Result{}, err
}

func (r *ReconcileCreateAPIManagerStep) Completed() (bool, error) {
	apiManagerToRestoreName := r.cr.Status.APIManagerToRestoreRef.Name

	err := r.GetResource(types.NamespacedName{Name: apiManagerToRestoreName, Namespace: r.cr.Namespace}, &appsv1alpha1.APIManager{})
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}
	if err != nil {
		if errors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}
	return true, nil
}

func (r *ReconcileCreateAPIManagerStep) Identifier() string {
	return "ReconcileRestoreAPIManagerStep"
}

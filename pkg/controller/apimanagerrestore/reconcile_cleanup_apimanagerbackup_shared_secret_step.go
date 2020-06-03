package apimanagerrestore

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcileCleanupAPIManagerBackupSharedSecretStep struct {
	APIManagerRestoreBaseStep
}

func (r *ReconcileCleanupAPIManagerBackupSharedSecretStep) Execute() (reconcile.Result, error) {
	desiredSecret, err := r.sharedBackupSecret()
	if err != nil {
		return reconcile.Result{}, err
	}
	if desiredSecret == nil {
		// TODO is this correct? Should the code in the execute step assume that always
		// that this method is executed the shared secrets should exist beforehand?
		return reconcile.Result{}, fmt.Errorf("Shared secret should exist at this point")
	}

	common.TagObjectToDelete(desiredSecret)
	err = r.ReconcileResource(&v1.Secret{}, desiredSecret, reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileCleanupAPIManagerBackupSharedSecretStep) Completed() (bool, error) {
	existingSecret, err := r.sharedBackupSecret()
	if err != nil {
		return false, err
	}
	if existingSecret != nil {
		return false, nil
	}
	return true, nil
}

func (r *ReconcileCleanupAPIManagerBackupSharedSecretStep) Identifier() string {
	return "ReconcileCleanupAPIManagerBackupSharedSecretStep"
}

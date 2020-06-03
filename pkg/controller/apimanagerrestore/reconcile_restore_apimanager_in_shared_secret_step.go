package apimanagerrestore

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcileAPIManagerInSharedSecretStep struct {
	APIManagerRestoreBaseStep
}

func (r *ReconcileAPIManagerInSharedSecretStep) Execute() (reconcile.Result, error) {
	desired := r.apiManagerRestore.CreateAPIManagerSharedSecretJob()
	if desired == nil {
		return reconcile.Result{}, fmt.Errorf("Unknown error executing step '%s'", r.Identifier())
	}

	// TODO is this correct? It seems in the execute method itself we perform some
	// part of reconciliation
	res, err := r.reconcileJob(desired)
	if res.Requeue || err != nil {
		return res, nil
	}

	secret, err := r.sharedBackupSecret()
	if err != nil {
		return reconcile.Result{}, err
	}
	if secret == nil {
		r.Logger().Info("Shared secret '%s' not found. Waiting...", r.apiManagerRestore.SecretToShareName())
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}
	apimanager, err := r.apiManagerFromSharedBackupSecret()

	if err != nil {
		return reconcile.Result{}, err
	}
	r.cr.Status.APIManagerToRestoreRef = &v1.LocalObjectReference{
		Name: apimanager.Name,
	}
	err = r.UpdateResourceStatus(r.cr)
	return reconcile.Result{}, err
}

func (r *ReconcileAPIManagerInSharedSecretStep) Completed() (bool, error) {
	return r.cr.Status.APIManagerToRestoreRef != nil, nil
}

func (r *ReconcileAPIManagerInSharedSecretStep) Identifier() string {
	return "ReconcileAPIManagerInSharedSecretStep"
}

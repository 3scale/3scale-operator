package apimanager

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Upgrade26_to_27 struct {
	BaseUpgrade
}

func (u *Upgrade26_to_27) Upgrade() (reconcile.Result, error) {
	res, err := u.upgradeAMPImageStreams()
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *Upgrade26_to_27) upgradeAMPImageStreams() (reconcile.Result, error) {
	// implement upgrade procedure by reconcile procedure
	baseReconciler := operator.NewBaseReconciler(u.client, u.apiClientReader, u.scheme, u.logger)
	baseLogicReconciler := operator.NewBaseLogicReconciler(baseReconciler)
	reconciler := operator.NewAMPImagesReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.cr))
	return reconciler.Reconcile()
}

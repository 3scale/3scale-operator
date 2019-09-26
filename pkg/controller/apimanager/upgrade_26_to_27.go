package apimanager

import (
	"fmt"

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

	if !u.highAvailabilityModeEnabled() {
		res, err = u.upgradeBackendRedisImageStream()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.upgradeSystemRedisImageStream()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.upgradeSystemDatabaseImageStream()
		if res.Requeue || err != nil {
			return res, err
		}
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

func (u *Upgrade26_to_27) highAvailabilityModeEnabled() bool {
	return u.cr.Spec.HighAvailability != nil && u.cr.Spec.HighAvailability.Enabled
}

func (u *Upgrade26_to_27) upgradeBackendRedisImageStream() (reconcile.Result, error) {
	redis, err := operator.Redis(u.cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	baseReconciler := operator.NewBaseReconciler(u.client, u.apiClientReader, u.scheme, u.logger)
	baseLogicReconciler := operator.NewBaseLogicReconciler(baseReconciler)
	reconciler := operator.NewImageStreamBaseReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.cr), operator.NewImageStreamGenericReconciler())
	return reconcile.Result{}, reconciler.Reconcile(redis.BackendImageStream())
}

func (u *Upgrade26_to_27) upgradeSystemRedisImageStream() (reconcile.Result, error) {
	redis, err := operator.Redis(u.cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	baseReconciler := operator.NewBaseReconciler(u.client, u.apiClientReader, u.scheme, u.logger)
	baseLogicReconciler := operator.NewBaseLogicReconciler(baseReconciler)
	reconciler := operator.NewImageStreamBaseReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.cr), operator.NewImageStreamGenericReconciler())
	return reconcile.Result{}, reconciler.Reconcile(redis.SystemImageStream())
}

func (u *Upgrade26_to_27) upgradeSystemDatabaseImageStream() (reconcile.Result, error) {
	if u.cr.Spec.System.DatabaseSpec.MySQL != nil {
		return u.upgradeSystemMySQLImageStream()
	}

	if u.cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return u.upgradeSystemPostgreSQLImageStream()
	}

	return reconcile.Result{}, fmt.Errorf("System database is not set")
}

func (u *Upgrade26_to_27) upgradeSystemMySQLImageStream() (reconcile.Result, error) {
	baseReconciler := operator.NewBaseReconciler(u.client, u.apiClientReader, u.scheme, u.logger)
	baseLogicReconciler := operator.NewBaseLogicReconciler(baseReconciler)
	reconciler := operator.NewSystemMySQLImageReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.cr))
	return reconciler.Reconcile()
}

func (u *Upgrade26_to_27) upgradeSystemPostgreSQLImageStream() (reconcile.Result, error) {
	baseReconciler := operator.NewBaseReconciler(u.client, u.apiClientReader, u.scheme, u.logger)
	baseLogicReconciler := operator.NewBaseLogicReconciler(baseReconciler)
	reconciler := operator.NewSystemPostgreSQLImageReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.cr))
	return reconciler.Reconcile()
}

package helper

import (
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ReconcileErrorHandler reads the error type and based on that, it responsds with either a requeue true or false
func ReconcileErrorHandler(err error, reqLogger logr.Logger) ctrl.Result {
	// On validation error, do not retry since there's something wrong with the spec - do not re-trigger the reconciler
	if IsInvalidSpecError(err) {
		reqLogger.Info("ERROR", "spec validation error", err)
		return ctrl.Result{}
	}

	// On wait error - requeue the reconciler
	if IsWaitError(err) {
		reqLogger.Info("ERROR", "wait error", err)
		return ctrl.Result{Requeue: true}
	}

	// On orphan error - do not re-trigger the reconciler
	if IsOrphanSpecError(err) {
		reqLogger.Info("ERROR", "orphan error", err)
		return ctrl.Result{}
	}

	reqLogger.Info("ERROR", "generic k8s error", err)
	return ctrl.Result{Requeue: true}
}

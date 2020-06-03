package apimanagerrestore

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcileRestoreCompletionStep struct {
	APIManagerRestoreBaseStep
}

func (r *ReconcileRestoreCompletionStep) Execute() (reconcile.Result, error) {
	completionTimeUTC := metav1.Time{Time: clock.Now().UTC()}
	restoreFinished := true
	r.cr.Status.Completed = &restoreFinished
	r.cr.Status.CompletionTime = &completionTimeUTC
	err := r.UpdateResourceStatus(r.cr)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileRestoreCompletionStep) Completed() (bool, error) {
	return r.cr.RestoreCompleted(), nil
}

func (r *ReconcileRestoreCompletionStep) Identifier() string {
	return "ReconcileRestoreCompletionStep"
}

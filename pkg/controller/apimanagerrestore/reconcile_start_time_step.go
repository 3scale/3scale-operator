package apimanagerrestore

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcileStartTimeStep struct {
	APIManagerRestoreBaseStep
}

func (r *ReconcileStartTimeStep) Execute() (reconcile.Result, error) {
	startTimeUTC := metav1.Time{Time: clock.Now().UTC()}
	r.cr.Status.StartTime = &startTimeUTC
	err := r.UpdateResourceStatus(r.cr)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileStartTimeStep) Completed() (bool, error) {
	return r.cr.Status.StartTime != nil, nil
}

func (r *ReconcileStartTimeStep) Identifier() string {
	return "ReconcileStartTimeStep"
}

package apimanagerrestore

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/common"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcileRestoreSystemFileStoragePVCFromPVCStep struct {
	APIManagerRestoreBaseStep
}

func (r *ReconcileRestoreSystemFileStoragePVCFromPVCStep) Execute() (reconcile.Result, error) {
	desired := r.apiManagerRestore.RestoreSystemFileStoragePVCFromPVCJob()
	if desired == nil {
		return reconcile.Result{}, fmt.Errorf("Unknown error executing step '%s'", r.Identifier())
	}

	return r.reconcileJob(desired)
}

func (r *ReconcileRestoreSystemFileStoragePVCFromPVCStep) Completed() (bool, error) {
	desired := r.apiManagerRestore.RestoreSystemFileStoragePVCFromPVCJob()
	if desired == nil {
		return false, fmt.Errorf("Unknown error executing step '%s'", r.Identifier())
	}

	existing := &batchv1.Job{}
	err := r.GetResource(common.ObjectKey(desired), existing)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	// TODO is this implementation good enough?
	// Should we check whether both conditions are set?
	// Does the order matter?
	// Should we just check succeeded vs completions instead of checking
	// the conditions?
	for _, cond := range existing.Status.Conditions {
		if cond.Type == batchv1.JobComplete && cond.Status == v1.ConditionTrue {
			return true, nil
		}
		if cond.Type == batchv1.JobFailed && cond.Status == v1.ConditionTrue {
			return false, fmt.Errorf("Job '%s' failed to complete", existing.Name)
		}
	}

	r.Logger().Info("Job has still not finished", "Job Name", desired.Name, "Actively running Pods", existing.Status.Active, "Failed pods", existing.Status.Failed)
	return false, nil
}

func (r *ReconcileRestoreSystemFileStoragePVCFromPVCStep) Identifier() string {
	return "ReconcileRestoreSystemFileStoragePVCFromPVCStep"
}

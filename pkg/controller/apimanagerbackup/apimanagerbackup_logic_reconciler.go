package apimanagerbackup

import (
	"time"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/backup"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeclock "k8s.io/apimachinery/pkg/util/clock"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var clock kubeclock.Clock = &kubeclock.RealClock{}

type APIManagerBackupLogicReconciler struct {
	*reconcilers.BaseReconciler
	logger           logr.Logger
	apiManagerBackup *backup.APIManagerBackup
	cr               *appsv1alpha1.APIManagerBackup // TODO we use the cr to access and update status fields. Is there an alternative to not depend on status fields?
}

func NewAPIManagerBackupLogicReconciler(b *reconcilers.BaseReconciler, cr *appsv1alpha1.APIManagerBackup) (*APIManagerBackupLogicReconciler, error) {
	res := &APIManagerBackupLogicReconciler{
		BaseReconciler: b,
		logger:         b.Logger().WithValues("APIManagerBackup Controller", cr.Name),
		cr:             cr,
	}

	if cr.BackupCompleted() {
		return res, nil
	}

	// We only set the apiManagerBackup field when
	// the backup has not completed. The reason for this is that
	// The creation of the APIManagerBackup fills an option which is
	// the APIManager which requires the APIManager to exist.
	// The downside of this approach is that we must make sure at the beginning
	// of the Reconcile funtion that we don't do anything else if the backup
	// has completed. Otherwise we would potentially get a nil pointer
	// exception
	// TODO is there an alternative or better way to do this?
	apiManagerBackupOptionsProvider := backup.NewAPIManagerBackupOptionsProvider(cr, b.Client())
	options, err := apiManagerBackupOptionsProvider.Options()
	if err != nil {
		return nil, err
	}
	apiManagerBackup := backup.NewAPIManagerBackup(options)
	res.apiManagerBackup = apiManagerBackup

	return res, nil
}

func (r *APIManagerBackupLogicReconciler) Logger() logr.Logger {
	return r.logger
}

func (r *APIManagerBackupLogicReconciler) Reconcile() (reconcile.Result, error) {
	if r.cr.BackupCompleted() {
		r.Logger().Info("Backup completed. End of reconciliation")
		return reconcile.Result{}, nil
	}

	result, err := r.reconcileAPIManagerSourceStatusField()
	if result.Requeue || err != nil {
		return result, err
	}

	result, err = r.reconcileStartTimeField()
	if result.Requeue || err != nil {
		return result, err
	}

	result, err = r.reconcileBackupInS3Destination()
	if result.Requeue || err != nil {
		return result, err
	}

	result, err = r.reconcileBackupInPVCDestination()
	if result.Requeue || err != nil {
		return result, err
	}

	result, err = r.reconcileBackupCompletion()
	if result.Requeue || err != nil {
		return result, err
	}

	return result, err
}

func (r *APIManagerBackupLogicReconciler) reconcileBackupInS3Destination() (reconcile.Result, error) {
	// TODO implement
	return reconcile.Result{}, nil
}

func (r *APIManagerBackupLogicReconciler) reconcileBackupInPVCDestination() (reconcile.Result, error) {
	var res reconcile.Result
	var err error

	err = r.reconcileBackupDestinationPVC()
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err = r.reconcileBackupDestinationPVCStatus()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = r.reconcileBackupSecretsAndConfigMapsToPVCJob()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = r.reconcileAPIManagerCustomResourceBackupToPVCJob()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = r.reconcileBackupSystemFileStoragePVCToPVCJob()
	if res.Requeue || err != nil {
		return res, err
	}

	return res, err
}

func (r *APIManagerBackupLogicReconciler) reconcileBackupDestinationPVC() error {
	desired := r.apiManagerBackup.BackupDestinationPVC()
	if desired == nil {
		return nil
	}

	// TODO create mutator function for PVC ?
	err := r.ReconcileResource(&v1.PersistentVolumeClaim{}, desired, reconcilers.CreateOnlyMutator)

	return err
}

func (r *APIManagerBackupLogicReconciler) setOwnerReference(obj common.KubernetesObject) error {
	err := controllerutil.SetControllerReference(r.cr, obj, r.BaseReconciler.Scheme())
	if err != nil {
		r.Logger().Error(err, "Error setting OwnerReference on object",
			"Kind", obj.GetObjectKind().GroupVersionKind().String(),
			"Namespace", obj.GetNamespace(),
			"Name", obj.GetName(),
		)
	}
	return err
}

func (r *APIManagerBackupLogicReconciler) reconcileJob(desired *batchv1.Job) (reconcile.Result, error) {
	if err := r.setOwnerReference(desired); err != nil {
		return reconcile.Result{}, err
	}

	existing := &batchv1.Job{}
	err := r.GetResource(types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	if errors.IsNotFound(err) {
		err := r.CreateResource(desired)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	// TODO We do not reconcile ownerReference or labels nor annotations
	// should we do it? Jobs are one-shot so there's not much point on
	// making updates to them

	if existing.Status.Succeeded != *desired.Spec.Completions {
		r.Logger().Info("Job has still not finished", "Job Name", desired.Name, "Actively running Pods", existing.Status.Active, "Failed pods", existing.Status.Failed)
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	r.Logger().Info("Job finished successfully", "Job Name", desired.Name)
	return reconcile.Result{}, nil
}

func (r *APIManagerBackupLogicReconciler) reconcileBackupSecretsAndConfigMapsToPVCJob() (reconcile.Result, error) {
	desired := r.apiManagerBackup.BackupSecretsAndConfigMapsToPVCJob()
	if desired == nil {
		return reconcile.Result{}, nil
	}

	return r.reconcileJob(desired)
}

func (r *APIManagerBackupLogicReconciler) reconcileAPIManagerCustomResourceBackupToPVCJob() (reconcile.Result, error) {
	desired := r.apiManagerBackup.BackupAPIManagerCustomResourceToPVCJob()
	if desired == nil {
		return reconcile.Result{}, nil
	}

	return r.reconcileJob(desired)
}

func (r *APIManagerBackupLogicReconciler) reconcileBackupSystemFileStoragePVCToPVCJob() (reconcile.Result, error) {
	desired := r.apiManagerBackup.BackupSystemFileStoragePVCToPVCJob()
	if desired == nil {
		return reconcile.Result{}, nil
	}

	return r.reconcileJob(desired)
}

func (r *APIManagerBackupLogicReconciler) reconcileBackupCompletion() (reconcile.Result, error) {
	if !r.cr.BackupCompleted() {
		// TODO make this more robust only setting it in case all substeps have been completed?
		// It might be a little bit redundant because the steps are checked during the reconciliation
		backupFinished := true
		completionTimeUTC := metav1.Time{Time: clock.Now().UTC()}
		r.cr.Status.Completed = &backupFinished
		r.cr.Status.CompletionTime = &completionTimeUTC
		err := r.UpdateResourceStatus(r.cr)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *APIManagerBackupLogicReconciler) reconcileAPIManagerSourceStatusField() (reconcile.Result, error) {
	apiManager := r.apiManagerBackup.APIManager()

	if r.cr.Status.APIManagerSourceName == nil {
		r.cr.Status.APIManagerSourceName = &apiManager.Name
		err := r.UpdateResourceStatus(r.cr)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.Logger().Info("APIManager source name set in status. Requeuing", "APIManager source name", r.cr.Status.APIManagerSourceName)
		return reconcile.Result{Requeue: true}, err
	}
	if *r.cr.Status.APIManagerSourceName != apiManager.Name { // TODO should we reconcile this case?
		r.cr.Status.APIManagerSourceName = &apiManager.Name
		err := r.UpdateResourceStatus(r.cr)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.Logger().Info("APIManager source name changed in status. Requeuing", "APIManager source name", r.cr.Status.APIManagerSourceName)
		return reconcile.Result{Requeue: true}, err
	}
	return reconcile.Result{}, nil
}

func (r *APIManagerBackupLogicReconciler) reconcileStartTimeField() (reconcile.Result, error) {
	if r.cr.Status.StartTime == nil {
		startTimeUTC := metav1.Time{Time: clock.Now().UTC()}
		r.cr.Status.StartTime = &startTimeUTC
		err := r.UpdateResourceStatus(r.cr)
		return reconcile.Result{Requeue: true}, err
	}
	return reconcile.Result{}, nil
}

func (r *APIManagerBackupLogicReconciler) reconcileBackupDestinationPVCStatus() (reconcile.Result, error) {
	if r.cr.Spec.BackupSource.PersistentVolumeClaim == nil {
		return reconcile.Result{}, nil
	}

	if r.cr.Status.BackupPersistentVolumeClaimName == nil {
		r.cr.Status.BackupPersistentVolumeClaimName = &r.apiManagerBackup.BackupDestinationPVC().Name
		err := r.UpdateResourceStatus(r.cr)
		return reconcile.Result{Requeue: true}, err
	}
	return reconcile.Result{}, nil
}

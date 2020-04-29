package apimanagerrestore

import (
	"fmt"
	"time"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/backup"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/restore"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	kubeclock "k8s.io/apimachinery/pkg/util/clock"
)

var clock kubeclock.Clock = &kubeclock.RealClock{}

type APIManagerRestoreLogicReconciler struct {
	*reconcilers.BaseReconciler
	logger            logr.Logger
	cr                *appsv1alpha1.APIManagerRestore // TODO we use the cr to access and update status fields. Is there an alternative to not depend on status fields?
	apiManagerRestore *restore.APIManagerRestore
}

func NewAPIManagerRestoreLogicReconciler(b *reconcilers.BaseReconciler, cr *appsv1alpha1.APIManagerRestore, apiManagerRestore *restore.APIManagerRestore) *APIManagerRestoreLogicReconciler {
	return &APIManagerRestoreLogicReconciler{
		BaseReconciler:    b,
		logger:            b.Logger().WithValues("APIManagerRestore Controller", cr.Name),
		cr:                cr,
		apiManagerRestore: apiManagerRestore,
	}
}

func (r *APIManagerRestoreLogicReconciler) Logger() logr.Logger {
	return r.logger
}

func (r *APIManagerRestoreLogicReconciler) Reconcile() (reconcile.Result, error) {
	if r.cr.RestoreCompleted() {
		r.Logger().Info("Restore completed. End of reconciliation")
		return reconcile.Result{}, nil
	}

	result, err := r.reconcileRestoreFromS3Source()
	if result.Requeue || err != nil {
		return result, err
	}

	result, err = r.reconcileRestoreFromPVCSource()
	if result.Requeue || err != nil {
		return result, err
	}

	result, err = r.reconcileRestoreCompletion()
	if result.Requeue || err != nil {
		return result, err
	}

	return result, err
}

func (r *APIManagerRestoreLogicReconciler) reconcileRestoreFromS3Source() (reconcile.Result, error) {
	// TODO implement
	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileRestoreFromPVCSource() (reconcile.Result, error) {
	var res reconcile.Result
	var err error

	res, err = r.reconcileRestoreSecretsAndConfigMapsFromPVCJob()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = r.reconcileRestoreAPIManagerInSharedSecret()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = r.reconcileRestoreSystemFileStoragePVCFromPVCJob()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = r.reconcileRestoreAPIManager()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = r.reconcileAPIManagerBackupSharedInSecretCleanup()
	if res.Requeue || err != nil {
		return res, err
	}

	return res, err
}

func (r *APIManagerRestoreLogicReconciler) reconcileRestoreSecretsAndConfigMapsFromPVCJob() (reconcile.Result, error) {
	desired := r.apiManagerRestore.RestoreSecretsAndConfigMapsFromPVCJob()
	if desired == nil {
		return reconcile.Result{}, nil
	}

	// TODO create Mutator function for Jobs??
	existing := &batchv1.Job{}
	// Check if this step has finished
	if r.cr.SecretsAndConfigMapsRestoreStepFinished() {
		return reconcile.Result{}, nil
	}

	// Check if backup substep has completed
	if !r.cr.SecretsAndConfigMapsRestoreSubStepFinished() {
		err := r.ReconcileResource(existing, desired, reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
		substepFinished := true
		if existing.Status.Succeeded != *desired.Spec.Completions {
			r.Logger().Info("Job has still not finished", "Job Name", desired.Name, "Actively running Pods", existing.Status.Active, "Failed pods", existing.Status.Failed)
			return reconcile.Result{Requeue: true}, nil
		}

		r.cr.Status.SecretsAndConfigMapsRestoreSubStepFinished = &substepFinished
		err = r.UpdateResourceStatus(r.cr)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.Logger().Info("Job finished successfully. Requeing", "Job Name", desired.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	// Check if cleanup substep has completed
	if !r.cr.SecretsAndConfigMapsCleanupSubStepFinished() {
		common.TagToObjectDeleteWithPropagationPolicy(desired, metav1.DeletePropagationForeground)
		err := r.ReconcileResource(existing, desired, reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}

		// TODO is this needed or it's somehow handled correctly in ReconcileResource?
		// The behavior we want is that we only update the state of the CR when we are
		// sure the object does not exist anymore. The reason of that is that
		// if we do not do it this way it could happen that we mark the state of this
		// substep as completed but the Job is not completely deleted and we move to the
		// next step, which would fail because Job is not deleted nor the PVC associated
		// to it, which would provoke an error when trying to reuse it for the next job PVC
		err = r.GetResource(common.ObjectKey(desired), existing)
		if err != nil && !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		if err == nil {
			r.Logger().Info("Job still not completely deleted. Requeuing", "Job Name", desired.Name)
			return reconcile.Result{Requeue: true}, nil
		}

		substepFinished := true
		r.cr.Status.SecretsAndConfigMapsCleanupSubStepFinished = &substepFinished
		err = r.UpdateResourceStatus(r.cr)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.Logger().Info("Job cleant up successfully. Requeuing", "Job Name", desired.Name)
	}

	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileSystemStoragePVC() (reconcile.Result, error) {
	// TODO is it enough with just calling ReconcileResource???
	exists, err := r.systemStoragePVCExists()
	if err != nil {
		return reconcile.Result{}, err
	}
	if !exists {
		apimanager, apiManagerErr := r.apiManagerFromSharedBackupSecret()
		if apiManagerErr != nil {
			return reconcile.Result{}, apiManagerErr
		}
		restoreInfo, restoreInfoErr := r.runtimeRestoreInfoFromAPIManager(apimanager)
		if restoreInfoErr != nil {
			return reconcile.Result{}, restoreInfoErr
		}
		err := r.ReconcileResource(&v1.PersistentVolumeClaim{}, r.apiManagerRestore.SystemStoragePVC(restoreInfo), reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) runtimeRestoreInfoFromAPIManager(apimanager *appsv1alpha1.APIManager) (*restore.RuntimeAPIManagerRestoreInfo, error) {
	var storageClass *string
	if apimanager.Spec.System != nil && apimanager.Spec.System.FileStorageSpec != nil && apimanager.Spec.System.FileStorageSpec.PVC != nil {
		storageClass = apimanager.Spec.System.FileStorageSpec.PVC.StorageClassName
	}
	restoreInfo := &restore.RuntimeAPIManagerRestoreInfo{
		PVCStorageClass: storageClass,
	}
	return restoreInfo, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileRestoreSystemFileStoragePVCFromPVCJob() (reconcile.Result, error) {
	desired := r.apiManagerRestore.RestoreSystemFileStoragePVCFromPVCJob()
	if desired == nil {
		return reconcile.Result{}, nil
	}

	// TODO create Mutator function for Jobs??
	existing := &batchv1.Job{}
	if r.cr.SystemFileStorageRestoreStepFinished() {
		return reconcile.Result{}, nil
	}

	// Check if backup substep has completed
	if !r.cr.SystemFileStorageRestoreSubStepFinished() {
		res, err := r.reconcileSystemStoragePVC()
		if res.Requeue || err != nil {
			return res, err
		}

		err = r.ReconcileResource(existing, desired, reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
		backupFinished := true
		if existing.Status.Succeeded != *desired.Spec.Completions {
			r.Logger().Info("Job has still not finished", "Job Name", desired.Name, "Actively running Pods", existing.Status.Active, "Failed pods", existing.Status.Failed)
			return reconcile.Result{Requeue: true}, nil
		}

		r.cr.Status.SystemFileStorageRestoreSubStepFinished = &backupFinished
		err = r.UpdateResourceStatus(r.cr)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.Logger().Info("Job finished successfully. Requeing", "Job Name", desired.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	// Check if cleanup substep has completed
	if !r.cr.SystemFileStorageCleanupSubStepFinished() {
		common.TagToObjectDeleteWithPropagationPolicy(desired, metav1.DeletePropagationForeground)
		// TODO delete performed with ReconcileResource does not have the option to delete
		// dependants which means that if the Job is deleted the Pod is left there
		// Think of a way to do this. Maybe instead of using ReconcileResource perform
		// the manual deletion.
		// Apparently leaving the pod in terminated is not a problem regarding the volume
		// It seems that the volume can be specified in several Pods if only one of them
		// is really using (this is, the others are terminated)
		// In any case we should fix it. I've observed for example that you
		// cannot delete the pvc manually if it's being used by some Pod, even if
		// it is terminated
		err := r.ReconcileResource(existing, desired, reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = r.GetResource(common.ObjectKey(desired), existing)
		if err != nil && !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		if err == nil {
			r.Logger().Info("Job still not completely deleted. Requeuing", "Job Name", desired.Name)
			return reconcile.Result{Requeue: true}, nil
		}

		backupFinished := true
		r.cr.Status.SystemFileStorageCleanupSubStepFinished = &backupFinished
		err = r.UpdateResourceStatus(r.cr)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.Logger().Info("Job cleant up successfully. Requeuing", "Job Name", desired.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) systemStoragePVCExists() (bool, error) {
	pvc := &v1.PersistentVolumeClaim{}
	err := r.GetResource(types.NamespacedName{Name: component.SystemFileStoragePVCName, Namespace: r.cr.Namespace}, pvc)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileRestoreCompletion() (reconcile.Result, error) {
	if !r.cr.RestoreCompleted() {
		// TODO make this more robust only setting it in case all substeps have been completed?
		// It might be a little bit redundant because the steps are checked during the reconciliation
		backupFinished := true
		r.cr.Status.Completed = &backupFinished
		err := r.UpdateResourceStatus(r.cr)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileRestoreAPIManagerInSharedSecret() (reconcile.Result, error) {
	desired := r.apiManagerRestore.CreateAPIManagerSharedSecretJob()
	if desired == nil {
		return reconcile.Result{}, nil
	}

	existing := &batchv1.Job{}
	if r.cr.APIManagerBackupSharedInSecret() {
		return reconcile.Result{}, nil
	}

	err := r.ReconcileResource(existing, desired, reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}
	if existing.Status.Succeeded != *desired.Spec.Completions {
		r.Logger().Info("Job has still not finished", "Job Name", desired.Name, "Actively running Pods", existing.Status.Active, "Failed pods", existing.Status.Failed)
		return reconcile.Result{Requeue: true}, nil
	}

	secret, err := r.sharedBackupSecret()
	if err != nil {
		return reconcile.Result{}, err
	}
	if secret == nil {
		r.Logger().Info("shared APIManager backup secret still does not exist", "secret", r.apiManagerRestore.SecretToShareName())
		return reconcile.Result{Requeue: true}, nil
	}

	stepFinished := true
	r.cr.Status.APIManagerBackupSharedInSecret = &stepFinished
	err = r.UpdateResourceStatus(r.cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileRestoreAPIManagerFromPVCJob() (reconcile.Result, error) {
	desired := r.apiManagerRestore.RestoreAPIManagerFromPVCJob()
	if desired == nil {
		return reconcile.Result{}, nil
	}

	// TODO create Mutator function for Jobs??
	existing := &batchv1.Job{}
	// Check if this step has finished
	if r.cr.APIManagerRestoreStepFinished() {
		return reconcile.Result{}, nil
	}

	// Check if backup substep has completed
	if !r.cr.APIManagerRestoreStepFinished() {
		err := r.ReconcileResource(existing, desired, reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
		substepFinished := true
		if existing.Status.Succeeded != *desired.Spec.Completions {
			r.Logger().Info("Job has still not finished", "Job Name", desired.Name, "Actively running Pods", existing.Status.Active, "Failed pods", existing.Status.Failed)
			return reconcile.Result{Requeue: true}, nil
		}

		r.cr.Status.APIManagerRestoreStepFinished = &substepFinished
		err = r.UpdateResourceStatus(r.cr)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.Logger().Info("Job finished successfully. Requeing", "Job Name", desired.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) sharedBackupSecret() (*v1.Secret, error) {
	secret := &v1.Secret{}
	err := r.GetResource(types.NamespacedName{Name: r.apiManagerRestore.SecretToShareName(), Namespace: r.cr.Namespace}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	_, ok := secret.Data[backup.APIManagerSerializedBackupFileName]
	if !ok {
		return nil, fmt.Errorf("Expected key '%s' in secret '%s' not found", r.apiManagerRestore.SecretToShareName(), backup.APIManagerSerializedBackupFileName)
	}

	return secret, nil
}

func (r *APIManagerRestoreLogicReconciler) apiManagerFromSharedBackupSecret() (*appsv1alpha1.APIManager, error) {
	secret, err := r.sharedBackupSecret()
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, fmt.Errorf("Secret '%s' not found", r.apiManagerRestore.SecretToShareName())
	}

	// https://godoc.org/k8s.io/apimachinery/pkg/runtime/serializer#CodecFactory
	codecFactory := k8sserializer.NewCodecFactory(r.Scheme())
	// https://godoc.org/k8s.io/apimachinery/pkg/runtime#Dec
	deserializer := codecFactory.UniversalDeserializer()

	serializedAPIManager := secret.Data[backup.APIManagerSerializedBackupFileName]

	apimanagerRuntimeObj, _, err := deserializer.Decode(serializedAPIManager, nil, &appsv1alpha1.APIManager{})
	apimanager, ok := apimanagerRuntimeObj.(*appsv1alpha1.APIManager)
	if !ok {
		return nil, fmt.Errorf("%T is not a *appsv1alpha1.APIManager", apimanagerRuntimeObj)
	}

	apimanager.Namespace = r.cr.Namespace

	return apimanager, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileRestoreAPIManager() (reconcile.Result, error) {
	apimanager, err := r.apiManagerFromSharedBackupSecret()
	if err != nil {
		return reconcile.Result{}, err
	}

	existing := &appsv1alpha1.APIManager{}
	err = r.ReconcileResource(existing, apimanager, reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	stepFinished := true
	r.cr.Status.APIManagerRestoreStepFinished = &stepFinished
	err = r.UpdateResourceStatus(r.cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileAPIManagerBackupSharedInSecretCleanup() (reconcile.Result, error) {
	if r.cr.APIManagerBackupSharedInSecretCleanupFinished() {
		return reconcile.Result{}, nil
	}
	desired := r.apiManagerRestore.CreateAPIManagerSharedSecretJob()
	existing := &batchv1.Job{}
	common.TagToObjectDeleteWithPropagationPolicy(desired, metav1.DeletePropagationForeground)

	err := r.ReconcileResource(&batchv1.Job{}, desired, reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.GetResource(common.ObjectKey(desired), existing)
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}
	if err == nil {
		r.Logger().Info("Job still not completely deleted. Requeuing", "Job Name", desired.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	desiredSecret, err := r.sharedBackupSecret()
	existingSecret := &v1.Secret{}
	if err != nil {
		return reconcile.Result{}, err
	}
	if desiredSecret != nil {
		common.TagObjectToDelete(desiredSecret)
		err = r.ReconcileResource(&v1.Secret{}, desiredSecret, reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = r.GetResource(common.ObjectKey(desiredSecret), existingSecret)
		if err != nil && !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		if err == nil {
			r.Logger().Info("Secret still not completely deleted. Requeuing", "Secret Name", desiredSecret.Name)
			return reconcile.Result{Requeue: true}, nil
		}
	}

	stepFinished := true
	r.cr.Status.APIManagerBackupSharedInSecretCleanup = &stepFinished
	err = r.UpdateResourceStatus(r.cr)
	if err != nil {
		return reconcile.Result{}, err
	}
	r.Logger().Info("Job and cleant up successfully. Requeuing", "Job Name", desired.Name)

	return reconcile.Result{}, nil
}

package apimanagerrestore

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/backup"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/restore"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	res, err = r.reconcileWaitForAPIManagerReady()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = r.reconcileResynchronizeZyncDomains()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = r.reconcileAPIManagerBackupSharedInSecretCleanup()
	if res.Requeue || err != nil {
		return res, err
	}

	return res, err
}

func (r *APIManagerRestoreLogicReconciler) setOwnerReference(obj common.KubernetesObject) error {
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

func (r *APIManagerRestoreLogicReconciler) reconcileJob(desired *batchv1.Job) (reconcile.Result, error) {
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

func (r *APIManagerRestoreLogicReconciler) reconcileRestoreSecretsAndConfigMapsFromPVCJob() (reconcile.Result, error) {
	desired := r.apiManagerRestore.RestoreSecretsAndConfigMapsFromPVCJob()
	if desired == nil {
		return reconcile.Result{}, nil
	}

	return r.reconcileJob(desired)
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

	res, err := r.reconcileSystemStoragePVC()
	if res.Requeue || err != nil {
		return res, err
	}

	return r.reconcileJob(desired)
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

	res, err := r.reconcileJob(desired)
	if res.Requeue || err != nil {
		return res, err
	}

	// TODO placing this check means that we cannot perform cleanup at the end
	// because otherwise each time is reconciled it will fail because it won't be able
	// to find the secret
	//secret, err := r.sharedBackupSecret()
	//if err != nil {
	//	return reconcile.Result{}, err
	//}
	//if secret == nil {
	//	r.Logger().Info("shared APIManager backup secret has not been created", "secret", r.apiManagerRestore.SecretToShareName())
	// TODO at this point the job was terminated successfully but the secret would not be here.
	// This could happen if there's some bug in the Job code logic. Should we requeue?
	//	return reconcile.Result{Requeue: true}, nil
	//}

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
	return reconcile.Result{}, err

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

func (r *APIManagerRestoreLogicReconciler) reconcileWaitForAPIManagerReady() (reconcile.Result, error) {
	desiredAPIManager, err := r.apiManagerFromSharedBackupSecret()
	if err != nil {
		return reconcile.Result{}, err
	}

	existingAPIManager := &appsv1alpha1.APIManager{}
	err = r.GetResource(common.ObjectKey(desiredAPIManager), existingAPIManager)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Logger().Info("APIManager not found. Waiting until it exists", "APIManager", desiredAPIManager.Name)
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		}
		return reconcile.Result{}, err
	}

	// External databases scenario assumed
	expectedDeploymentNames := []string{
		"apicast-production",
		"apicast-staging",
		"backend-listener",
		"backend-worker",
		"backend-cron",
		"zync",
		"zync-que",
		"zync-database",
		"system-app",
		"system-sphinx",
		"system-sidekiq",
		"system-memcache",
	}

	existingReadyDeployments := existingAPIManager.Status.Deployments.Ready
	sort.Slice(expectedDeploymentNames, func(i, j int) bool { return expectedDeploymentNames[i] < expectedDeploymentNames[j] })
	sort.Slice(existingReadyDeployments, func(i, j int) bool { return existingReadyDeployments[i] < existingReadyDeployments[j] })

	if !reflect.DeepEqual(existingReadyDeployments, expectedDeploymentNames) {
		r.Logger().Info("all APIManager Deployments not ready. Waiting", "APIManager", desiredAPIManager.Name, "expected-ready-deployments", expectedDeploymentNames, "ready-deployments", existingReadyDeployments)
		return reconcile.Result{RequeueAfter: 5 * time.Second, Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileResynchronizeZyncDomains() (reconcile.Result, error) {
	desired := r.apiManagerRestore.ZyncResyncDomainsJob()
	if desired == nil {
		return reconcile.Result{}, nil
	}

	return r.reconcileJob(desired)
}

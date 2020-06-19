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

	if !r.cr.MainStepsCompleted() {
		r.Logger().Info("Reconciling restore steps")
		result, err := r.reconcileMainSteps()
		if result.Requeue || err != nil {
			return result, err
		}

		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	r.Logger().Info("Reconciling post-restore steps")
	result, err := r.reconcilePostRestoreSteps()
	if result.Requeue || err != nil {
		return result, err
	}

	return result, err
}

func (r *APIManagerRestoreLogicReconciler) reconcileMainSteps() (reconcile.Result, error) {
	result, err := r.reconcileStartTimeField()
	if result.Requeue || err != nil {
		return result, err
	}

	result, err = r.reconcileRestoreFromPVCSource()
	if result.Requeue || err != nil {
		return result, err
	}

	result, err = r.reconcileSetMainStepsCompleted()
	if result.Requeue || err != nil {
		return result, err
	}

	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcilePostRestoreSteps() (reconcile.Result, error) {
	result, err := r.reconcileJobsCleanup()
	if result.Requeue || err != nil {
		return result, err
	}

	result, err = r.reconcileRestoreCompletion()
	if result.Requeue || err != nil {
		return result, err
	}

	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileSetMainStepsCompleted() (reconcile.Result, error) {
	if !r.cr.MainStepsCompleted() {
		mainStepsCompleted := true
		r.cr.Status.MainStepsCompleted = &mainStepsCompleted
		err := r.UpdateResourceStatus(r.cr)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil

	}
	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileStartTimeField() (reconcile.Result, error) {
	if r.cr.Status.StartTime == nil {
		startTimeUTC := metav1.Time{Time: clock.Now().UTC()}
		r.cr.Status.StartTime = &startTimeUTC
		err := r.UpdateResourceStatus(r.cr)
		return reconcile.Result{Requeue: true}, err
	}
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

	// Jobs ownerReference or labels nor annotations not reconciled
	// Jobs are one-shot so there's not much point on making updates to them

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
		completionTimeUTC := metav1.Time{Time: clock.Now().UTC()}
		restoreFinished := true
		r.cr.Status.Completed = &restoreFinished
		r.cr.Status.CompletionTime = &completionTimeUTC
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

	if r.cr.Status.APIManagerToRestoreRef == nil {
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
		if err != nil {
			return reconcile.Result{}, err
		}
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

	// Deserialize APIManager using K8s apimachinery decoder in order to
	// deserialize using GKV information (from Scheme) and into
	// a K8s runtime.Object

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
	// At this point and subsequent steps APIManagerToRestoreRef should never be
	// nil and Name should be a non-empty string. Thus, no checks related to
	// that are performed each time the attribute is referenced in steps that
	// are after the step that should set this status value
	apiManagerToRestoreName := r.cr.Status.APIManagerToRestoreRef.Name

	err := r.GetResource(types.NamespacedName{Name: apiManagerToRestoreName, Namespace: r.cr.Namespace}, &appsv1alpha1.APIManager{})
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}
	if err == nil {
		return reconcile.Result{}, nil
	}

	// We proceed here when APIManager has not been found
	apimanager, err := r.apiManagerFromSharedBackupSecret()
	if err != nil {
		return reconcile.Result{}, err
	}

	existing := &appsv1alpha1.APIManager{}
	err = r.ReconcileResource(existing, apimanager, reconcilers.CreateOnlyMutator)
	return reconcile.Result{}, err
}

func (r *APIManagerRestoreLogicReconciler) reconcileAPIManagerBackupSharedInSecretCleanup() (reconcile.Result, error) {
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

	return reconcile.Result{}, nil
}

func (r *APIManagerRestoreLogicReconciler) reconcileWaitForAPIManagerReady() (reconcile.Result, error) {
	existingAPIManager := &appsv1alpha1.APIManager{}
	err := r.GetResource(types.NamespacedName{Name: r.cr.Status.APIManagerToRestoreRef.Name, Namespace: r.cr.Namespace}, existingAPIManager)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Logger().Info("APIManager not found. Waiting until it exists", "APIManager", r.cr.Status.APIManagerToRestoreRef.Name)
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
		r.Logger().Info("all APIManager Deployments not ready. Waiting", "APIManager", existingAPIManager.Name, "expected-ready-deployments", expectedDeploymentNames, "ready-deployments", existingReadyDeployments)
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

// Delete all K8s jobs created during the backup. The reason for this is that
// some PVCs are referenced in the K8s Jobs and those PVCs cannot be deleted
// while some pods reference them, even if in state Completed. By deleting the
// K8s jobs we allow the cleanup to be possible
func (r *APIManagerRestoreLogicReconciler) reconcileJobsCleanup() (reconcile.Result, error) {
	jobsToDelete := []*batchv1.Job{
		r.apiManagerRestore.RestoreSecretsAndConfigMapsFromPVCJob(),
		r.apiManagerRestore.RestoreSystemFileStoragePVCFromPVCJob(),
		r.apiManagerRestore.CreateAPIManagerSharedSecretJob(),
		r.apiManagerRestore.ZyncResyncDomainsJob(),
	}

	existingJobFound := false
	for _, job := range jobsToDelete {
		existingJob := &batchv1.Job{}
		err := r.GetResource(types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, existingJob)
		if err != nil && !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		if err != nil && errors.IsNotFound(err) {
			continue
		}
		existingJobFound = true
		common.TagToObjectDeleteWithPropagationPolicy(job, metav1.DeletePropagationForeground)
		err = r.ReconcileResource(&batchv1.Job{}, job, reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	if existingJobFound {
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

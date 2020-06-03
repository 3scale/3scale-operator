package apimanagerrestore

import (
	"fmt"
	"time"

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

	pipeline, err := r.buildPipeline()
	if err != nil {
		return reconcile.Result{}, err
	}

	pipelineCompleted, err := pipeline.Completed()
	if err != nil {
		return reconcile.Result{}, err
	}
	if pipelineCompleted {
		return reconcile.Result{}, nil
	}
	return pipeline.Execute()
}

func (r *APIManagerRestoreLogicReconciler) buildPipeline() (Pipeline, error) {
	// TODO in the future probably different pipelines will be built depending on
	// restore source (pvc, s3, ...)
	pipelineBuilder := NewRestorePipelineBuilder()
	apiManagerRestoreBaseStep := APIManagerRestoreBaseStep{
		APIManagerRestoreLogicReconciler: r,
	}

	steps := []Step{
		&ReconcileStartTimeStep{apiManagerRestoreBaseStep},
		&ReconcileSecretsAndCfgMapsFromPVCStep{apiManagerRestoreBaseStep},
		&ReconcileAPIManagerInSharedSecretStep{apiManagerRestoreBaseStep},
		&ReconcileCreateSystemFileStoragePVCStep{apiManagerRestoreBaseStep},
		&ReconcileRestoreSystemFileStoragePVCFromPVCStep{apiManagerRestoreBaseStep},
		&ReconcileCreateAPIManagerStep{apiManagerRestoreBaseStep},
		&ReconcileWaitForAPIManagerReadyStep{apiManagerRestoreBaseStep},
		&ReconcileResyncZyncDomainsStep{apiManagerRestoreBaseStep},
		&ReconcileCleanupAPIManagerBackupSharedSecretStep{apiManagerRestoreBaseStep},
		&ReconcileRestoreCompletionStep{apiManagerRestoreBaseStep},
	}
	for _, step := range steps {
		err := pipelineBuilder.AddStep(step)
		if err != nil {
			return nil, nil
		}
	}
	return pipelineBuilder.Build()
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

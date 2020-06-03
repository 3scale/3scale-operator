package apimanagerrestore

import (
	"fmt"
	"time"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/restore"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcileCreateSystemFileStoragePVCStep struct {
	APIManagerRestoreBaseStep
}

func (r *ReconcileCreateSystemFileStoragePVCStep) Execute() (reconcile.Result, error) {
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

func (r *ReconcileCreateSystemFileStoragePVCStep) Completed() (bool, error) {
	exists, err := r.systemStoragePVCExists()
	if err != nil {
		return false, fmt.Errorf("Unknown error executing step '%s'", r.Identifier())
	}

	return exists, nil
}

func (r *ReconcileCreateSystemFileStoragePVCStep) Identifier() string {
	return "ReconcileCreateSystemFileStoragePVCStep"
}

func (r *ReconcileCreateSystemFileStoragePVCStep) systemStoragePVCExists() (bool, error) {
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

func (r *ReconcileCreateSystemFileStoragePVCStep) runtimeRestoreInfoFromAPIManager(apimanager *appsv1alpha1.APIManager) (*restore.RuntimeAPIManagerRestoreInfo, error) {
	var storageClass *string
	if apimanager.Spec.System != nil && apimanager.Spec.System.FileStorageSpec != nil && apimanager.Spec.System.FileStorageSpec.PVC != nil {
		storageClass = apimanager.Spec.System.FileStorageSpec.PVC.StorageClassName
	}
	restoreInfo := &restore.RuntimeAPIManagerRestoreInfo{
		PVCStorageClass: storageClass,
	}
	return restoreInfo, nil
}

package backup

import (
	"context"
	"fmt"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type APIManagerBackupOptionsProvider struct {
	APIManagerBackupCR *appsv1alpha1.APIManagerBackup
	Client             client.Client
}

func NewAPIManagerBackupOptionsProvider(cr *appsv1alpha1.APIManagerBackup, client client.Client) *APIManagerBackupOptionsProvider {
	return &APIManagerBackupOptionsProvider{
		APIManagerBackupCR: cr,
		Client:             client,
	}
}

func (a *APIManagerBackupOptionsProvider) Options() (*APIManagerBackupOptions, error) {
	res := NewAPIManagerBackupOptions()
	res.APIManagerBackupName = a.APIManagerBackupCR.Name
	res.APIManagerBackupUID = a.APIManagerBackupCR.UID
	res.Namespace = a.APIManagerBackupCR.Namespace

	// Should we rely on always having the APIManager existing before doing something?
	// In restores for example it is desirable to not mandate it at all times so it
	// won't be able to properly obtained at option retrieval time. We'll only be able
	// to use the name and Get it when appropriate
	apiManager, err := a.apiManager()
	if err != nil {
		return nil, err
	}
	res.APIManager = apiManager
	res.APIManagerName = apiManager.Name
	res.OCCLIImageURL = a.ocCLIImageURL()

	pvcOptions, err := a.pvcBackupOptions()
	if err != nil {
		return nil, err
	}

	// TODO can this checks be omitted and just rely on the validator package in the APIManagerBackup struct?
	if pvcOptions == nil {
		return nil, fmt.Errorf("at least one backup destination has to be specified")
	}

	res.APIManagerBackupPVCOptions = pvcOptions

	return res, res.Validate()
}

func (a *APIManagerBackupOptionsProvider) pvcBackupOptions() (*APIManagerBackupPVCOptions, error) {
	if a.APIManagerBackupCR.Spec.BackupDestination.PersistentVolumeClaim == nil {
		return nil, nil
	}

	res := NewAPIManagerBackupPVCOptions()
	res.BackupDestinationPVC.Name = fmt.Sprintf("apimanager-backup-%s", a.APIManagerBackupCR.Name)
	res.BackupDestinationPVC.StorageClass = a.APIManagerBackupCR.Spec.BackupDestination.PersistentVolumeClaim.StorageClass
	res.BackupDestinationPVC.VolumeName = a.APIManagerBackupCR.Spec.BackupDestination.PersistentVolumeClaim.VolumeName
	if a.APIManagerBackupCR.Spec.BackupDestination.PersistentVolumeClaim.Resources != nil {
		res.BackupDestinationPVC.StorageRequests = &a.APIManagerBackupCR.Spec.BackupDestination.PersistentVolumeClaim.Resources.Requests
	}

	return res, res.Validate()
}

func (a *APIManagerBackupOptionsProvider) apiManager() (*appsv1alpha1.APIManager, error) {
	return a.autodiscoveredAPIManager()
}

func (a *APIManagerBackupOptionsProvider) autodiscoveredAPIManager() (*appsv1alpha1.APIManager, error) {
	resList := &appsv1alpha1.APIManagerList{}
	err := a.Client.List(context.TODO(), resList, client.InNamespace(a.APIManagerBackupCR.Namespace))
	if err != nil {
		return nil, err
	}

	var res *appsv1alpha1.APIManager
	if len(resList.Items) == 0 {
		return nil, fmt.Errorf("no APIManagers found in namespace '%s'", a.APIManagerBackupCR.Namespace)
	}
	if len(resList.Items) > 1 {
		return nil, fmt.Errorf("multiple APIManagers found in namespace '%s'. Unsupported scenario", a.APIManagerBackupCR.Namespace)
	}

	res = &resList.Items[0]
	return res, nil
}

func (a *APIManagerBackupOptionsProvider) ocCLIImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_OC_CLI", component.OCCLIImageURL())
}

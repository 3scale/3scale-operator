package restore

import (
	"fmt"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type APIManagerRestoreOptionsProvider struct {
	APIManagerRestoreCR *appsv1alpha1.APIManagerRestore
	Client              client.Client
}

func NewAPIManagerRestoreOptionsProvider(cr *appsv1alpha1.APIManagerRestore, client client.Client) *APIManagerRestoreOptionsProvider {
	return &APIManagerRestoreOptionsProvider{
		APIManagerRestoreCR: cr,
		Client:              client,
	}
}

func (a *APIManagerRestoreOptionsProvider) Options() (*APIManagerRestoreOptions, error) {
	res := NewAPIManagerRestoreOptions()
	res.APIManagerRestoreName = a.APIManagerRestoreCR.Name
	res.APIManagerRestoreUID = a.APIManagerRestoreCR.UID
	res.Namespace = a.APIManagerRestoreCR.Namespace

	res.OCCLIImageURL = a.ocCLIImageURL()

	pvcOptions, err := a.pvcRestoreOptions()
	if err != nil {
		return nil, err
	}

	// TODO can this checks be omitted and just rely on the validator package in the APIManagerRestore struct?
	if pvcOptions == nil {
		return nil, fmt.Errorf("At least one restore source has to be specified")
	}

	res.APIManagerRestorePVCOptions = pvcOptions

	return res, res.Validate()
}

func (a *APIManagerRestoreOptionsProvider) pvcRestoreOptions() (*APIManagerRestorePVCOptions, error) {
	if a.APIManagerRestoreCR.Spec.RestoreSource.PersistentVolumeClaim == nil {
		return nil, nil
	}

	res := NewAPIManagerRestorePVCOptions()
	res.PersistentVolumeClaimVolumeSource = a.APIManagerRestoreCR.Spec.RestoreSource.PersistentVolumeClaim.ClaimSource

	return res, res.Validate()
}

func (a *APIManagerRestoreOptionsProvider) ocCLIImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_OC_CLI", component.OCCLIImageURL())
}

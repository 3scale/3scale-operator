package restore

import (
	"fmt"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
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
	res.Namespace = a.APIManagerRestoreCR.Namespace

	pvcOptions, err := a.pvcRestoreOptions()
	if err != nil {
		return nil, err
	}

	s3Options, err := a.s3RestoreOptions()
	if err != nil {
		return nil, err
	}

	// TODO can this checks be omitted and just rely on the validator package in the APIManagerRestore struct?
	if pvcOptions == nil && s3Options == nil {
		return nil, fmt.Errorf("At least one restore source has to be specified")
	}

	if pvcOptions != nil && s3Options != nil {
		return nil, fmt.Errorf("Only one restore source can be specified")
	}

	res.APIManagerRestorePVCOptions = pvcOptions
	res.APIManagerRestoreS3Options = s3Options

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

func (a *APIManagerRestoreOptionsProvider) s3RestoreOptions() (*APIManagerRestoreS3Options, error) {
	if a.APIManagerRestoreCR.Spec.RestoreSource.SimpleStorageService == nil {
		return nil, nil
	}

	res := NewAPIManagerRestoreS3Options()
	return res, res.Validate()
}

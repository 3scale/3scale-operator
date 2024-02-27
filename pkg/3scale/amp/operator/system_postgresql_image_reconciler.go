package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

type SystemPostgreSQLImageReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func SystemPostgreSQLImage(apimanager *appsv1alpha1.APIManager) (*component.SystemPostgreSQLImage, error) {
	optsProvider := NewSystemPostgreSQLImageOptionsProvider(apimanager)
	opts, err := optsProvider.GetSystemPostgreSQLImageOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemPostgreSQLImage(opts), nil
}

package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

type SystemMySQLImageReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func SystemMySQLImage(apimanager *appsv1alpha1.APIManager) (*component.SystemMySQLImage, error) {
	optsProvider := NewSystemMysqlImageOptionsProvider(apimanager)
	opts, err := optsProvider.GetSystemMySQLImageOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemMySQLImage(opts), nil
}

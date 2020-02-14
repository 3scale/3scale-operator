package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
)

type SystemMysqlImageOptionsProvider struct {
	apimanager        *appsv1alpha1.APIManager
	mysqlImageOptions *component.SystemMySQLImageOptions
}

func NewSystemMysqlImageOptionsProvider(apimanager *appsv1alpha1.APIManager) *SystemMysqlImageOptionsProvider {
	return &SystemMysqlImageOptionsProvider{
		apimanager:        apimanager,
		mysqlImageOptions: component.NewSystemMySQLImageOptions(),
	}
}

func (s *SystemMysqlImageOptionsProvider) GetSystemMySQLImageOptions() (*component.SystemMySQLImageOptions, error) {
	s.mysqlImageOptions.AppLabel = *s.apimanager.Spec.AppLabel
	s.mysqlImageOptions.AmpRelease = product.ThreescaleRelease
	s.mysqlImageOptions.InsecureImportPolicy = s.apimanager.Spec.ImageStreamTagImportInsecure

	s.mysqlImageOptions.Image = component.SystemMySQLImageURL()
	if s.apimanager.Spec.System.DatabaseSpec != nil &&
		s.apimanager.Spec.System.DatabaseSpec.MySQL != nil &&
		s.apimanager.Spec.System.DatabaseSpec.MySQL.Image != nil {
		s.mysqlImageOptions.Image = *s.apimanager.Spec.System.DatabaseSpec.MySQL.Image
	}

	err := s.mysqlImageOptions.Validate()
	return s.mysqlImageOptions, err
}

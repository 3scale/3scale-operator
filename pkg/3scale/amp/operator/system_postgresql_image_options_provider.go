package operator

import (
	appsv1beta1 "github.com/3scale/3scale-operator/apis/apps/v1beta1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
)

type SystemPostgreSQLImageOptionsProvider struct {
	apimanager *appsv1beta1.APIManager
	options    *component.SystemPostgreSQLImageOptions
}

func NewSystemPostgreSQLImageOptionsProvider(apimanager *appsv1beta1.APIManager) *SystemPostgreSQLImageOptionsProvider {
	return &SystemPostgreSQLImageOptionsProvider{
		apimanager: apimanager,
		options:    component.NewSystemPostgreSQLImageOptions(),
	}
}

func (s *SystemPostgreSQLImageOptionsProvider) GetSystemPostgreSQLImageOptions() (*component.SystemPostgreSQLImageOptions, error) {
	s.options.AppLabel = *s.apimanager.Spec.AppLabel
	s.options.AmpRelease = product.ThreescaleRelease
	s.options.InsecureImportPolicy = s.apimanager.Spec.ImageStreamTagImportInsecure

	s.options.Image = SystemPostgreSQLImageURL()
	if s.apimanager.Spec.System.DatabaseSpec != nil &&
		s.apimanager.Spec.System.DatabaseSpec.PostgreSQL != nil &&
		s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.Image != nil {
		s.options.Image = *s.apimanager.Spec.System.DatabaseSpec.PostgreSQL.Image
	}

	err := s.options.Validate()
	return s.options, err
}

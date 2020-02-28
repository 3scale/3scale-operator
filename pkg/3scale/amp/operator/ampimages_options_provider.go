package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
)

type AmpImagesOptionsProvider struct {
	apimanager       *appsv1alpha1.APIManager
	ampImagesOptions *component.AmpImagesOptions
}

func NewAmpImagesOptionsProvider(apimanager *appsv1alpha1.APIManager) *AmpImagesOptionsProvider {
	return &AmpImagesOptionsProvider{
		apimanager:       apimanager,
		ampImagesOptions: component.NewAmpImagesOptions(),
	}
}

func (a *AmpImagesOptionsProvider) GetAmpImagesOptions() (*component.AmpImagesOptions, error) {
	a.ampImagesOptions.AppLabel = *a.apimanager.Spec.AppLabel
	a.ampImagesOptions.AmpRelease = product.ThreescaleRelease
	a.ampImagesOptions.InsecureImportPolicy = *a.apimanager.Spec.ImageStreamTagImportInsecure

	a.ampImagesOptions.ApicastImage = ApicastImageURL()
	if a.apimanager.Spec.Apicast != nil && a.apimanager.Spec.Apicast.Image != nil {
		a.ampImagesOptions.ApicastImage = *a.apimanager.Spec.Apicast.Image
	}

	a.ampImagesOptions.BackendImage = BackendImageURL()
	if a.apimanager.Spec.Backend != nil && a.apimanager.Spec.Backend.Image != nil {
		a.ampImagesOptions.BackendImage = *a.apimanager.Spec.Backend.Image
	}

	a.ampImagesOptions.SystemImage = SystemImageURL()
	if a.apimanager.Spec.System != nil && a.apimanager.Spec.System.Image != nil {
		a.ampImagesOptions.SystemImage = *a.apimanager.Spec.System.Image
	}

	a.ampImagesOptions.ZyncImage = ZyncImageURL()
	if a.apimanager.Spec.Zync != nil && a.apimanager.Spec.Zync.Image != nil {
		a.ampImagesOptions.ZyncImage = *a.apimanager.Spec.Zync.Image
	}

	a.ampImagesOptions.ZyncDatabasePostgreSQLImage = ZyncPostgreSQLImageURL()
	if a.apimanager.Spec.Zync != nil && a.apimanager.Spec.Zync.PostgreSQLImage != nil {
		a.ampImagesOptions.ZyncDatabasePostgreSQLImage = *a.apimanager.Spec.Zync.PostgreSQLImage
	}

	a.ampImagesOptions.SystemMemcachedImage = SystemMemcachedImageURL()
	if a.apimanager.Spec.System != nil && a.apimanager.Spec.System.MemcachedImage != nil {
		a.ampImagesOptions.SystemMemcachedImage = *a.apimanager.Spec.System.MemcachedImage
	}

	err := a.ampImagesOptions.Validate()
	return a.ampImagesOptions, err
}

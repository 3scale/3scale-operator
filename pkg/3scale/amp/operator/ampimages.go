package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorAmpImagesOptionsProvider) GetAmpImagesOptions() (*component.AmpImagesOptions, error) {
	optProv := component.AmpImagesOptionsBuilder{}
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.AMPRelease(o.APIManagerSpec.AmpRelease)
	optProv.ApicastImage(*o.APIManagerSpec.AmpApicastImage)
	optProv.BackendImage(*o.APIManagerSpec.AmpBackendImage)
	optProv.RouterImage(*o.APIManagerSpec.AmpRouterImage)
	optProv.SystemImage(*o.APIManagerSpec.AmpSystemImage)
	optProv.ZyncImage(*o.APIManagerSpec.AmpZyncImage)
	optProv.PostgreSQLImage(*o.APIManagerSpec.PostgreSQLImage)
	optProv.InsecureImportPolicy(*o.APIManagerSpec.ImageStreamTagImportInsecure)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create AMPImages Options - %s", err)
	}
	return res, nil
}

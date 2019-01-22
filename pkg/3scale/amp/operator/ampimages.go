package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorAmpImagesOptionsProvider) GetAmpImagesOptions() (*component.AmpImagesOptions, error) {
	optProv := component.AmpImagesOptionsBuilder{}
	optProv.AppLabel(*o.AmpSpec.AppLabel)
	optProv.AMPRelease(o.AmpSpec.AmpRelease)
	optProv.ApicastImage(*o.AmpSpec.AmpApicastImage)
	optProv.BackendImage(*o.AmpSpec.AmpBackendImage)
	optProv.RouterImage(*o.AmpSpec.AmpRouterImage)
	optProv.SystemImage(*o.AmpSpec.AmpSystemImage)
	optProv.ZyncImage(*o.AmpSpec.AmpZyncImage)
	optProv.PostgreSQLImage(*o.AmpSpec.PostgreSQLImage)
	optProv.InsecureImportPolicy(*o.AmpSpec.ImageStreamTagImportInsecure)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create AMPImages Options - %s", err)
	}
	return res, nil
}

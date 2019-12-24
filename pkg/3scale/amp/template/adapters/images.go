package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type ImagesAdapter struct {
}

func NewImagesAdapter() Adapter {
	return NewAppenderAdapter(&ImagesAdapter{})
}

func (i *ImagesAdapter) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		templatev1.Parameter{
			Name:     "AMP_BACKEND_IMAGE",
			Required: true,
			Value:    component.BackendImageURL(),
		},
		templatev1.Parameter{
			Name:     "AMP_ZYNC_IMAGE",
			Value:    component.ZyncImageURL(),
			Required: true,
		},
		templatev1.Parameter{
			Name:     "AMP_APICAST_IMAGE",
			Value:    component.ApicastImageURL(),
			Required: true,
		},
		templatev1.Parameter{
			Name:     "AMP_SYSTEM_IMAGE",
			Value:    component.SystemImageURL(),
			Required: true,
		},
		templatev1.Parameter{
			Name:        "ZYNC_DATABASE_IMAGE",
			Description: "Zync's PostgreSQL image to use",
			Value:       component.ZyncPostgreSQLImageURL(),
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "MEMCACHED_IMAGE",
			Description: "Memcached image to use",
			Value:       component.SystemMemcachedImageURL(),
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "IMAGESTREAM_TAG_IMPORT_INSECURE",
			Description: "Set to true if the server may bypass certificate verification or connect directly over HTTP during image import.",
			Value:       "false",
			Required:    true,
		},
	}
}

func (i *ImagesAdapter) Objects() ([]common.KubernetesObject, error) {
	imagesOptions, err := i.options()
	if err != nil {
		return nil, err
	}
	imagesComponent := component.NewAmpImages(imagesOptions)
	return imagesComponent.Objects(), nil
}

func (i *ImagesAdapter) options() (*component.AmpImagesOptions, error) {
	ao := component.NewAmpImagesOptions()
	ao.AppLabel = "${APP_LABEL}"
	ao.AmpRelease = "${AMP_RELEASE}"
	ao.ApicastImage = "${AMP_APICAST_IMAGE}"
	ao.BackendImage = "${AMP_BACKEND_IMAGE}"
	ao.SystemImage = "${AMP_SYSTEM_IMAGE}"
	ao.ZyncImage = "${AMP_ZYNC_IMAGE}"
	ao.ZyncDatabasePostgreSQLImage = "${ZYNC_DATABASE_IMAGE}"
	ao.SystemMemcachedImage = "${MEMCACHED_IMAGE}"
	ao.InsecureImportPolicy = false

	err := ao.Validate()
	return ao, err
}

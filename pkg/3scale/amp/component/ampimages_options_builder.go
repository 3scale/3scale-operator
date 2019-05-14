package component

import "fmt"

type AmpImagesOptionsBuilder struct {
	options AmpImagesOptions
}

func (ampImages *AmpImagesOptionsBuilder) AppLabel(appLabel string) {
	ampImages.options.appLabel = appLabel
}

func (ampImages *AmpImagesOptionsBuilder) AMPRelease(ampRelease string) {
	ampImages.options.ampRelease = ampRelease
}

func (ampImages *AmpImagesOptionsBuilder) ApicastImage(apicastImage string) {
	ampImages.options.apicastImage = apicastImage
}

func (ampImages *AmpImagesOptionsBuilder) BackendImage(backendImage string) {
	ampImages.options.backendImage = backendImage
}

func (ampImages *AmpImagesOptionsBuilder) RouterImage(routerImage string) {
	ampImages.options.routerImage = routerImage
}

func (ampImages *AmpImagesOptionsBuilder) SystemImage(systemImage string) {
	ampImages.options.systemImage = systemImage
}

func (ampImages *AmpImagesOptionsBuilder) ZyncImage(zyncImage string) {
	ampImages.options.zyncImage = zyncImage
}

func (ampImages *AmpImagesOptionsBuilder) PostgreSQLImage(postgreSQLImage string) {
	ampImages.options.postgreSQLImage = postgreSQLImage
}

func (ampImages *AmpImagesOptionsBuilder) BackendRedisImage(image string) {
	ampImages.options.backendRedisImage = image
}

func (ampImages *AmpImagesOptionsBuilder) SystemRedisImage(image string) {
	ampImages.options.systemRedisImage = image
}

func (ampImages *AmpImagesOptionsBuilder) SystemMemcachedImage(image string) {
	ampImages.options.systemMemcachedImage = image
}

func (ampImages *AmpImagesOptionsBuilder) InsecureImportPolicy(insecureImportPolicy bool) {
	ampImages.options.insecureImportPolicy = insecureImportPolicy
}

func (ampImages *AmpImagesOptionsBuilder) Build() (*AmpImagesOptions, error) {
	if ampImages.options.appLabel == "" {
		return nil, fmt.Errorf("no AppLabel has been provided")
	}
	if ampImages.options.ampRelease == "" {
		return nil, fmt.Errorf("no AMP release has been provided")
	}
	if ampImages.options.apicastImage == "" {
		return nil, fmt.Errorf("no Apicast image has been provided")
	}
	if ampImages.options.backendImage == "" {
		return nil, fmt.Errorf("no Backend image has been provided")
	}
	if ampImages.options.routerImage == "" {
		return nil, fmt.Errorf("no Router image been provided")
	}
	if ampImages.options.systemImage == "" {
		return nil, fmt.Errorf("no System image has been provided")
	}
	if ampImages.options.zyncImage == "" {
		return nil, fmt.Errorf("no Zync image has been provided")
	}
	if ampImages.options.postgreSQLImage == "" {
		return nil, fmt.Errorf("no PostgreSQL image has been provided")
	}
	if ampImages.options.backendRedisImage == "" {
		return nil, fmt.Errorf("no Backend Redis image has been provided")
	}
	if ampImages.options.systemRedisImage == "" {
		return nil, fmt.Errorf("no System Redis image has been provided")
	}
	if ampImages.options.systemMemcachedImage == "" {
		return nil, fmt.Errorf("no System Memcached image has been provided")
	}

	return &ampImages.options, nil
}

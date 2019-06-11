package component

import "fmt"

type ProductizedOptionsBuilder struct {
	options ProductizedOptions
}

func (productized *ProductizedOptionsBuilder) AmpRelease(ampRelease string) {
	productized.options.ampRelease = ampRelease
}

func (productized *ProductizedOptionsBuilder) ApicastImage(apicastImage string) {
	productized.options.apicastImage = apicastImage
}

func (productized *ProductizedOptionsBuilder) BackendImage(backendImage string) {
	productized.options.backendImage = backendImage
}

func (productized *ProductizedOptionsBuilder) SystemImage(systemImage string) {
	productized.options.systemImage = systemImage
}

func (productized *ProductizedOptionsBuilder) ZyncImage(zyncImage string) {
	productized.options.zyncImage = zyncImage
}

func (productized *ProductizedOptionsBuilder) Build() (*ProductizedOptions, error) {
	err := productized.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	productized.setNonRequiredOptions()

	return &productized.options, nil
}

func (productized *ProductizedOptionsBuilder) setRequiredOptions() error {
	if productized.options.ampRelease == "" {
		return fmt.Errorf("no AMP release has been provided")
	}
	if productized.options.apicastImage == "" {
		return fmt.Errorf("no Apicast image has been provided")
	}
	if productized.options.backendImage == "" {
		return fmt.Errorf("no Backend image has been provided")
	}
	if productized.options.systemImage == "" {
		return fmt.Errorf("no System image has been provided")
	}
	if productized.options.zyncImage == "" {
		return fmt.Errorf("no Zync image has been provided")
	}
	return nil
}

func (productized *ProductizedOptionsBuilder) setNonRequiredOptions() {
}

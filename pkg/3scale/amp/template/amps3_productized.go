package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

// AmpS3ProductizedTemplateAdapters defines the list of adapters to build the template
func AmpS3ProductizedTemplateAdapters(options []string) []adapters.Adapter {
	adapterList := AmpS3TemplateAdapters(options)

	return append(adapterList, adapters.NewProductizedAdapter(options))
}

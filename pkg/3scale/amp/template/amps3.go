package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

// AmpS3TemplateAdapters defines the list of adapters to build the template
func AmpS3TemplateAdapters(options []string) []adapters.Adapter {
	adapterList := AmpTemplateAdapters(options)

	return append(adapterList, adapters.NewS3Adapter(options))
}

package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

// AmpEvalS3ProductizedTemplateAdapters defines the list of adapters to build the template
func AmpEvalS3ProductizedTemplateAdapters(options []string) []adapters.Adapter {
	adapterList := AmpEvalS3TemplateAdapters(options)

	return append(adapterList, adapters.NewProductizedAdapter(options))
}

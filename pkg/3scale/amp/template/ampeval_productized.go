package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

// AmpEvalProductizedTemplateAdapters defines the list of adapters to build the template
func AmpEvalProductizedTemplateAdapters(options []string) []adapters.Adapter {
	adapterList := AmpEvalTemplateAdapters(options)

	return append(adapterList, adapters.NewProductizedAdapter(options))
}

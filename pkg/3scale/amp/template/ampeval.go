package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

// AmpEvalTemplateAdapters defines the list of adapters to build the template
func AmpEvalTemplateAdapters(options []string) []adapters.Adapter {
	adapterList := AmpTemplateAdapters(options)

	return append(adapterList, adapters.NewEvalAdapter(options))
}

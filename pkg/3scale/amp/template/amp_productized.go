package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

// AmpProductizedTemplateAdapters defines the list of adapters to build the template
func AmpProductizedTemplateAdapters(options []string) []adapters.Adapter {
	adapterList := AmpTemplateAdapters(options)

	return append(adapterList, adapters.NewProductizedAdapter(options))
}

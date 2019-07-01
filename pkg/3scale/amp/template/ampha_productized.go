package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

// AmpHAProductizedTemplateAdapters defines the list of adapters to build the template
func AmpHAProductizedTemplateAdapters(options []string) []adapters.Adapter {
	adapterList := AmpHATemplateAdapters(options)

	return append(adapterList, adapters.NewProductizedAdapter(options))
}

package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

// AmpS3ProductizedTemplateAdapters defines the list of adapters to build the template
func AmpPostgresqlProductizedTemplateAdapters(options []string) []adapters.Adapter {
	adapterList := AmpPostgresqlTemplateAdapters(options)

	return append(adapterList, adapters.NewProductizedAdapter(options))
}

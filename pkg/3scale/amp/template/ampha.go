package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

// AmpHATemplateAdapters defines the list of adapters to build the template
func AmpHATemplateAdapters(options []string) []adapters.Adapter {
	return []adapters.Adapter{
		adapters.NewImagesAdapter(options),
		adapters.NewRedisAdapter(options),
		adapters.NewBackendAdapter(options),
		adapters.NewMemcachedAdapter(options),
		adapters.NewSystemAdapter(options),
		adapters.NewZyncAdapter(options),
		adapters.NewApicastAdapter(options),
		adapters.NewHAAdapter(options),
	}
}

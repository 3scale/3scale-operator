package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

func init() {
	// TemplateFactories is a list of template factories
	TemplateFactories = append(TemplateFactories, NewAmpHATemplateFactory)
}

type AmpHATemplateFactory struct {
}

func (f *AmpHATemplateFactory) Adapters() []adapters.Adapter {
	return []adapters.Adapter{
		adapters.NewImagesAdapter(),
		adapters.NewBackendRedisAdapter(),
		adapters.NewSystemRedisAdapter(),
		adapters.NewBackendAdapter(true),
		adapters.NewMemcachedAdapter(),
		adapters.NewSystemAdapter(true),
		adapters.NewZyncAdapter(true),
		adapters.NewApicastAdapter(true),
		adapters.NewHAAdapter(),
	}
}

func (f *AmpHATemplateFactory) Type() TemplateType {
	return "amp-ha-template"
}

func NewAmpHATemplateFactory() TemplateFactory {
	return &AmpHATemplateFactory{}
}

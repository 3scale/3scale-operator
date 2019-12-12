package template

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"
	templatev1 "github.com/openshift/api/template/v1"
)

func init() {
	// TemplateFactories is a list of template factories
	TemplateFactories = append(TemplateFactories, NewAmpHATemplateFactory)
}

type AmpHATemplateAdapter struct {
}

func (e *AmpHATemplateAdapter) Adapt(template *templatev1.Template) {
	template.Name = "3scale-api-management-ha"
	template.Annotations["description"] = "3scale API Management main system (High Availability)"
}

type AmpHATemplateFactory struct {
}

func (f *AmpHATemplateFactory) Adapters() []adapters.Adapter {
	return []adapters.Adapter{
		adapters.NewImagesAdapter(),
		adapters.NewRedisAdapter(),
		adapters.NewBackendAdapter(),
		adapters.NewMemcachedAdapter(),
		adapters.NewSystemAdapter(),
		adapters.NewZyncAdapter(),
		adapters.NewApicastAdapter(),
		adapters.NewHAAdapter(),
		&AmpHATemplateAdapter{},
	}
}

func (f *AmpHATemplateFactory) Type() TemplateType {
	return "amp-ha-template"
}

func NewAmpHATemplateFactory() TemplateFactory {
	return &AmpHATemplateFactory{}
}

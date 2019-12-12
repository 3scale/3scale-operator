package template

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"
	templatev1 "github.com/openshift/api/template/v1"
)

func init() {
	// TemplateFactories is a list of template factories
	TemplateFactories = append(TemplateFactories, NewAmpTemplateFactory)
}

type AmpTemplateAdapter struct {
}

func (a *AmpTemplateAdapter) Adapt(template *templatev1.Template) {
	template.Name = "3scale-api-management"
}

type AmpTemplateFactory struct {
}

func (f *AmpTemplateFactory) Adapters() []adapters.Adapter {
	return []adapters.Adapter{
		adapters.NewImagesAdapter(),
		adapters.NewSystemMysqlImageAdapter(),
		adapters.NewRedisAdapter(),
		adapters.NewBackendAdapter(),
		adapters.NewMysqlAdapter(),
		adapters.NewMemcachedAdapter(),
		adapters.NewSystemAdapter(component.SystemFileStorageTypePVC),
		adapters.NewZyncAdapter(),
		adapters.NewApicastAdapter(),
		&AmpTemplateAdapter{},
	}
}

func (f *AmpTemplateFactory) Type() TemplateType {
	return "amp-template"
}

func NewAmpTemplateFactory() TemplateFactory {
	return &AmpTemplateFactory{}
}

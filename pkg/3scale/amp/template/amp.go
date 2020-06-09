package template

import (
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
	template.ObjectMeta.Name = "3scale-api-management"
	template.Message = "Login on https://${TENANT_NAME}-admin.${WILDCARD_DOMAIN} as ${ADMIN_USERNAME}/${ADMIN_PASSWORD}"
}

type AmpTemplateFactory struct {
}

func (f *AmpTemplateFactory) Adapters() []adapters.Adapter {
	return []adapters.Adapter{
		adapters.NewImagesAdapter(),
		adapters.NewSystemMysqlImageAdapter(),
		adapters.NewCommonEmbeddedRedisAdapter(),
		adapters.NewBackendRedisAdapter(),
		adapters.NewSystemRedisAdapter(),
		adapters.NewBackendAdapter(false),
		adapters.NewMysqlAdapter(),
		adapters.NewMemcachedAdapter(),
		adapters.NewSystemAdapter(false),
		adapters.NewZyncAdapter(false),
		adapters.NewApicastAdapter(false),
		&AmpTemplateAdapter{},
	}
}

func (f *AmpTemplateFactory) Type() TemplateType {
	return "amp-template"
}

func NewAmpTemplateFactory() TemplateFactory {
	return &AmpTemplateFactory{}
}

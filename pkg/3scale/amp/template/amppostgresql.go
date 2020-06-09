package template

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"
	templatev1 "github.com/openshift/api/template/v1"
)

func init() {
	// TemplateFactories is a list of template factories
	TemplateFactories = append(TemplateFactories, NewAmpPostgresqlTemplateFactory)
}

type AmpPostgresqlTemplateAdapter struct {
}

func (a *AmpPostgresqlTemplateAdapter) Adapt(template *templatev1.Template) {
	template.ObjectMeta.Name = "3scale-api-management-postgresql"
	template.Message = "Login on https://${TENANT_NAME}-admin.${WILDCARD_DOMAIN} as ${ADMIN_USERNAME}/${ADMIN_PASSWORD}"
	template.ObjectMeta.Annotations = a.buildAmpTemplateMetaAnnotations()
}

func (a *AmpPostgresqlTemplateAdapter) buildAmpTemplateMetaAnnotations() map[string]string {
	annotations := map[string]string{
		"openshift.io/display-name":          "3scale API Management",
		"openshift.io/provider-display-name": "Red Hat, Inc.",
		"iconClass":                          "icon-3scale",
		"description":                        "3scale API Management main system with PostgreSQL as System's database",
		"tags":                               "integration, api management, 3scale",
	}

	return annotations
}

type AmpPostgresqlTemplateFactory struct {
}

func (f *AmpPostgresqlTemplateFactory) Adapters() []adapters.Adapter {
	return []adapters.Adapter{
		adapters.NewImagesAdapter(),
		adapters.NewSystemPostgreSQLImageAdapter(),
		adapters.NewCommonEmbeddedRedisAdapter(),
		adapters.NewBackendRedisAdapter(),
		adapters.NewSystemRedisAdapter(),
		adapters.NewBackendAdapter(false),
		adapters.NewSystemPostgreSQLAdapter(),
		adapters.NewMemcachedAdapter(),
		adapters.NewSystemAdapter(false),
		adapters.NewZyncAdapter(false),
		adapters.NewApicastAdapter(false),
		&AmpPostgresqlTemplateAdapter{},
	}
}

func (f *AmpPostgresqlTemplateFactory) Type() TemplateType {
	return "amp-postgresql-template"
}

func NewAmpPostgresqlTemplateFactory() TemplateFactory {
	return &AmpPostgresqlTemplateFactory{}
}

package template

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"
	templatev1 "github.com/openshift/api/template/v1"
)

func init() {
	// TemplateFactories is a list of template factories
	TemplateFactories = append(TemplateFactories, NewAmpPostgresqlTemplateFactory)
}

type AmpPostgreSQLTemplateAdapter struct {
}

func (a *AmpPostgreSQLTemplateAdapter) Adapt(template *templatev1.Template) {
	template.Name = "3scale-api-management-postgresql"
	template.Annotations["description"] = "3scale API Management main system with PostgreSQL as System's database"
}

type AmpPostgresqlTemplateFactory struct {
}

func (f *AmpPostgresqlTemplateFactory) Adapters() []adapters.Adapter {
	return []adapters.Adapter{
		adapters.NewImagesAdapter(),
		adapters.NewSystemPostgreSQLImageAdapter(),
		adapters.NewRedisAdapter(),
		adapters.NewBackendAdapter(),
		adapters.NewSystemPostgreSQLAdapter(),
		adapters.NewMemcachedAdapter(),
		adapters.NewSystemAdapter(),
		adapters.NewZyncAdapter(),
		adapters.NewApicastAdapter(),
		&AmpPostgreSQLTemplateAdapter{},
	}
}

func (f *AmpPostgresqlTemplateFactory) Type() TemplateType {
	return "amp-postgresql-template"
}

func NewAmpPostgresqlTemplateFactory() TemplateFactory {
	return &AmpPostgresqlTemplateFactory{}
}

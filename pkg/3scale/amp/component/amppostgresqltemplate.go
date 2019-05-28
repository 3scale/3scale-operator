package component

import (
	templatev1 "github.com/openshift/api/template/v1"
)

type AmpPostgreSQLTemplate struct {
	options    []string
	components []Component
}

type AmpPostgreSQLTemplateOptions struct {
	ampImagesOptions             AmpImagesOptions
	redisOptions                 RedisOptions
	backendOptions               BackendOptions
	systemPostgreSQLOptions      SystemPostgreSQLOptions
	systemPostgreSQLImageOptions SystemPostgreSQLImageOptions
	memcachedOptions             MemcachedOptions
	systemOptions                SystemOptions
	zyncOptions                  ZyncOptions
	apicastOptions               ApicastOptions
	wildcardRouterOptions        WildcardRouterOptions
}

func NewAmpPostgreSQLTemplate(options []string) *AmpTemplate {
	components := []Component{
		NewAmpImages(options),
		NewSystemPostgreSQLImage(options),
		NewRedis(options),
		NewBackend(options),
		NewSystemPostgreSQL(options),
		NewMemcached(options),
		NewSystem(options),
		NewZync(options),
		NewApicast(options),
		NewWildcardRouter(options),
	}

	ampTemplate := &AmpTemplate{
		options:    options,
		components: components,
	}
	return ampTemplate
}

func (a *AmpPostgreSQLTemplate) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	for componentIdx := range a.components {
		a.components[componentIdx].AssembleIntoTemplate(template, otherComponents)
	}
	a.setTemplateFields(template)
}

func (a *AmpPostgreSQLTemplate) PostProcess(template *templatev1.Template, otherComponents []Component) {
	for componentIdx := range a.components {
		a.components[componentIdx].PostProcess(template, otherComponents)
	}
}

func (a *AmpPostgreSQLTemplate) setTemplateFields(template *templatev1.Template) {
	template.ObjectMeta.Name = "3scale-api-management-postgresql"
	template.ObjectMeta.Annotations = a.buildAmpTemplateMetaAnnotations()
	template.Message = "Login on https://${TENANT_NAME}-admin.${WILDCARD_DOMAIN} as ${ADMIN_USERNAME}/${ADMIN_PASSWORD}"
}

func (a *AmpPostgreSQLTemplate) buildAmpTemplateMetaAnnotations() map[string]string {
	annotations := map[string]string{
		"openshift.io/display-name":          "3scale API Management",
		"openshift.io/provider-display-name": "Red Hat, Inc.",
		"iconClass":                          "icon-3scale",
		"description":                        "3scale API Management main system with PostgreSQL as System's database",
		"tags":                               "integration, api management, 3scale",
	}

	return annotations
}

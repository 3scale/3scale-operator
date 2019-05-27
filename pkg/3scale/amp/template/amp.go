package template

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"
	templatev1 "github.com/openshift/api/template/v1"
)

type AmpTemplateAdapter struct {
}

func (a *AmpTemplateAdapter) Adapt(template *templatev1.Template) {
	template.ObjectMeta.Name = "3scale-api-management"
	template.Message = "Login on https://${TENANT_NAME}-admin.${WILDCARD_DOMAIN} as ${ADMIN_USERNAME}/${ADMIN_PASSWORD}"
}

// AmpTemplateAdapters defines the list of adapters to build the template
func AmpTemplateAdapters(options []string) []adapters.Adapter {
	return []adapters.Adapter{
		adapters.NewImagesAdapter(options),
		adapters.NewSystemMysqlImageAdapter(options),
		adapters.NewRedisAdapter(options),
		adapters.NewBackendAdapter(options),
		adapters.NewMysqlAdapter(options),
		adapters.NewMemcachedAdapter(options),
		adapters.NewSystemAdapter(options),
		adapters.NewZyncAdapter(options),
		adapters.NewApicastAdapter(options),
		&AmpTemplateAdapter{},
	}
}

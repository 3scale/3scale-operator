package component

import (
	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	"k8s.io/api/core/v1"
)

type Productized struct {
	options []string
}

func NewProductized(options []string) *Productized {
	productized := &Productized{
		options: options,
	}
	return productized
}

func (productized *Productized) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
}

func (productized *Productized) PostProcess(template *templatev1.Template, otherComponents []Component) {
	productized.removeAmpServiceAccount(template)
	productized.removeAmpServiceAccountReferences(template)
	productized.UpdateAmpImagesURIs(template)
}

func (productized *Productized) removeAmpServiceAccount(template *templatev1.Template) {
	for idx, rawExtension := range template.Objects {
		obj := rawExtension.Object
		sa, ok := obj.(*v1.ServiceAccount)
		if ok {
			if sa.ObjectMeta.Name == "amp" {
				template.Objects = append(template.Objects[:idx], template.Objects[idx+1:]...) // This deletes the element in the array
				break
			}
		}
	}
}

func (productized *Productized) removeAmpServiceAccountReferences(template *templatev1.Template) {
	for _, rawExtension := range template.Objects {
		obj := rawExtension.Object
		dc, ok := obj.(*appsv1.DeploymentConfig)
		if ok {
			dc.Spec.Template.Spec.ServiceAccountName = ""
		}
	}
}

func (productized *Productized) UpdateAmpImagesURIs(template *templatev1.Template) {
	for paramIdx := range template.Parameters {
		param := &template.Parameters[paramIdx]
		switch param.Name {
		case "AMP_SYSTEM_IMAGE":
			param.Value = "registry.access.redhat.com/3scale-amp24/system"
		case "AMP_BACKEND_IMAGE":
			param.Value = "registry.access.redhat.com/3scale-amp24/backend"
		case "AMP_APICAST_IMAGE":
			param.Value = "registry.access.redhat.com/3scale-amp24/apicast-gateway"
		case "AMP_ROUTER_IMAGE":
			param.Value = "registry.access.redhat.com/3scale-amp22/wildcard-router"
		case "AMP_ZYNC_IMAGE":
			param.Value = "registry.access.redhat.com/3scale-amp24/zync"
		case "MEMCACHED_IMAGE":
			param.Value = "registry.access.redhat.com/3scale-amp20/memcached"
		}
	}
}

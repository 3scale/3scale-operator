package component

import (
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Productized struct {
	options []string
	Options *ProductizedOptions
}

func NewProductized(options []string) *Productized {
	productized := &Productized{
		options: options,
	}
	return productized
}

type ProductizedOptions struct {
	productizedNonRequiredOptions
	productizedRequiredOptions
}

type productizedRequiredOptions struct {
	ampRelease     string
	apicastImage   string
	backendImage   string
	routerImage    string
	systemImage    string
	zyncImage      string
	memcachedImage string
}

type productizedNonRequiredOptions struct {
}

type ProductizedOptionsBuilder struct {
	options ProductizedOptions
}

func (productized *ProductizedOptionsBuilder) AmpRelease(ampRelease string) {
	productized.options.ampRelease = ampRelease
}

func (productized *ProductizedOptionsBuilder) ApicastImage(apicastImage string) {
	productized.options.apicastImage = apicastImage
}

func (productized *ProductizedOptionsBuilder) BackendImage(backendImage string) {
	productized.options.backendImage = backendImage
}

func (productized *ProductizedOptionsBuilder) RouterImage(routerImage string) {
	productized.options.routerImage = routerImage
}

func (productized *ProductizedOptionsBuilder) SystemImage(systemImage string) {
	productized.options.systemImage = systemImage
}

func (productized *ProductizedOptionsBuilder) ZyncImage(zyncImage string) {
	productized.options.zyncImage = zyncImage
}

func (productized *ProductizedOptionsBuilder) MemcachedImage(memcachedImage string) {
	productized.options.memcachedImage = memcachedImage
}

func (productized *ProductizedOptionsBuilder) Build() (*ProductizedOptions, error) {
	err := productized.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	productized.setNonRequiredOptions()

	return &productized.options, nil
}

func (productized *ProductizedOptionsBuilder) setRequiredOptions() error {
	if productized.options.ampRelease == "" {
		return fmt.Errorf("no AMP release has been provided")
	}
	if productized.options.apicastImage == "" {
		return fmt.Errorf("no Apicast image has been provided")
	}
	if productized.options.backendImage == "" {
		return fmt.Errorf("no Backend image has been provided")
	}
	if productized.options.routerImage == "" {
		return fmt.Errorf("no Router image has been provided")
	}
	if productized.options.systemImage == "" {
		return fmt.Errorf("no System image has been provided")
	}
	if productized.options.zyncImage == "" {
		return fmt.Errorf("no Zync image has been provided")
	}
	if productized.options.memcachedImage == "" {
		return fmt.Errorf("no Memcached image has been provided")
	}
	return nil
}

func (productized *ProductizedOptionsBuilder) setNonRequiredOptions() {
}

type ProductizedOptionsProvider interface {
	GetProductizedOptions() *ProductizedOptions
}
type CLIProductizedOptionsProvider struct {
}

func (o *CLIProductizedOptionsProvider) GetProductizedOptions() (*ProductizedOptions, error) {
	pob := ProductizedOptionsBuilder{}
	pob.ApicastImage("${AMP_APICAST_IMAGE}")
	pob.BackendImage("${AMP_BACKEND_IMAGE}")
	pob.RouterImage("${AMP_ROUTER_IMAGE}")
	pob.SystemImage("${AMP_SYSTEM_IMAGE}")
	pob.ZyncImage("${AMP_ZYNC_IMAGE}")
	pob.MemcachedImage("${MEMCACHED_IMAGE}")
	res, err := pob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Productized Options - %s", err)
	}
	return res, nil
}

func (productized *Productized) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
}

func (productized *Productized) PostProcess(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLIProductizedOptionsProvider{}
	productizedOpts, err := optionsProvider.GetProductizedOptions()
	_ = err
	productized.Options = productizedOpts
	res := template.Objects
	res = productized.removeAmpServiceAccount(res)
	res = productized.removeAmpServiceAccountReferences(res)
	productized.updateAmpImagesParameters(template)
	template.Objects = res
}

func (productized *Productized) PostProcessObjects(objects []runtime.RawExtension) []runtime.RawExtension {
	res := objects
	res = productized.removeAmpServiceAccount(res)
	res = productized.removeAmpServiceAccountReferences(res)
	res = productized.updateAmpImagesURIs(res)

	return res
}

func (productized *Productized) removeAmpServiceAccount(objects []runtime.RawExtension) []runtime.RawExtension {
	res := objects
	for idx, rawExtension := range res {
		obj := rawExtension.Object
		sa, ok := obj.(*v1.ServiceAccount)
		if ok {
			if sa.ObjectMeta.Name == "amp" {
				res = append(res[:idx], res[idx+1:]...) // This deletes the element in the array
				break
			}
		}
	}
	return res
}

func (productized *Productized) removeAmpServiceAccountReferences(objects []runtime.RawExtension) []runtime.RawExtension {
	res := objects
	for _, rawExtension := range res {
		obj := rawExtension.Object
		dc, ok := obj.(*appsv1.DeploymentConfig)
		if ok {
			dc.Spec.Template.Spec.ServiceAccountName = ""
		}
	}

	return res
}

func (productized *Productized) updateAmpImagesParameters(template *templatev1.Template) {
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

func (productized *Productized) updateAmpImagesURIs(objects []runtime.RawExtension) []runtime.RawExtension {
	res := objects

	for _, rawExtension := range res {
		obj := rawExtension.Object
		is, ok := obj.(*imagev1.ImageStream)
		if ok {
			for tagIdx := range is.Spec.Tags {
				// Only change the ImageStream tag name that has the ampRelease
				// value. We do not modify the latest tag
				if is.Spec.Tags[tagIdx].Name == productized.Options.ampRelease {
					switch is.Name {
					case "amp-apicast":
						is.Spec.Tags[tagIdx].From.Name = productized.Options.apicastImage
					case "amp-system":
						is.Spec.Tags[tagIdx].From.Name = productized.Options.systemImage
					case "amp-backend":
						is.Spec.Tags[tagIdx].From.Name = productized.Options.backendImage
					case "amp-wildcard-router":
						is.Spec.Tags[tagIdx].From.Name = productized.Options.routerImage
					case "amp-zync":
						is.Spec.Tags[tagIdx].From.Name = productized.Options.zyncImage
					}
				}
			}
		} else {
			dc, ok := obj.(*appsv1.DeploymentConfig)
			if ok && dc.Name == "system-memcache" {
				dc.Spec.Template.Spec.Containers[0].Image = productized.Options.memcachedImage
			}
		}
	}

	return res
}

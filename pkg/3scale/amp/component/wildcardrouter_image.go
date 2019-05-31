package component

import (
	"fmt"

	imagev1 "github.com/openshift/api/image/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type WildcardRouterImage struct {
	options []string
	Options *WildcardRouterImageOptions
}

type WildcardRouterImageOptions struct {
	appLabel             string
	ampRelease           string
	image                string
	insecureImportPolicy bool
}

func NewWildcardRouterImage(options []string) *WildcardRouterImage {
	wildcardRouterImage := &WildcardRouterImage{
		options: options,
	}
	return wildcardRouterImage
}

type WildcardRouterImageOptionsProvider interface {
	GetWildcardRouterImageOptions() *AmpImagesOptions
}
type CLIWildcardRouterImageOptionsProvider struct {
}

func (o *CLIWildcardRouterImageOptionsProvider) GetWildcardRouterImageOptions() (*WildcardRouterImageOptions, error) {
	aob := WildcardRouterImageOptionsBuilder{}
	aob.AppLabel("${APP_LABEL}")
	aob.AmpRelease("${AMP_RELEASE}")
	aob.Image("${AMP_ROUTER_IMAGE}")
	aob.InsecureImportPolicy(insecureImportPolicy)

	res, err := aob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create WildcardRouterImage Options - %s", err)
	}
	return res, nil
}

func (w *WildcardRouterImage) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLIWildcardRouterImageOptionsProvider{}
	imagesOpts, err := optionsProvider.GetWildcardRouterImageOptions()
	_ = err
	w.Options = imagesOpts
	w.buildParameters(template)
	w.addObjectsIntoTemplate(template)
}

func (w *WildcardRouterImage) GetObjects() ([]runtime.RawExtension, error) {
	objects := w.buildObjects()
	return objects, nil
}

func (w *WildcardRouterImage) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := w.buildObjects()
	template.Objects = append(template.Objects, objects...)
}

func (w *WildcardRouterImage) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (w *WildcardRouterImage) buildObjects() []runtime.RawExtension {
	systemPostgreSQLImageStream := w.buildWildcardRouterImageStream()

	objects := []runtime.RawExtension{
		runtime.RawExtension{Object: systemPostgreSQLImageStream},
	}

	return objects
}

func (w *WildcardRouterImage) buildWildcardRouterImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-wildcard-router",
			Labels: map[string]string{
				"app":                  w.Options.appLabel,
				"threescale_component": "wildcard-router",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "AMP APIcast Wildcard Router",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP APIcast Wildcard Router (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: w.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: w.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP APIcast Wildcard Router " + w.Options.ampRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: w.Options.image,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						// TODO this was originally a double brace expansion from a variable, that is not possible
						// natively with kubernetes so we replaced it with a const
						Insecure: insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (w *WildcardRouterImage) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:     "AMP_ROUTER_IMAGE",
			Value:    "quay.io/3scale/wildcard-router:nightly",
			Required: true,
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}

package component

import (
	"fmt"

	imagev1 "github.com/openshift/api/image/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	insecureImportPolicy = false
)

type AmpImages struct {
	options []string
	Options *AmpImagesOptions
}

type AmpImagesOptions struct {
	appLabel             string
	ampRelease           string
	apicastImage         string
	backendImage         string
	routerImage          string
	systemImage          string
	zyncImage            string
	postgreSQLImage      string
	insecureImportPolicy bool
}

func NewAmpImages(options []string) *AmpImages {
	ampImages := &AmpImages{
		options: options,
	}
	return ampImages
}

type AmpImagesOptionsBuilder struct {
	options AmpImagesOptions
}

func (ampImages *AmpImagesOptionsBuilder) AppLabel(appLabel string) {
	ampImages.options.appLabel = appLabel
}

func (ampImages *AmpImagesOptionsBuilder) AMPRelease(ampRelease string) {
	ampImages.options.ampRelease = ampRelease
}

func (ampImages *AmpImagesOptionsBuilder) ApicastImage(apicastImage string) {
	ampImages.options.apicastImage = apicastImage
}

func (ampImages *AmpImagesOptionsBuilder) BackendImage(backendImage string) {
	ampImages.options.backendImage = backendImage
}

func (ampImages *AmpImagesOptionsBuilder) RouterImage(routerImage string) {
	ampImages.options.routerImage = routerImage
}

func (ampImages *AmpImagesOptionsBuilder) SystemImage(systemImage string) {
	ampImages.options.systemImage = systemImage
}

func (ampImages *AmpImagesOptionsBuilder) ZyncImage(zyncImage string) {
	ampImages.options.zyncImage = zyncImage
}

func (ampImages *AmpImagesOptionsBuilder) PostgreSQLImage(postgreSQLImage string) {
	ampImages.options.postgreSQLImage = postgreSQLImage
}

func (ampImages *AmpImagesOptionsBuilder) InsecureImportPolicy(insecureImportPolicy bool) {
	ampImages.options.insecureImportPolicy = insecureImportPolicy
}

func (ampImages *AmpImagesOptionsBuilder) Build() (*AmpImagesOptions, error) {
	if ampImages.options.appLabel == "" {
		return nil, fmt.Errorf("no AppLabel has been provided")
	}
	if ampImages.options.ampRelease == "" {
		return nil, fmt.Errorf("no AMP release has been provided")
	}
	if ampImages.options.apicastImage == "" {
		return nil, fmt.Errorf("no Apicast image has been provided")
	}
	if ampImages.options.backendImage == "" {
		return nil, fmt.Errorf("no Backend image has been provided")
	}
	if ampImages.options.routerImage == "" {
		return nil, fmt.Errorf("no Router image been provided")
	}
	if ampImages.options.systemImage == "" {
		return nil, fmt.Errorf("no System image has been provided")
	}
	if ampImages.options.zyncImage == "" {
		return nil, fmt.Errorf("no Zync image has been provided")
	}
	if ampImages.options.postgreSQLImage == "" {
		return nil, fmt.Errorf("no PostgreSQL image has been provided")
	}

	return &ampImages.options, nil
}

type AmpImagesOptionsProvider interface {
	GetAmpImagesOptions() *AmpImagesOptions
}
type CLIAmpImagesOptionsProvider struct {
}

func (o *CLIAmpImagesOptionsProvider) GetAmpImagesOptions() (*AmpImagesOptions, error) {
	aob := AmpImagesOptionsBuilder{}
	aob.AppLabel("${APP_LABEL}")
	aob.AMPRelease("${AMP_RELEASE}")
	aob.ApicastImage("${AMP_APICAST_IMAGE}")
	aob.BackendImage("${AMP_BACKEND_IMAGE}")
	aob.RouterImage("${AMP_ROUTER_IMAGE}")
	aob.SystemImage("${AMP_SYSTEM_IMAGE}")
	aob.ZyncImage("${AMP_ZYNC_IMAGE}")
	aob.PostgreSQLImage("${POSTGRESQL_IMAGE}")
	aob.InsecureImportPolicy(false)

	res, err := aob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create AMPImages Options - %s", err)
	}
	return res, nil
}

func (ampImages *AmpImages) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLIAmpImagesOptionsProvider{}
	ampImagesOpts, err := optionsProvider.GetAmpImagesOptions()
	_ = err
	ampImages.Options = ampImagesOpts
	ampImages.buildParameters(template)
	ampImages.addObjectsIntoTemplate(template)
}

func (ampImages *AmpImages) GetObjects() ([]runtime.RawExtension, error) {
	objects := ampImages.buildObjects()
	return objects, nil
}

func (ampImages *AmpImages) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := ampImages.buildObjects()
	template.Objects = append(template.Objects, objects...)
}

func (ampImages *AmpImages) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (ampImages *AmpImages) buildObjects() []runtime.RawExtension {
	backendImageStream := ampImages.buildAmpBackendImageStream()
	zyncImageStream := ampImages.buildAmpZyncImageStream()
	apicastImageStream := ampImages.buildApicastImageStream()
	wildcardRouterImageStream := ampImages.buildWildcardRouterImageStream()
	systemImageStream := ampImages.buildAmpSystemImageStream()
	postgreSQLImageStream := ampImages.buildPostgreSQLImageStream()

	quayServiceAccount := ampImages.buildQuayServiceAccount()

	objects := []runtime.RawExtension{
		runtime.RawExtension{Object: backendImageStream},
		runtime.RawExtension{Object: zyncImageStream},
		runtime.RawExtension{Object: apicastImageStream},
		runtime.RawExtension{Object: wildcardRouterImageStream},
		runtime.RawExtension{Object: systemImageStream},
		runtime.RawExtension{Object: postgreSQLImageStream},
		runtime.RawExtension{Object: quayServiceAccount},
	}
	return objects
}

func (ampImages *AmpImages) buildAmpBackendImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-backend",
			Labels: map[string]string{
				"app":              ampImages.Options.appLabel,
				"3scale.component": "backend",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "AMP backend",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "amp-backend (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "amp-backend " + ampImages.Options.ampRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.backendImage,
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

func (ampImages *AmpImages) buildAmpZyncImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-zync",
			Labels: map[string]string{
				"app":              ampImages.Options.appLabel,
				"3scale.component": "zync",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "AMP Zync",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP Zync (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP Zync " + ampImages.Options.ampRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.zyncImage,
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

func (ampImages *AmpImages) buildApicastImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-apicast",
			Labels: map[string]string{
				"app":              ampImages.Options.appLabel,
				"3scale.component": "apicast",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "AMP APIcast",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP APIcast (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP APIcast " + ampImages.Options.ampRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.apicastImage,
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

func (ampImages *AmpImages) buildWildcardRouterImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-wildcard-router",
			Labels: map[string]string{
				"app":              ampImages.Options.appLabel,
				"3scale.component": "wildcard-router",
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
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP APIcast Wildcard Router " + ampImages.Options.ampRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.routerImage,
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

func (ampImages *AmpImages) buildAmpSystemImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-system",
			Labels: map[string]string{
				"app":              ampImages.Options.appLabel,
				"3scale.component": "system",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "AMP System",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP System (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP system " + ampImages.Options.ampRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.systemImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildPostgreSQLImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ImageStream",
			APIVersion: "image.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "postgresql",
			Labels: map[string]string{"3scale.component": "system", "3scale.component-element": "postgresql", "app": ampImages.Options.appLabel},
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "9.5",
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.postgreSQLImage,
					},
					Reference: false,
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: false,
					}}}}}
}

func (ampImages *AmpImages) buildQuayServiceAccount() *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp",
		},
		ImagePullSecrets: []v1.LocalObjectReference{
			v1.LocalObjectReference{
				Name: "quay-auth"}}}
}

func (ampImages *AmpImages) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:     "AMP_BACKEND_IMAGE",
			Required: true,
			Value:    "quay.io/3scale/apisonator:nightly",
		},
		templatev1.Parameter{
			Name:     "AMP_ZYNC_IMAGE",
			Value:    "quay.io/3scale/zync:nightly",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "AMP_APICAST_IMAGE",
			Value:    "quay.io/3scale/apicast:nightly",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "AMP_ROUTER_IMAGE",
			Value:    "quay.io/3scale/wildcard-router:nightly",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "AMP_SYSTEM_IMAGE",
			Value:    "quay.io/3scale/porta:nightly",
			Required: true,
		},
		templatev1.Parameter{
			Name:        "POSTGRESQL_IMAGE",
			Description: "Postgresql image to use",
			Value:       "registry.access.redhat.com/rhscl/postgresql-95-rhel7:9.5",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "MYSQL_IMAGE",
			Description: "Mysql image to use",
			Value:       "registry.access.redhat.com/rhscl/mysql-57-rhel7:5.7",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "MEMCACHED_IMAGE",
			Description: "Memcached image to use",
			Value:       "registry.access.redhat.com/3scale-amp20/memcached",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "IMAGESTREAM_TAG_IMPORT_INSECURE",
			Description: "Set to true if the server may bypass certificate verification or connect directly over HTTP during image import.",
			Value:       "false",
			Required:    true,
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}

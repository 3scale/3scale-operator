package component

import (
	"fmt"

	imagev1 "github.com/openshift/api/image/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type SystemMySQLImage struct {
	options []string
	Options *SystemMySQLImageOptions
}

type SystemMySQLImageOptions struct {
	appLabel             string
	ampRelease           string
	image                string
	insecureImportPolicy bool
}

func NewSystemMySQLImage(options []string) *SystemMySQLImage {
	systemMySQLImage := &SystemMySQLImage{
		options: options,
	}
	return systemMySQLImage
}

type SystemMySQLImageOptionsProvider interface {
	GetSystemMySQLImageOptions() *AmpImagesOptions
}
type CLISystemMySQLImageOptionsProvider struct {
}

func (o *CLISystemMySQLImageOptionsProvider) GetSystemMySQLImageOptions() (*SystemMySQLImageOptions, error) {
	aob := SystemMySQLImageOptionsBuilder{}
	aob.AppLabel("${APP_LABEL}")
	aob.AmpRelease("${AMP_RELEASE}")
	aob.Image("${SYSTEM_DATABASE_IMAGE}")
	aob.InsecureImportPolicy(insecureImportPolicy)

	res, err := aob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create SystemMySQLImage Options - %s", err)
	}
	return res, nil
}

func (s *SystemMySQLImage) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLISystemMySQLImageOptionsProvider{}
	imagesOpts, err := optionsProvider.GetSystemMySQLImageOptions()
	_ = err
	s.Options = imagesOpts
	s.buildParameters(template)
	s.addObjectsIntoTemplate(template)
}

func (s *SystemMySQLImage) GetObjects() ([]runtime.RawExtension, error) {
	objects := s.buildObjects()
	return objects, nil
}

func (s *SystemMySQLImage) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := s.buildObjects()
	template.Objects = append(template.Objects, objects...)
}

func (s *SystemMySQLImage) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (s *SystemMySQLImage) buildObjects() []runtime.RawExtension {
	systemMySQLImageStream := s.buildSystemMySQLImageStream()

	objects := []runtime.RawExtension{
		runtime.RawExtension{Object: systemMySQLImageStream},
	}

	return objects
}

func (s *SystemMySQLImage) buildSystemMySQLImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-mysql",
			Labels: map[string]string{
				"app":                  s.Options.appLabel,
				"threescale_component": "system",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "System MySQL",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "System MySQL (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: s.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: s.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "System " + s.Options.ampRelease + " MySQL",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: s.Options.image,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (s *SystemMySQLImage) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE_IMAGE",
			Description: "System MySQL image to use",
			Value:       "registry.access.redhat.com/rhscl/mysql-57-rhel7:5.7",
			Required:    true,
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}

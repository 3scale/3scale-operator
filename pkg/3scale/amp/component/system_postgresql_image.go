package component

import (
	"fmt"

	imagev1 "github.com/openshift/api/image/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type SystemPostgreSQLImage struct {
	options []string
	Options *SystemPostgreSQLImageOptions
}

type SystemPostgreSQLImageOptions struct {
	appLabel             string
	ampRelease           string
	image                string
	insecureImportPolicy bool
}

func NewSystemPostgreSQLImage(options []string) *SystemPostgreSQLImage {
	systemPostgreSQLImage := &SystemPostgreSQLImage{
		options: options,
	}
	return systemPostgreSQLImage
}

type SystemPostgreSQLImageOptionsProvider interface {
	GetSystemPostgreSQLImageOptions() *AmpImagesOptions
}
type CLISystemPostgreSQLImageOptionsProvider struct {
}

func (o *CLISystemPostgreSQLImageOptionsProvider) GetSystemPostgreSQLImageOptions() (*SystemPostgreSQLImageOptions, error) {
	aob := SystemPostgreSQLImageOptionsBuilder{}
	aob.AppLabel("${APP_LABEL}")
	aob.AmpRelease("${AMP_RELEASE}")
	aob.Image("${SYSTEM_DATABASE_IMAGE}")
	aob.InsecureImportPolicy(insecureImportPolicy)

	res, err := aob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create SystemPostgreSQLImage Options - %s", err)
	}
	return res, nil
}

func (s *SystemPostgreSQLImage) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLISystemPostgreSQLImageOptionsProvider{}
	imagesOpts, err := optionsProvider.GetSystemPostgreSQLImageOptions()
	_ = err
	s.Options = imagesOpts
	s.buildParameters(template)
	s.addObjectsIntoTemplate(template)
}

func (s *SystemPostgreSQLImage) GetObjects() ([]runtime.RawExtension, error) {
	objects := s.buildObjects()
	return objects, nil
}

func (s *SystemPostgreSQLImage) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := s.buildObjects()
	template.Objects = append(template.Objects, objects...)
}

func (s *SystemPostgreSQLImage) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (s *SystemPostgreSQLImage) buildObjects() []runtime.RawExtension {
	systemPostgreSQLImageStream := s.buildSystemPostgreSQLImageStream()

	objects := []runtime.RawExtension{
		runtime.RawExtension{Object: systemPostgreSQLImageStream},
	}

	return objects
}

func (s *SystemPostgreSQLImage) buildSystemPostgreSQLImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-postgresql",
			Labels: map[string]string{
				"app":                  s.Options.appLabel,
				"threescale_component": "system",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "System database",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "System PostgreSQL (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: s.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: s.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "System " + s.Options.ampRelease + " PostgreSQL",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: s.Options.image,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: s.Options.insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (s *SystemPostgreSQLImage) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE_IMAGE",
			Description: "System PostgreSQL image to use",
			Required:    true,
			Value:       "registry.access.redhat.com/rhscl/postgresql-10-rhel7",
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}

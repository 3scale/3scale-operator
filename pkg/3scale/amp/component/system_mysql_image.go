package component

import (
	"github.com/3scale/3scale-operator/pkg/common"

	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SystemMySQLImage struct {
	Options *SystemMySQLImageOptions
}

func NewSystemMySQLImage(options *SystemMySQLImageOptions) *SystemMySQLImage {
	return &SystemMySQLImage{Options: options}
}

func (s *SystemMySQLImage) Objects() []common.KubernetesObject {
	imageStream := s.ImageStream()

	objects := []common.KubernetesObject{
		imageStream,
	}

	return objects
}

func (s *SystemMySQLImage) ImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-mysql",
			Labels: map[string]string{
				"app":                  s.Options.AppLabel,
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
						Name: s.Options.AmpRelease,
					},
				},
				imagev1.TagReference{
					Name: s.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "System " + s.Options.AmpRelease + " MySQL",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: s.Options.Image,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: *s.Options.InsecureImportPolicy,
					},
				},
			},
		},
	}
}

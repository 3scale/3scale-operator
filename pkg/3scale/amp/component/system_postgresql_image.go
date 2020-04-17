package component

import (
	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SystemPostgreSQLImage struct {
	Options *SystemPostgreSQLImageOptions
}

func NewSystemPostgreSQLImage(options *SystemPostgreSQLImageOptions) *SystemPostgreSQLImage {
	return &SystemPostgreSQLImage{Options: options}
}

func (s *SystemPostgreSQLImage) ImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-postgresql",
			Labels: map[string]string{
				"app":                  s.Options.AppLabel,
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
					Name: s.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "System " + s.Options.AmpRelease + " PostgreSQL",
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

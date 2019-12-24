package component

import (
	"github.com/3scale/3scale-operator/pkg/common"

	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	InsecureImportPolicy = false
)

type AmpImages struct {
	Options *AmpImagesOptions
}

func NewAmpImages(options *AmpImagesOptions) *AmpImages {
	return &AmpImages{Options: options}
}

func (ampImages *AmpImages) Objects() []common.KubernetesObject {
	backendImageStream := ampImages.BackendImageStream()
	zyncImageStream := ampImages.ZyncImageStream()
	apicastImageStream := ampImages.APICastImageStream()
	systemImageStream := ampImages.SystemImageStream()
	zyncDatabasePostgreSQLImageStream := ampImages.ZyncDatabasePostgreSQLImageStream()
	systemMemcachedImageStream := ampImages.SystemMemcachedImageStream()

	deploymentsServiceAccount := ampImages.DeploymentsServiceAccount()

	objects := []common.KubernetesObject{
		backendImageStream,
		zyncImageStream,
		apicastImageStream,
		systemImageStream,
		zyncDatabasePostgreSQLImageStream,
		systemMemcachedImageStream,
		deploymentsServiceAccount,
	}
	return objects
}

func (ampImages *AmpImages) BackendImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-backend",
			Labels: map[string]string{
				"app":                  ampImages.Options.AppLabel,
				"threescale_component": "backend",
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
						Name: ampImages.Options.AmpRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "amp-backend " + ampImages.Options.AmpRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.BackendImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: ampImages.Options.InsecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) ZyncImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-zync",
			Labels: map[string]string{
				"app":                  ampImages.Options.AppLabel,
				"threescale_component": "zync",
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
						Name: ampImages.Options.AmpRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP Zync " + ampImages.Options.AmpRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.ZyncImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: ampImages.Options.InsecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) APICastImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-apicast",
			Labels: map[string]string{
				"app":                  ampImages.Options.AppLabel,
				"threescale_component": "apicast",
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
						Name: ampImages.Options.AmpRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP APIcast " + ampImages.Options.AmpRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.ApicastImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: ampImages.Options.InsecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) SystemImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-system",
			Labels: map[string]string{
				"app":                  ampImages.Options.AppLabel,
				"threescale_component": "system",
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
						Name: ampImages.Options.AmpRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP system " + ampImages.Options.AmpRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.SystemImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: ampImages.Options.InsecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) ZyncDatabasePostgreSQLImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-database-postgresql",
			Labels: map[string]string{
				"app":                  ampImages.Options.AppLabel,
				"threescale_component": "system",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "Zync database PostgreSQL",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "Zync PostgreSQL (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.AmpRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "Zync " + ampImages.Options.AmpRelease + " PostgreSQL",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.ZyncDatabasePostgreSQLImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: ampImages.Options.InsecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) SystemMemcachedImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-memcached",
			Labels: map[string]string{
				"app":                  ampImages.Options.AppLabel,
				"threescale_component": "system",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "System Memcached",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "System Memcached (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.AmpRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "System " + ampImages.Options.AmpRelease + " Memcached",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.SystemMemcachedImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: ampImages.Options.InsecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) DeploymentsServiceAccount() *v1.ServiceAccount {
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
				Name: "threescale-registry-auth"}}}
}

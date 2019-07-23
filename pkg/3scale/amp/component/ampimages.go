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
	backendImageStream := ampImages.buildAmpBackendImageStream()
	zyncImageStream := ampImages.buildAmpZyncImageStream()
	apicastImageStream := ampImages.buildApicastImageStream()
	systemImageStream := ampImages.buildAmpSystemImageStream()
	zyncDatabasePostgreSQL := ampImages.buildZyncDatabasePostgreSQLImageStream()
	backendRedisImageStream := ampImages.buildBackendRedisImageStream()
	systemRedisImageStream := ampImages.buildSystemRedisImageStream()
	systemMemcachedImageStream := ampImages.buildSystemMemcachedImageStream()

	deploymentsServiceAccount := ampImages.buildDeploymentsServiceAccount()

	objects := []common.KubernetesObject{
		backendImageStream,
		zyncImageStream,
		apicastImageStream,
		systemImageStream,
		zyncDatabasePostgreSQL,
		backendRedisImageStream,
		systemRedisImageStream,
		systemMemcachedImageStream,
		deploymentsServiceAccount,
	}
	return objects
}

func (ampImages *AmpImages) buildAmpBackendImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-backend",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
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
						Insecure: ampImages.Options.insecureImportPolicy,
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
				"app":                  ampImages.Options.appLabel,
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
						Insecure: ampImages.Options.insecureImportPolicy,
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
				"app":                  ampImages.Options.appLabel,
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
						Insecure: ampImages.Options.insecureImportPolicy,
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
				"app":                  ampImages.Options.appLabel,
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
						Insecure: ampImages.Options.insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildZyncDatabasePostgreSQLImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-database-postgresql",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
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
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "Zync " + ampImages.Options.ampRelease + " PostgreSQL",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.ZyncDatabasePostgreSQLImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: ampImages.Options.insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildBackendRedisImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-redis",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
				"threescale_component": "backend",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "Backend Redis",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "Backend Redis (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "Backend " + ampImages.Options.ampRelease + " Redis",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.backendRedisImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: ampImages.Options.insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildSystemRedisImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-redis",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
				"threescale_component": "system",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "System Redis",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "System Redis (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "System " + ampImages.Options.ampRelease + " Redis",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.systemRedisImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: ampImages.Options.insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildSystemMemcachedImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-memcached",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
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
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "System " + ampImages.Options.ampRelease + " Memcached",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.systemMemcachedImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: ampImages.Options.insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildDeploymentsServiceAccount() *v1.ServiceAccount {
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

package component

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AmpImages struct {
	Options *AmpImagesOptions
}

func NewAmpImages(options *AmpImagesOptions) *AmpImages {
	return &AmpImages{Options: options}
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
		ImagePullSecrets: ampImages.Options.ImagePullSecrets,
	}
}

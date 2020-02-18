package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type OperatorS3OptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
	Namespace      string
	Client         k8sclient.Client
}

type OperatorHighAvailabilityOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
	Namespace      string
	Client         k8sclient.Client
}

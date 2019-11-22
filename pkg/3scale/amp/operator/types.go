package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO probably this could be in only one type where the different
// OptionsProvider interfaces are implemented. For the moment we
// duplicate the data to allow having possible future differences in
// the required data
type OperatorAmpImagesOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorRedisOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorMysqlOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
	Namespace      string
	Client         k8sclient.Client
}

type OperatorSystemPostgreSQLOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
	Namespace      string
	Client         k8sclient.Client
}

type OperatorSystemMySQLImageOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorSystemPostgreSQLImageOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
	Namespace      string
	Client         k8sclient.Client
}

type OperatorMemcachedOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorSystemOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
	Namespace      string
	Client         k8sclient.Client
}

type OperatorZyncOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
	Namespace      string
	Client         k8sclient.Client
}

type OperatorApicastOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
	Namespace      string
	Client         k8sclient.Client
}

type OperatorHighAvailabilityOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
	Namespace      string
	Client         k8sclient.Client
}

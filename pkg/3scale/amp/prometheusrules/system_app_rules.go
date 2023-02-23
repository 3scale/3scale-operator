package prometheusrules

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func init() {
	PrometheusRuleFactories = append(PrometheusRuleFactories, NewSystemAppPrometheusRuleFactory)
}

type SystemAppPrometheusRuleFactory struct {
}

func NewSystemAppPrometheusRuleFactory() PrometheusRuleFactory {
	return &SystemAppPrometheusRuleFactory{}
}

func (s *SystemAppPrometheusRuleFactory) Type() string {
	return "system-app"
}

func (s *SystemAppPrometheusRuleFactory) PrometheusRule(_ bool, ns string) *monitoringv1.PrometheusRule {
	options, err := systemOptions(ns)
	if err != nil {
		panic(err)
	}
	return component.NewSystem(options).SystemAppPrometheusRules()
}

func systemOptions(ns string) (*component.SystemOptions, error) {
	o := component.NewSystemOptions()

	tmp := "_"

	// Required options for generating PrometheusRules
	o.CommonLabels = commonSystemLabels()
	o.Namespace = ns

	// Required options for passing validation, but not needed for generating the prometheus rules
	// To fix this, more granularity at options level.
	o.WildcardDomain = "_"
	o.ImageTag = "_"
	o.EventHooksURL = "_"
	o.ApicastSystemMasterProxyConfigEndpoint = "_"
	o.MemcachedServers = "_"
	o.AdminEmail = &tmp
	o.AppProviderContainerResourceRequirements = &corev1.ResourceRequirements{}
	o.AppMasterContainerResourceRequirements = &corev1.ResourceRequirements{}
	o.AppDeveloperContainerResourceRequirements = &corev1.ResourceRequirements{}
	o.SphinxContainerResourceRequirements = &corev1.ResourceRequirements{}
	o.SidekiqContainerResourceRequirements = &corev1.ResourceRequirements{}
	o.AdminAccessToken = "_"
	o.AdminPassword = "_"
	o.AdminUsername = "_"
	o.ApicastAccessToken = "_"
	o.ApicastRegistryURL = "_"
	o.MasterAccessToken = "_"
	o.MasterName = "_"
	o.MasterUsername = "_"
	o.MasterPassword = "_"
	o.RecaptchaPublicKey = &tmp
	o.RecaptchaPrivateKey = &tmp
	o.AppSecretKeyBase = "_"
	o.BackendSharedSecret = "_"
	o.TenantName = "_"
	o.AppReplicas = 1
	o.SidekiqReplicas = 1
	o.S3FileStorageOptions = &component.S3FileStorageOptions{ConfigurationSecretName: "_"}
	o.SmtpSecretOptions = component.SystemSMTPSecretOptions{
		Address:           &tmp,
		Authentication:    &tmp,
		Domain:            &tmp,
		OpenSSLVerifyMode: &tmp,
		Password:          &tmp,
		Port:              &tmp,
		Username:          &tmp,
	}
	o.CommonAppLabels = map[string]string{}
	o.AppPodTemplateLabels = map[string]string{}
	o.CommonSidekiqLabels = map[string]string{}
	o.SidekiqPodTemplateLabels = map[string]string{}
	o.ProviderUILabels = map[string]string{}
	o.MasterUILabels = map[string]string{}
	o.DeveloperUILabels = map[string]string{}
	o.SphinxLabels = map[string]string{}
	o.SphinxPodTemplateLabels = map[string]string{}
	o.MemcachedLabels = map[string]string{}
	o.SMTPLabels = map[string]string{}
	o.BackendServiceEndpoint = "_"
	o.BackendServiceEndpoint = "_"
	o.UserSessionTTL = &tmp

	return o, o.Validate()
}

func commonSystemLabels() map[string]string {
	return map[string]string{
		"app":                  appsv1alpha1.Default3scaleAppLabel,
		"threescale_component": "system",
	}
}

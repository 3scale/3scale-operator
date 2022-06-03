package prometheusrules

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	appsv1beta1 "github.com/3scale/3scale-operator/apis/apps/v1beta1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func init() {
	PrometheusRuleFactories = append(PrometheusRuleFactories, NewZyncPrometheusRuleFactory)
}

type ZyncPrometheusRuleFactory struct {
}

func NewZyncPrometheusRuleFactory() PrometheusRuleFactory {
	return &ZyncPrometheusRuleFactory{}
}

func (s *ZyncPrometheusRuleFactory) Type() string {
	return "zync"
}

func (s *ZyncPrometheusRuleFactory) PrometheusRule(ns string) *monitoringv1.PrometheusRule {
	options, err := zyncOptions(ns)
	if err != nil {
		panic(err)
	}
	return component.NewZync(options).ZyncPrometheusRules()
}

func zyncOptions(ns string) (*component.ZyncOptions, error) {
	o := component.NewZyncOptions()

	// Required options for generating PrometheusRules
	o.CommonLabels = commonZyncLabels()
	o.Namespace = ns

	// Required options for passing validation, but not needed for generating the prometheus rules
	// To fix this, more granularity at options level.
	o.CommonZyncLabels = map[string]string{}
	o.CommonZyncQueLabels = map[string]string{}
	o.CommonZyncDatabaseLabels = map[string]string{}
	o.ZyncPodTemplateLabels = map[string]string{}
	o.ZyncQuePodTemplateLabels = map[string]string{}
	o.ZyncDatabasePodTemplateLabels = map[string]string{}

	o.AuthenticationToken = "_"
	o.DatabasePassword = "_"
	o.SecretKeyBase = "_"
	o.ImageTag = "_"
	o.DatabaseImageTag = "_"

	o.ZyncReplicas = 1
	o.ZyncQueReplicas = 1

	o.DatabaseURL = "_"
	o.ZyncQueServiceAccountImagePullSecrets = component.DefaultZyncQueServiceAccountImagePullSecrets()

	return o, o.Validate()
}

func commonZyncLabels() map[string]string {
	return map[string]string{
		"app":                  appsv1beta1.Default3scaleAppLabel,
		"threescale_component": "zync",
	}
}

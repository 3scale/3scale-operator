package operator

import (
	"fmt"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OperatorBackendOptionsProvider struct {
	apimanager     *appsv1alpha1.APIManager
	namespace      string
	client         client.Client
	backendOptions *component.BackendOptions
	secretSource   *helper.SecretSource
}

func NewOperatorBackendOptionsProvider(apimanager *appsv1alpha1.APIManager, namespace string, client client.Client) *OperatorBackendOptionsProvider {
	return &OperatorBackendOptionsProvider{
		apimanager:     apimanager,
		namespace:      namespace,
		client:         client,
		backendOptions: component.NewBackendOptions(),
		secretSource:   helper.NewSecretSource(client, namespace),
	}
}

func (o *OperatorBackendOptionsProvider) GetBackendOptions() (*component.BackendOptions, error) {
	o.backendOptions.TenantName = *o.apimanager.Spec.TenantName
	o.backendOptions.WildcardDomain = o.apimanager.Spec.WildcardDomain
	o.backendOptions.ImageTag = version.ThreescaleVersionMajorMinor()

	err := o.setSecretBasedOptions()
	if err != nil {
		return nil, fmt.Errorf("GetBackendOptions reading secret options: %w", err)
	}

	o.setResourceRequirementsOptions()
	o.setNodeAffinityAndTolerationsOptions()
	o.setReplicas()
	o.setPriorityClassNames()
	o.setTopologySpreadConstraints()
	o.setPodTemplateAnnotations()
	o.setBackendRedisTLSEnabled()
	o.setQueuesRedisTLSEnabled()
	o.setRedisAsyncEnabled()

	o.backendOptions.CommonLabels = o.commonLabels()
	o.backendOptions.CommonListenerLabels = o.commonListenerLabels()
	o.backendOptions.CommonWorkerLabels = o.commonWorkerLabels()
	o.backendOptions.CommonCronLabels = o.commonCronLabels()
	o.backendOptions.ListenerPodTemplateLabels = o.listenerPodTemplateLabels()
	o.backendOptions.WorkerPodTemplateLabels = o.workerPodTemplateLabels()
	o.backendOptions.CronPodTemplateLabels = o.cronPodTemplateLabels()

	o.backendOptions.WorkerMetrics = true
	o.backendOptions.ListenerMetrics = true
	o.backendOptions.Namespace = o.apimanager.Namespace

	err = o.backendOptions.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetBackendOptions validating: %w", err)
	}
	return o.backendOptions, err
}

func (o *OperatorBackendOptionsProvider) setSecretBasedOptions() error {
	cases := []struct {
		field       *string
		secretName  string
		secretField string
		defValue    string
	}{
		{
			&o.backendOptions.SystemBackendUsername,
			component.BackendSecretInternalApiSecretName,
			component.BackendSecretInternalApiUsernameFieldName,
			component.DefaultSystemBackendUsername(),
		},
		{
			&o.backendOptions.SystemBackendPassword,
			component.BackendSecretInternalApiSecretName,
			component.BackendSecretInternalApiPasswordFieldName,
			component.DefaultSystemBackendPassword(),
		},
		{
			&o.backendOptions.ServiceEndpoint,
			component.BackendSecretBackendListenerSecretName,
			component.BackendSecretBackendListenerServiceEndpointFieldName,
			component.DefaultBackendServiceEndpoint(),
		},
		{
			&o.backendOptions.RouteEndpoint,
			component.BackendSecretBackendListenerSecretName,
			component.BackendSecretBackendListenerRouteEndpointFieldName,
			fmt.Sprintf("https://backend-%s.%s", *o.apimanager.Spec.TenantName, o.apimanager.Spec.WildcardDomain),
		},
	}

	for _, option := range cases {
		val, err := o.secretSource.FieldValue(option.secretName, option.secretField, option.defValue)
		if err != nil {
			return err
		}
		*option.field = val
	}

	return nil
}

func (o *OperatorBackendOptionsProvider) setResourceRequirementsOptions() {
	if *o.apimanager.Spec.ResourceRequirementsEnabled {
		if o.apimanager.Spec.Backend.ListenerSpec.Hpa {
			o.backendOptions.ListenerResourceRequirements = component.DefaultHPABackendListenerResourceRequirements()
		} else {
			o.backendOptions.ListenerResourceRequirements = component.DefaultBackendListenerResourceRequirements()
		}
		if o.apimanager.Spec.Backend.WorkerSpec.Hpa {
			o.backendOptions.WorkerResourceRequirements = component.DefaultHPABackendWorkerResourceRequirements()
		} else {
			o.backendOptions.WorkerResourceRequirements = component.DefaultBackendWorkerResourceRequirements()
		}
		o.backendOptions.CronResourceRequirements = component.DefaultCronResourceRequirements()
	} else {
		o.backendOptions.ListenerResourceRequirements = v1.ResourceRequirements{}
		o.backendOptions.WorkerResourceRequirements = v1.ResourceRequirements{}
		o.backendOptions.CronResourceRequirements = v1.ResourceRequirements{}
	}

	// Deployment-level ResourceRequirements CR fields have priority over
	// spec.resourceRequirementsEnabled, overwriting that setting when they are
	// defined
	if o.apimanager.Spec.Backend.ListenerSpec.Resources != nil {
		o.backendOptions.ListenerResourceRequirements = *o.apimanager.Spec.Backend.ListenerSpec.Resources
	}
	if o.apimanager.Spec.Backend.WorkerSpec.Resources != nil {
		o.backendOptions.WorkerResourceRequirements = *o.apimanager.Spec.Backend.WorkerSpec.Resources
	}
	if o.apimanager.Spec.Backend.CronSpec.Resources != nil {
		o.backendOptions.CronResourceRequirements = *o.apimanager.Spec.Backend.CronSpec.Resources
	}
}

func (o *OperatorBackendOptionsProvider) setNodeAffinityAndTolerationsOptions() {
	o.backendOptions.ListenerAffinity = o.apimanager.Spec.Backend.ListenerSpec.Affinity
	o.backendOptions.ListenerTolerations = o.apimanager.Spec.Backend.ListenerSpec.Tolerations
	o.backendOptions.WorkerAffinity = o.apimanager.Spec.Backend.WorkerSpec.Affinity
	o.backendOptions.WorkerTolerations = o.apimanager.Spec.Backend.WorkerSpec.Tolerations
	o.backendOptions.CronAffinity = o.apimanager.Spec.Backend.CronSpec.Affinity
	o.backendOptions.CronTolerations = o.apimanager.Spec.Backend.CronSpec.Tolerations
}

func (o *OperatorBackendOptionsProvider) setReplicas() {
	o.backendOptions.ListenerReplicas = 1
	if o.apimanager.Spec.Backend.ListenerSpec.Replicas != nil {
		if !o.apimanager.Spec.Backend.ListenerSpec.Hpa {
			o.backendOptions.ListenerReplicas = int32(*o.apimanager.Spec.Backend.ListenerSpec.Replicas)
		}
	}

	o.backendOptions.WorkerReplicas = 1
	if o.apimanager.Spec.Backend.WorkerSpec.Replicas != nil {
		if !o.apimanager.Spec.Backend.WorkerSpec.Hpa {
			o.backendOptions.WorkerReplicas = int32(*o.apimanager.Spec.Backend.WorkerSpec.Replicas)
		}
	}

	o.backendOptions.CronReplicas = 1
	if o.apimanager.Spec.Backend.CronSpec.Replicas != nil {
		o.backendOptions.CronReplicas = int32(*o.apimanager.Spec.Backend.CronSpec.Replicas)
	}
}

func (o *OperatorBackendOptionsProvider) commonLabels() map[string]string {
	return map[string]string{
		"app":                  *o.apimanager.Spec.AppLabel,
		"threescale_component": "backend",
	}
}

func (o *OperatorBackendOptionsProvider) commonListenerLabels() map[string]string {
	labels := o.commonLabels()
	labels["threescale_component_element"] = "listener"
	return labels
}

func (o *OperatorBackendOptionsProvider) commonWorkerLabels() map[string]string {
	labels := o.commonLabels()
	labels["threescale_component_element"] = "worker"
	return labels
}

func (o *OperatorBackendOptionsProvider) commonCronLabels() map[string]string {
	labels := o.commonLabels()
	labels["threescale_component_element"] = "cron"
	return labels
}

func (o *OperatorBackendOptionsProvider) listenerPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("backend-listener", helper.ApplicationType)

	for k, v := range o.commonListenerLabels() {
		labels[k] = v
	}

	for k, v := range o.apimanager.Spec.Backend.ListenerSpec.Labels {
		labels[k] = v
	}

	labels[reconcilers.DeploymentLabelSelector] = "backend-listener"

	return labels
}

func (o *OperatorBackendOptionsProvider) workerPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("backend-worker", helper.ApplicationType)

	for k, v := range o.commonWorkerLabels() {
		labels[k] = v
	}

	for k, v := range o.apimanager.Spec.Backend.WorkerSpec.Labels {
		labels[k] = v
	}

	labels[reconcilers.DeploymentLabelSelector] = "backend-worker"

	return labels
}

func (o *OperatorBackendOptionsProvider) cronPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("backend-cron", helper.ApplicationType)

	for k, v := range o.commonCronLabels() {
		labels[k] = v
	}

	for k, v := range o.apimanager.Spec.Backend.CronSpec.Labels {
		labels[k] = v
	}

	labels[reconcilers.DeploymentLabelSelector] = "backend-cron"

	return labels
}

func (o *OperatorBackendOptionsProvider) setPriorityClassNames() {
	if o.apimanager.Spec.Backend.ListenerSpec.PriorityClassName != nil {
		o.backendOptions.PriorityClassNameListener = *o.apimanager.Spec.Backend.ListenerSpec.PriorityClassName
	}
	if o.apimanager.Spec.Backend.WorkerSpec.PriorityClassName != nil {
		o.backendOptions.PriorityClassNameWorker = *o.apimanager.Spec.Backend.WorkerSpec.PriorityClassName
	}
	if o.apimanager.Spec.Backend.CronSpec.PriorityClassName != nil {
		o.backendOptions.PriorityClassNameCron = *o.apimanager.Spec.Backend.CronSpec.PriorityClassName
	}
}

func (o *OperatorBackendOptionsProvider) setPodTemplateAnnotations() {
	o.backendOptions.ListenerPodTemplateAnnotations = o.apimanager.Spec.Backend.ListenerSpec.Annotations
	o.backendOptions.WorkerPodTemplateAnnotations = o.apimanager.Spec.Backend.WorkerSpec.Annotations
	o.backendOptions.CronPodTemplateAnnotations = o.apimanager.Spec.Backend.CronSpec.Annotations
}

func (o *OperatorBackendOptionsProvider) setTopologySpreadConstraints() {
	if o.apimanager.Spec.Backend.ListenerSpec.TopologySpreadConstraints != nil {
		o.backendOptions.TopologySpreadConstraintsListener = o.apimanager.Spec.Backend.ListenerSpec.TopologySpreadConstraints
	}
	if o.apimanager.Spec.Backend.WorkerSpec.TopologySpreadConstraints != nil {
		o.backendOptions.TopologySpreadConstraintsWorker = o.apimanager.Spec.Backend.WorkerSpec.TopologySpreadConstraints
	}
	if o.apimanager.Spec.Backend.CronSpec.TopologySpreadConstraints != nil {
		o.backendOptions.TopologySpreadConstraintsCron = o.apimanager.Spec.Backend.CronSpec.TopologySpreadConstraints
	}
}

func (o *OperatorBackendOptionsProvider) setBackendRedisTLSEnabled() {
	o.backendOptions.BackendRedisTLSEnabled = o.apimanager.IsBackendRedisTLSEnabled()
}

func (o *OperatorBackendOptionsProvider) setQueuesRedisTLSEnabled() {
	o.backendOptions.QueuesRedisTLSEnabled = o.apimanager.IsQueuesRedisTLSEnabled()
}

func (o *OperatorBackendOptionsProvider) setRedisAsyncEnabled() {
	o.backendOptions.RedisAsyncEnabled = !o.apimanager.IsAsyncDisableAnnotationPresent()
}

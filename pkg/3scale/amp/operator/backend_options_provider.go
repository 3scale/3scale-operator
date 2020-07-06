package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"

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
	o.backendOptions.ImageTag = product.ThreescaleRelease

	err := o.setSecretBasedOptions()
	if err != nil {
		return nil, fmt.Errorf("GetBackendOptions reading secret options: %w", err)
	}

	o.setResourceRequirementsOptions()
	o.setNodeAffinityAndTolerationsOptions()
	o.setReplicas()

	imageOpts, err := NewAmpImagesOptionsProvider(o.apimanager).GetAmpImagesOptions()
	if err != nil {
		return nil, fmt.Errorf("GetBackendOptions reading image options: %w", err)
	}
	o.backendOptions.CommonLabels = o.commonLabels()
	o.backendOptions.CommonListenerLabels = o.commonListenerLabels()
	o.backendOptions.CommonWorkerLabels = o.commonWorkerLabels()
	o.backendOptions.CommonCronLabels = o.commonCronLabels()
	o.backendOptions.ListenerPodTemplateLabels = o.listenerPodTemplateLabels(imageOpts.BackendImage)
	o.backendOptions.WorkerPodTemplateLabels = o.workerPodTemplateLabels(imageOpts.BackendImage)
	o.backendOptions.CronPodTemplateLabels = o.cronPodTemplateLabels(imageOpts.BackendImage)

	o.backendOptions.WorkerMetrics = true
	o.backendOptions.ListenerMetrics = true

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
		{
			&o.backendOptions.StorageURL,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageURLFieldName,
			component.DefaultBackendRedisStorageURL(),
		},
		{
			&o.backendOptions.QueuesURL,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesURLFieldName,
			component.DefaultBackendRedisQueuesURL(),
		},
		{
			&o.backendOptions.StorageSentinelHosts,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageSentinelHostsFieldName,
			component.DefaultBackendStorageSentinelHosts(),
		},
		{
			&o.backendOptions.StorageSentinelRole,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageSentinelRoleFieldName,
			component.DefaultBackendStorageSentinelRole(),
		},
		{
			&o.backendOptions.QueuesSentinelHosts,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesSentinelHostsFieldName,
			component.DefaultBackendQueuesSentinelHosts(),
		},
		{
			&o.backendOptions.QueuesSentinelRole,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesSentinelRoleFieldName,
			component.DefaultBackendQueuesSentinelRole(),
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
		o.backendOptions.ListenerResourceRequirements = component.DefaultBackendListenerResourceRequirements()
		o.backendOptions.WorkerResourceRequirements = component.DefaultBackendWorkerResourceRequirements()
		o.backendOptions.CronResourceRequirements = component.DefaultCronResourceRequirements()
	} else {
		o.backendOptions.ListenerResourceRequirements = v1.ResourceRequirements{}
		o.backendOptions.WorkerResourceRequirements = v1.ResourceRequirements{}
		o.backendOptions.CronResourceRequirements = v1.ResourceRequirements{}
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
	o.backendOptions.ListenerReplicas = int32(*o.apimanager.Spec.Backend.ListenerSpec.Replicas)
	o.backendOptions.WorkerReplicas = int32(*o.apimanager.Spec.Backend.WorkerSpec.Replicas)
	o.backendOptions.CronReplicas = int32(*o.apimanager.Spec.Backend.CronSpec.Replicas)
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

func (o *OperatorBackendOptionsProvider) listenerPodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("backend-listener", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range o.commonListenerLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "backend-listener"

	return labels
}

func (o *OperatorBackendOptionsProvider) workerPodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("backend-worker", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range o.commonWorkerLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "backend-worker"

	return labels
}

func (o *OperatorBackendOptionsProvider) cronPodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("backend-cron", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range o.commonCronLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "backend-cron"

	return labels
}

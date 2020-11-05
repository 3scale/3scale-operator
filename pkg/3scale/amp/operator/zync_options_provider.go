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

type ZyncOptionsProvider struct {
	apimanager   *appsv1alpha1.APIManager
	namespace    string
	client       client.Client
	zyncOptions  *component.ZyncOptions
	secretSource *helper.SecretSource
}

func NewZyncOptionsProvider(apimanager *appsv1alpha1.APIManager, namespace string, client client.Client) *ZyncOptionsProvider {
	return &ZyncOptionsProvider{
		apimanager:   apimanager,
		namespace:    namespace,
		client:       client,
		zyncOptions:  component.NewZyncOptions(),
		secretSource: helper.NewSecretSource(client, namespace),
	}
}

func (z *ZyncOptionsProvider) GetZyncOptions() (*component.ZyncOptions, error) {
	z.zyncOptions.ImageTag = product.ThreescaleRelease
	z.zyncOptions.DatabaseImageTag = product.ThreescaleRelease

	err := z.setSecretBasedOptions()
	if err != nil {
		return nil, fmt.Errorf("GetZyncOptions reading secret options: %w", err)
	}

	// must be done after reading from secret
	// database url contains password
	z.zyncOptions.DatabaseURL = component.DefaultZyncDatabaseURL(z.zyncOptions.DatabasePassword)

	z.setResourceRequirementsOptions()
	z.setNodeAffinityAndTolerationsOptions()
	z.setReplicas()

	imageOpts, err := NewAmpImagesOptionsProvider(z.apimanager).GetAmpImagesOptions()
	if err != nil {
		return nil, fmt.Errorf("GetZyncOptions reading image options: %w", err)
	}

	z.zyncOptions.CommonLabels = z.commonLabels()
	z.zyncOptions.CommonZyncLabels = z.commonZyncLabels()
	z.zyncOptions.CommonZyncQueLabels = z.commonZyncQueLabels()
	z.zyncOptions.CommonZyncDatabaseLabels = z.commonZyncDatabaseLabels()
	z.zyncOptions.ZyncPodTemplateLabels = z.zyncPodTemplateLabels(imageOpts.ZyncImage)
	z.zyncOptions.ZyncQuePodTemplateLabels = z.zyncQuePodTemplateLabels(imageOpts.ZyncImage)
	z.zyncOptions.ZyncDatabasePodTemplateLabels = z.zyncDatabasePodTemplateLabels(imageOpts.ZyncDatabasePostgreSQLImage)

	z.zyncOptions.ZyncMetrics = true

	err = z.zyncOptions.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetZyncOptions validating: %w", err)
	}
	return z.zyncOptions, nil
}

func (z *ZyncOptionsProvider) setSecretBasedOptions() error {
	cases := []struct {
		field       *string
		secretName  string
		secretField string
		defValue    string
	}{
		{
			&z.zyncOptions.SecretKeyBase,
			component.ZyncSecretName,
			component.ZyncSecretKeyBaseFieldName,
			component.DefaultZyncSecretKeyBase(),
		},
		{
			&z.zyncOptions.DatabasePassword,
			component.ZyncSecretName,
			component.ZyncSecretDatabasePasswordFieldName,
			component.DefaultZyncDatabasePassword(),
		},
		{
			&z.zyncOptions.AuthenticationToken,
			component.ZyncSecretName,
			component.ZyncSecretAuthenticationTokenFieldName,
			component.DefaultZyncAuthenticationToken(),
		},
	}

	for _, option := range cases {
		val, err := z.secretSource.FieldValue(option.secretName, option.secretField, option.defValue)
		if err != nil {
			return err
		}
		*option.field = val
	}

	return nil
}

func (z *ZyncOptionsProvider) setResourceRequirementsOptions() {
	if *z.apimanager.Spec.ResourceRequirementsEnabled {
		z.zyncOptions.ContainerResourceRequirements = component.DefaultZyncContainerResourceRequirements()
		z.zyncOptions.QueContainerResourceRequirements = component.DefaultZyncQueContainerResourceRequirements()
		z.zyncOptions.DatabaseContainerResourceRequirements = component.DefaultZyncDatabaseContainerResourceRequirements()
	} else {
		z.zyncOptions.ContainerResourceRequirements = v1.ResourceRequirements{}
		z.zyncOptions.QueContainerResourceRequirements = v1.ResourceRequirements{}
		z.zyncOptions.DatabaseContainerResourceRequirements = v1.ResourceRequirements{}
	}
}

func (z *ZyncOptionsProvider) setNodeAffinityAndTolerationsOptions() {
	z.zyncOptions.ZyncAffinity = z.apimanager.Spec.Zync.AppSpec.Affinity
	z.zyncOptions.ZyncTolerations = z.apimanager.Spec.Zync.AppSpec.Tolerations
	z.zyncOptions.ZyncQueAffinity = z.apimanager.Spec.Zync.QueSpec.Affinity
	z.zyncOptions.ZyncQueTolerations = z.apimanager.Spec.Zync.QueSpec.Tolerations
	z.zyncOptions.ZyncDatabaseAffinity = z.apimanager.Spec.Zync.DatabaseAffinity
	z.zyncOptions.ZyncDatabaseTolerations = z.apimanager.Spec.Zync.DatabaseTolerations
}

func (z *ZyncOptionsProvider) setReplicas() {
	z.zyncOptions.ZyncReplicas = int32(*z.apimanager.Spec.Zync.AppSpec.Replicas)
	z.zyncOptions.ZyncQueReplicas = int32(*z.apimanager.Spec.Zync.QueSpec.Replicas)
}

func (z *ZyncOptionsProvider) commonLabels() map[string]string {
	return map[string]string{
		"app":                  *z.apimanager.Spec.AppLabel,
		"threescale_component": "zync",
	}
}

func (z *ZyncOptionsProvider) commonZyncLabels() map[string]string {
	labels := z.commonLabels()
	labels["threescale_component_element"] = "zync"
	return labels
}

func (z *ZyncOptionsProvider) commonZyncQueLabels() map[string]string {
	labels := z.commonLabels()
	labels["threescale_component_element"] = "zync-que"
	return labels
}

func (z *ZyncOptionsProvider) commonZyncDatabaseLabels() map[string]string {
	labels := z.commonLabels()
	labels["threescale_component_element"] = "database"
	return labels
}

func (z *ZyncOptionsProvider) zyncPodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("zync", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range z.commonZyncLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "zync"

	return labels
}

func (z *ZyncOptionsProvider) zyncQuePodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("zync-que", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range z.commonZyncQueLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "zync-que"

	return labels
}

func (z *ZyncOptionsProvider) zyncDatabasePodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("zync-database", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range z.commonZyncDatabaseLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "zync-database"

	return labels
}

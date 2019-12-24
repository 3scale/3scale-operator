package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
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
	z.zyncOptions.AppLabel = *z.apimanager.Spec.AppLabel

	err := z.setSecretBasedOptions()
	if err != nil {
		return nil, err
	}

	// must be done after reading from secret
	// database url contains password
	z.zyncOptions.DatabaseURL = component.DefaultZyncDatabaseURL(z.zyncOptions.DatabasePassword)

	z.setResourceRequirementsOptions()
	z.setReplicas()

	err = z.zyncOptions.Validate()
	return z.zyncOptions, err
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
		// not nil value is ensured
		*option.field = *val
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

func (z *ZyncOptionsProvider) setReplicas() {
	z.zyncOptions.ZyncReplicas = int32(*z.apimanager.Spec.Zync.AppSpec.Replicas)
	z.zyncOptions.ZyncQueReplicas = int32(*z.apimanager.Spec.Zync.QueSpec.Replicas)
}

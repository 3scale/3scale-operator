package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type Apicast struct {
	generatePodDisruptionBudget bool
}

func NewApicastAdapter(generatePDB bool) Adapter {
	return NewAppenderAdapter(&Apicast{generatePodDisruptionBudget: generatePDB})
}

func (a *Apicast) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "APICAST_ACCESS_TOKEN",
			Description: "Read Only Access Token that is APIcast going to use to download its configuration.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "APICAST_MANAGEMENT_API",
			Description: "Scope of the APIcast Management API. Can be disabled, status or debug. At least status required for health checks.",
			Value:       "status",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "APICAST_OPENSSL_VERIFY",
			Description: "Turn on/off the OpenSSL peer verification when downloading the configuration. Can be set to true/false.",
			Value:       "false",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "APICAST_RESPONSE_CODES",
			Description: "Enable logging response codes in APIcast.",
			Value:       "true",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "APICAST_REGISTRY_URL",
			Description: "The URL to point to APIcast policies registry management",
			Value:       "http://apicast-staging:8090/policies",
			Required:    true,
		},
	}
}

func (a *Apicast) Objects() ([]common.KubernetesObject, error) {
	apicastOptions, err := a.options()
	if err != nil {
		return nil, err
	}
	apicastComponent := component.NewApicast(apicastOptions)
	objects := a.componentObjects(apicastComponent)

	return objects, nil
}

func (a *Apicast) componentObjects(c *component.Apicast) []common.KubernetesObject {
	stagingDeploymentConfig := c.StagingDeploymentConfig()
	productionDeploymentConfig := c.ProductionDeploymentConfig()
	stagingService := c.StagingService()
	productionService := c.ProductionService()
	environmentConfigMap := c.EnvironmentConfigMap()

	objects := []common.KubernetesObject{
		stagingDeploymentConfig,
		productionDeploymentConfig,
		stagingService,
		productionService,
		environmentConfigMap,
	}

	if a.generatePodDisruptionBudget {
		objects = append(objects, a.componentPDBObjects(c)...)
	}

	return objects
}

func (a *Apicast) componentPDBObjects(c *component.Apicast) []common.KubernetesObject {
	stagingPDB := c.StagingPodDisruptionBudget()
	prodPDB := c.ProductionPodDisruptionBudget()
	return []common.KubernetesObject{
		stagingPDB,
		prodPDB,
	}
}

func (a *Apicast) options() (*component.ApicastOptions, error) {
	ao := component.NewApicastOptions()
	ao.ManagementAPI = "${APICAST_MANAGEMENT_API}"
	ao.OpenSSLVerify = "${APICAST_OPENSSL_VERIFY}"
	ao.ResponseCodes = "${APICAST_RESPONSE_CODES}"
	ao.ImageTag = "${AMP_RELEASE}"

	ao.ProductionResourceRequirements = component.DefaultProductionResourceRequirements()
	ao.StagingResourceRequirements = component.DefaultStagingResourceRequirements()

	ao.ProductionReplicas = 1
	ao.StagingReplicas = 1

	ao.CommonLabels = a.commonLabels()
	ao.CommonStagingLabels = a.commonStagingLabels()
	ao.CommonProductionLabels = a.commonProductionLabels()
	ao.StagingPodTemplateLabels = a.stagingPodTemplateLabels()
	ao.ProductionPodTemplateLabels = a.productionPodTemplateLabels()

	err := ao.Validate()
	return ao, err
}

func (a *Apicast) commonLabels() map[string]string {
	return map[string]string{
		"app":                  "${APP_LABEL}",
		"threescale_component": "apicast",
	}
}

func (a *Apicast) commonStagingLabels() map[string]string {
	labels := a.commonLabels()
	labels["threescale_component_element"] = "staging"
	return labels
}

func (a *Apicast) commonProductionLabels() map[string]string {
	labels := a.commonLabels()
	labels["threescale_component_element"] = "production"
	return labels
}

func (a *Apicast) stagingPodTemplateLabels() map[string]string {
	labels := a.commonStagingLabels()
	labels["deploymentConfig"] = "apicast-staging"
	return labels
}

func (a *Apicast) productionPodTemplateLabels() map[string]string {
	labels := a.commonProductionLabels()
	labels["deploymentConfig"] = "apicast-production"
	return labels
}

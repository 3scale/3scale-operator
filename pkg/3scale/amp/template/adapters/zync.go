package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type Zync struct {
	generatePodDisruptionBudget bool
}

func NewZyncAdapter(generatePDB bool) Adapter {
	return NewAppenderAdapter(&Zync{generatePodDisruptionBudget: generatePDB})
}

func (z *Zync) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "ZYNC_DATABASE_PASSWORD",
			DisplayName: "Zync Database PostgreSQL Connection Password",
			Description: "Password for the Zync Database PostgreSQL connection user.",
			Generate:    "expression",
			From:        "[a-zA-Z0-9]{16}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:     "ZYNC_SECRET_KEY_BASE",
			Generate: "expression",
			From:     "[a-zA-Z0-9]{16}",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "ZYNC_AUTHENTICATION_TOKEN",
			Generate: "expression",
			From:     "[a-zA-Z0-9]{16}",
			Required: true,
		},
	}
}

func (z *Zync) Objects() ([]common.KubernetesObject, error) {
	zyncOptions, err := z.options()
	if err != nil {
		return nil, err
	}
	zyncComponent := component.NewZync(zyncOptions)
	objects := z.componentObjects(zyncComponent)

	return objects, nil
}

func (z *Zync) componentObjects(c *component.Zync) []common.KubernetesObject {
	queRole := c.QueRole()
	queServiceAccount := c.QueServiceAccount()
	queRoleBinding := c.QueRoleBinding()
	deploymentConfig := c.DeploymentConfig()
	queDeploymentConfig := c.QueDeploymentConfig()
	databaseDeploymentConfig := c.DatabaseDeploymentConfig()
	service := c.Service()
	databaseService := c.DatabaseService()
	secret := c.Secret()

	objects := []common.KubernetesObject{
		queRole,
		queServiceAccount,
		queRoleBinding,
		deploymentConfig,
		queDeploymentConfig,
		databaseDeploymentConfig,
		service,
		databaseService,
		secret,
	}

	if z.generatePodDisruptionBudget {
		objects = append(objects, z.componentPDBObjects(c)...)
	}

	return objects
}

func (z *Zync) componentPDBObjects(c *component.Zync) []common.KubernetesObject {
	zyncPDB := c.ZyncPodDisruptionBudget()
	quePDB := c.QuePodDisruptionBudget()

	return []common.KubernetesObject{
		zyncPDB,
		quePDB,
	}
}

func (z *Zync) options() (*component.ZyncOptions, error) {
	zo := component.NewZyncOptions()

	zo.CommonLabels = z.commonLabels()
	zo.CommonZyncLabels = z.commonZyncLabels()
	zo.CommonZyncQueLabels = z.commonZyncQueLabels()
	zo.CommonZyncDatabaseLabels = z.commonZyncDatabaseLabels()
	zo.ZyncPodTemplateLabels = z.zyncPodTemplateLabels()
	zo.ZyncQuePodTemplateLabels = z.zyncQuePodTemplateLabels()
	zo.ZyncDatabasePodTemplateLabels = z.zyncDatabasePodTemplateLabels()

	zo.AuthenticationToken = "${ZYNC_AUTHENTICATION_TOKEN}"
	zo.DatabasePassword = "${ZYNC_DATABASE_PASSWORD}"
	zo.SecretKeyBase = "${ZYNC_SECRET_KEY_BASE}"
	zo.ImageTag = "${AMP_RELEASE}"
	zo.DatabaseImageTag = "${AMP_RELEASE}"

	zo.ZyncReplicas = 1
	zo.ZyncQueReplicas = 1

	zo.ContainerResourceRequirements = component.DefaultZyncContainerResourceRequirements()
	zo.QueContainerResourceRequirements = component.DefaultZyncQueContainerResourceRequirements()
	zo.DatabaseContainerResourceRequirements = component.DefaultZyncDatabaseContainerResourceRequirements()

	zo.DatabaseURL = component.DefaultZyncDatabaseURL(zo.DatabasePassword)

	err := zo.Validate()
	return zo, err
}

func (z *Zync) commonLabels() map[string]string {
	return map[string]string{
		"app":                  "${APP_LABEL}",
		"threescale_component": "zync",
	}
}

func (z *Zync) commonZyncLabels() map[string]string {
	return z.commonLabels()
}

func (z *Zync) commonZyncQueLabels() map[string]string {
	labels := z.commonLabels()
	return labels
}

func (z *Zync) commonZyncDatabaseLabels() map[string]string {
	labels := z.commonLabels()
	labels["threescale_component_element"] = "database"
	return labels
}

func (z *Zync) zyncPodTemplateLabels() map[string]string {
	labels := z.commonZyncLabels()
	labels["deploymentConfig"] = "zync"
	return labels
}

func (z *Zync) zyncQuePodTemplateLabels() map[string]string {
	return map[string]string{
		"app":              "${APP_LABEL}",
		"deploymentConfig": "zync-que",
	}
}

func (z *Zync) zyncDatabasePodTemplateLabels() map[string]string {
	labels := z.commonZyncDatabaseLabels()
	labels["deploymentConfig"] = "zync-database"
	return labels
}

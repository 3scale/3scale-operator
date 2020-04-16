package adapters

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type Backend struct {
	generatePodDisruptionBudget bool
}

func NewBackendAdapter(generatePDB bool) Adapter {
	return NewAppenderAdapter(&Backend{
		generatePodDisruptionBudget: generatePDB,
	})
}

func (b *Backend) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{}
}

func (b *Backend) Objects() ([]common.KubernetesObject, error) {
	backendOptions, err := b.options()
	if err != nil {
		return nil, err
	}
	backendComponent := component.NewBackend(backendOptions)
	objects := b.componentObjects(backendComponent)

	return objects, nil
}

func (b *Backend) componentObjects(c *component.Backend) []common.KubernetesObject {
	cronDeploymentConfig := c.CronDeploymentConfig()
	listenerDeploymentConfig := c.ListenerDeploymentConfig()
	listenerService := c.ListenerService()
	listenerRoute := c.ListenerRoute()
	workerDeploymentConfig := c.WorkerDeploymentConfig()
	environmentConfigMap := c.EnvironmentConfigMap()

	internalAPISecretForSystem := c.InternalAPISecretForSystem()
	redisSecret := c.RedisSecret()
	listenerSecret := c.ListenerSecret()

	objects := []common.KubernetesObject{
		cronDeploymentConfig,
		listenerDeploymentConfig,
		listenerService,
		listenerRoute,
		workerDeploymentConfig,
		environmentConfigMap,
		internalAPISecretForSystem,
		redisSecret,
		listenerSecret,
	}

	if b.generatePodDisruptionBudget {
		objects = append(objects, b.componentPDBObjects(c)...)
	}

	return objects
}

func (b *Backend) componentPDBObjects(c *component.Backend) []common.KubernetesObject {
	workerPDB := c.WorkerPodDisruptionBudget()
	cronPDB := c.CronPodDisruptionBudget()
	listenerPDB := c.ListenerPodDisruptionBudget()
	return []common.KubernetesObject{
		workerPDB,
		cronPDB,
		listenerPDB,
	}
}

func (b *Backend) options() (*component.BackendOptions, error) {
	bo := component.NewBackendOptions()
	bo.AppLabel = "${APP_LABEL}"
	bo.SystemBackendUsername = "${SYSTEM_BACKEND_USERNAME}"
	bo.SystemBackendPassword = "${SYSTEM_BACKEND_PASSWORD}"
	bo.TenantName = "${TENANT_NAME}"
	bo.WildcardDomain = "${WILDCARD_DOMAIN}"
	bo.ImageTag = "${AMP_RELEASE}"
	bo.RouteEndpoint = fmt.Sprintf("https://backend-%s.%s", "${TENANT_NAME}", "${WILDCARD_DOMAIN}")
	bo.ServiceEndpoint = component.DefaultBackendServiceEndpoint()
	bo.StorageURL = component.DefaultBackendRedisStorageURL()
	bo.QueuesURL = component.DefaultBackendRedisQueuesURL()
	bo.ListenerReplicas = 1
	bo.WorkerReplicas = 1
	bo.CronReplicas = 1
	bo.ListenerResourceRequirements = component.DefaultBackendListenerResourceRequirements()
	bo.WorkerResourceRequirements = component.DefaultBackendWorkerResourceRequirements()
	bo.CronResourceRequirements = component.DefaultCronResourceRequirements()
	err := bo.Validate()
	return bo, err
}

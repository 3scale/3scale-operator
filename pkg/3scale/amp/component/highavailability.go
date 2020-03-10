package component

import (
	"github.com/3scale/3scale-operator/pkg/common"

	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HighAvailability struct {
	Options *HighAvailabilityOptions
}

const (
	HighlyAvailableReplicas = 2
)

var highlyAvailableExternalDatabases = map[string]bool{
	"backend-redis": true,
	"system-redis":  true,
	"system-mysql":  true,
}

func NewHighAvailability(options *HighAvailabilityOptions) *HighAvailability {
	return &HighAvailability{Options: options}
}

func (ha *HighAvailability) Objects() []common.KubernetesObject {
	systemDatabaseSecret := ha.SystemDatabaseSecret()

	objects := []common.KubernetesObject{
		systemDatabaseSecret,
	}
	return objects
}

func (ha *HighAvailability) SystemDatabaseSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemSecretSystemDatabaseSecretName,
			Labels: map[string]string{
				"app":                  ha.Options.AppLabel,
				"threescale_component": "system",
			},
		},
		StringData: map[string]string{
			SystemSecretSystemDatabaseURLFieldName: ha.Options.SystemDatabaseURL,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (ha *HighAvailability) IncreaseReplicasNumber(objects []common.KubernetesObject) {
	// We do not increase the number of replicas in database DeploymentConfigs
	excludedDeploymentConfigs := map[string]bool{
		"system-memcache": true,
		"system-sphinx":   true,
		"zync-database":   true,
	}

	for _, obj := range objects {
		dc, ok := obj.(*appsv1.DeploymentConfig)
		if ok {
			if _, isExcluded := excludedDeploymentConfigs[dc.Name]; !isExcluded {
				dc.Spec.Replicas = HighlyAvailableReplicas
			}
		}
	}
}

func (ha *HighAvailability) DeleteInternalDatabasesObjects(objects []common.KubernetesObject) []common.KubernetesObject {
	keepObjects := []common.KubernetesObject{}

	for objIdx, object := range objects {
		switch obj := (object).(type) {
		case *appsv1.DeploymentConfig:
			if _, ok := highlyAvailableExternalDatabases[obj.ObjectMeta.Name]; !ok {
				// We create a new array and add to it the elements that will
				//NOT have to be deleted
				keepObjects = append(keepObjects, objects[objIdx])
			}
		case *v1.Service:
			if _, ok := highlyAvailableExternalDatabases[obj.ObjectMeta.Name]; !ok {
				keepObjects = append(keepObjects, objects[objIdx])
			}
		case *v1.PersistentVolumeClaim:
			if obj.ObjectMeta.Name != "backend-redis-storage" && obj.ObjectMeta.Name != "system-redis-storage" &&
				obj.ObjectMeta.Name != "mysql-storage" {
				keepObjects = append(keepObjects, objects[objIdx])
			}
		case *v1.ConfigMap:
			if obj.ObjectMeta.Name != "mysql-main-conf" && obj.ObjectMeta.Name != "mysql-extra-conf" &&
				obj.ObjectMeta.Name != "redis-config" {
				keepObjects = append(keepObjects, objects[objIdx])
			}
		case *imagev1.ImageStream:
			if _, ok := highlyAvailableExternalDatabases[obj.ObjectMeta.Name]; !ok {
				keepObjects = append(keepObjects, objects[objIdx])
			}
		default:
			keepObjects = append(keepObjects, objects[objIdx])
		}
	}

	return keepObjects
}

func (ha *HighAvailability) UnsetSystemRedisDBDefaultValues(template *templatev1.Template) {
	dbParamsToUpdate := map[string]bool{
		"SYSTEM_REDIS_URL":             true,
		"SYSTEM_MESSAGE_BUS_REDIS_URL": true,
	}

	for paramIdx := range template.Parameters {
		paramName := template.Parameters[paramIdx].Name
		if _, ok := dbParamsToUpdate[paramName]; ok {
			template.Parameters[paramIdx].Value = ""
		}
	}
}

func (ha *HighAvailability) DeleteDBRelatedParameters(template *templatev1.Template) {
	keepParams := []templatev1.Parameter{}
	dbParamsToDelete := map[string]bool{
		"REDIS_IMAGE":                   true,
		"SYSTEM_DATABASE_IMAGE":         true,
		"SYSTEM_DATABASE_USER":          true,
		"SYSTEM_DATABASE_PASSWORD":      true,
		"SYSTEM_DATABASE":               true,
		"SYSTEM_DATABASE_ROOT_PASSWORD": true,
	}

	for paramIdx := range template.Parameters {
		paramName := template.Parameters[paramIdx].Name
		if _, ok := dbParamsToDelete[paramName]; !ok {
			keepParams = append(keepParams, template.Parameters[paramIdx])
		}
	}

	template.Parameters = keepParams
}

func (ha *HighAvailability) UpdateDatabasesURLS(objects []common.KubernetesObject) {
	for objIdx := range objects {
		obj := objects[objIdx]
		secret, ok := obj.(*v1.Secret)
		if ok {
			switch secret.Name {
			case "system-redis":
				secret.StringData["URL"] = ha.Options.SystemRedisURL
				secret.StringData["MESSAGE_BUS_URL"] = ha.Options.SystemMessageBusRedisURL
				secret.StringData[SystemSecretSystemRedisSentinelHosts] = ha.Options.SystemRedisSentinelsHosts
				secret.StringData[SystemSecretSystemRedisSentinelRole] = ha.Options.SystemRedisSentinelsRole
				secret.StringData[SystemSecretSystemRedisMessageBusSentinelHosts] = ha.Options.SystemMessageBusRedisSentinelsHosts
				secret.StringData[SystemSecretSystemRedisMessageBusSentinelRole] = ha.Options.SystemMessageBusRedisSentinelsRole
			case "backend-redis":
				secret.StringData["REDIS_STORAGE_URL"] = ha.Options.BackendRedisStorageEndpoint
				secret.StringData["REDIS_QUEUES_URL"] = ha.Options.BackendRedisQueuesEndpoint
				secret.StringData[BackendSecretBackendRedisStorageSentinelHostsFieldName] = ha.Options.BackendRedisStorageSentinelHosts
				secret.StringData[BackendSecretBackendRedisStorageSentinelRoleFieldName] = ha.Options.BackendRedisStorageSentinelRole
				secret.StringData[BackendSecretBackendRedisQueuesSentinelHostsFieldName] = ha.Options.BackendRedisQueuesSentinelHosts
				secret.StringData[BackendSecretBackendRedisQueuesSentinelRoleFieldName] = ha.Options.BackendRedisQueuesSentinelRole
			}
		}
	}
}

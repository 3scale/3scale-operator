package component

import (
	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type HighAvailability struct {
	options []string
}

type HighAvailabilityOptions struct {
	apicastProductionRedisURL   string
	apicastStagingRedisURL      string
	backendRedisQueuesEndpoint  string
	backendRedisStorageEndpoint string
	systemDatabaseURL           string
	systemRedisURL              string
}

const (
	HighlyAvailableReplicas = 2
)

var highlyAvailableExternalDatabases = map[string]bool{
	"backend-redis": true,
	"system-redis":  true,
	"system-mysql":  true,
}

func NewHighAvailability(options []string) *HighAvailability {
	ha := &HighAvailability{
		options: options,
	}
	return ha
}

func (ha *HighAvailability) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	ha.buildParameters(template)
}

func (ha *HighAvailability) PostProcess(template *templatev1.Template, otherComponents []Component) {
	ha.increaseReplicasNumber(template)
	ha.deleteInternalDatabasesObjects(template)
	ha.deleteDBRelatedParameters(template)
	ha.updateDatabasesURLS(template)
}

func (ha *HighAvailability) increaseReplicasNumber(template *templatev1.Template) {
	// We do not increase the number of replicas in database DeploymentConfigs
	excludedDeploymentConfigs := map[string]bool{
		"system-memcache": true,
		"system-sphinx":   true,
		"zync-database":   true,
	}

	for _, rawExtension := range template.Objects {
		obj := rawExtension.Object
		dc, ok := obj.(*appsv1.DeploymentConfig)
		if ok {
			if _, isExcluded := excludedDeploymentConfigs[dc.Name]; !isExcluded {
				dc.Spec.Replicas = HighlyAvailableReplicas
			}
		}
	}
}

func (ha *HighAvailability) deleteInternalDatabasesObjects(template *templatev1.Template) {
	keepObjects := []runtime.RawExtension{}

	for rawExtIdx, rawExtension := range template.Objects {
		switch obj := (rawExtension.Object).(type) {
		case *appsv1.DeploymentConfig:
			if _, ok := highlyAvailableExternalDatabases[obj.ObjectMeta.Name]; !ok {
				// We create a new array and add to it the elements that will
				//NOT have to be deleted
				keepObjects = append(keepObjects, template.Objects[rawExtIdx])
			}
		case *v1.Service:
			if _, ok := highlyAvailableExternalDatabases[obj.ObjectMeta.Name]; !ok {
				keepObjects = append(keepObjects, template.Objects[rawExtIdx])
			}
		case *v1.PersistentVolumeClaim:
			if obj.ObjectMeta.Name != "backend-redis-storage" && obj.ObjectMeta.Name != "system-redis-storage" &&
				obj.ObjectMeta.Name != "mysql-storage" {
				keepObjects = append(keepObjects, template.Objects[rawExtIdx])
			}
		case *v1.ConfigMap:
			if obj.ObjectMeta.Name != "mysql-main-conf" && obj.ObjectMeta.Name != "mysql-extra-conf" &&
				obj.ObjectMeta.Name != "redis-config" {
				keepObjects = append(keepObjects, template.Objects[rawExtIdx])
			}
		default:
			keepObjects = append(keepObjects, template.Objects[rawExtIdx])
		}
	}

	template.Objects = keepObjects
}

func (ha *HighAvailability) deleteDBRelatedParameters(template *templatev1.Template) {
	keepParams := []templatev1.Parameter{}
	dbParamsToDelete := map[string]bool{
		"MYSQL_IMAGE":         true,
		"REDIS_IMAGE":         true,
		"MYSQL_USER":          true,
		"MYSQL_PASSWORD":      true,
		"MYSQL_DATABASE":      true,
		"MYSQL_ROOT_PASSWORD": true,
	}

	for paramIdx := range template.Parameters {
		paramName := template.Parameters[paramIdx].Name
		if _, ok := dbParamsToDelete[paramName]; !ok {
			keepParams = append(keepParams, template.Parameters[paramIdx])
		}
	}

	template.Parameters = keepParams
}

func (ha *HighAvailability) updateDatabasesURLS(template *templatev1.Template) {
	for rawExtIdx := range template.Objects {
		obj := template.Objects[rawExtIdx].Object
		secret, ok := obj.(*v1.Secret)
		if ok {
			switch secret.Name {
			case "system-redis":
				secret.StringData["URL"] = "${SYSTEM_REDIS_URL}"
			case "system-database":
				secret.StringData["URL"] = "${SYSTEM_DATABASE_URL}"
			case "apicast-redis":
				secret.StringData["PRODUCTION_URL"] = "${APICAST_PRODUCTION_REDIS_URL}"
				secret.StringData["STAGING_URL"] = "${APICAST_STAGING_REDIS_URL}"
			case "backend-redis":
				secret.StringData["REDIS_STORAGE_URL"] = "${BACKEND_REDIS_STORAGE_ENDPOINT}"
				secret.StringData["REDIS_QUEUES_URL"] = "${BACKEND_REDIS_QUEUES_ENDPOINT}"
			}
		}
	}
}

func (ha *HighAvailability) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "BACKEND_REDIS_STORAGE_ENDPOINT",
			Description: "Define the external backend-redis storage endpoint to connect to",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "BACKEND_REDIS_QUEUES_ENDPOINT",
			Description: "Define the external backend-redis queues endpoint to connect to",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_REDIS_URL",
			Description: "Define the external system-redis to connect to",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "APICAST_STAGING_REDIS_URL",
			Description: "Define the external apicast-staging redis to connect to",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "APICAST_PRODUCTION_REDIS_URL",
			Description: "Define the external apicast-staging redis to connect to",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE_URL",
			Description: "Define the external system-mysql to connect to",
			Required:    true,
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}

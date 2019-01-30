package component

import (
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type HighAvailability struct {
	options []string
	Options *HighAvailabilityOptions
}

type HighAvailabilityOptions struct {
	nonRequiredHighAvailabilityOptions
	requiredHighAvailabilityOptions
}

type requiredHighAvailabilityOptions struct {
	apicastProductionRedisURL   string
	apicastStagingRedisURL      string
	backendRedisQueuesEndpoint  string
	backendRedisStorageEndpoint string
	systemDatabaseURL           string
	systemRedisURL              string
}

type nonRequiredHighAvailabilityOptions struct {
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

type HighAvailabilityOptionsBuilder struct {
	options HighAvailabilityOptions
}

func (ha *HighAvailabilityOptionsBuilder) ApicastProductionRedisURL(apicastProductionRedisURL string) {
	ha.options.apicastProductionRedisURL = apicastProductionRedisURL
}

func (ha *HighAvailabilityOptionsBuilder) ApicastStagingRedisURL(apicastStagingRedisURL string) {
	ha.options.apicastStagingRedisURL = apicastStagingRedisURL
}

func (ha *HighAvailabilityOptionsBuilder) BackendRedisQueuesEndpoint(backendRedisQueuesEndpoint string) {
	ha.options.backendRedisQueuesEndpoint = backendRedisQueuesEndpoint
}

func (ha *HighAvailabilityOptionsBuilder) BackendRedisStorageEndpoint(backendRedisStorageEndpoint string) {
	ha.options.backendRedisStorageEndpoint = backendRedisStorageEndpoint
}

func (ha *HighAvailabilityOptionsBuilder) SystemDatabaseURL(systemDatabaseURL string) {
	ha.options.systemDatabaseURL = systemDatabaseURL
}

func (ha *HighAvailabilityOptionsBuilder) SystemRedisURL(systemRedisURL string) {
	ha.options.systemRedisURL = systemRedisURL
}

func (ha *HighAvailabilityOptionsBuilder) Build() (*HighAvailabilityOptions, error) {

	err := ha.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	ha.setNonRequiredOptions()

	return &ha.options, nil
}

func (ha *HighAvailabilityOptionsBuilder) setRequiredOptions() error {
	if ha.options.apicastProductionRedisURL == "" {
		return fmt.Errorf("no Apicast production URL has been provided")
	}
	if ha.options.apicastStagingRedisURL == "" {
		return fmt.Errorf("no Apicast staging redis URL has been provided")
	}
	if ha.options.backendRedisQueuesEndpoint == "" {
		return fmt.Errorf("no Backend Redis queues endpoint option has been provided")
	}
	if ha.options.backendRedisStorageEndpoint == "" {
		return fmt.Errorf("no Backend Redis storage endpoint has been provided")
	}
	if ha.options.systemDatabaseURL == "" {
		return fmt.Errorf("no System database URL has been provided")
	}
	if ha.options.systemRedisURL == "" {
		return fmt.Errorf("no System redis URL has been provided")
	}

	return nil
}

func (ha *HighAvailabilityOptionsBuilder) setNonRequiredOptions() {

}

type HighAvailabilityOptionsProvider interface {
	GetHighAvailabilityOptions() *HighAvailabilityOptions
}
type CLIHighAvailabilityOptionsProvider struct {
}

func (o *CLIHighAvailabilityOptionsProvider) GetHighAvailabilityOptions() (*HighAvailabilityOptions, error) {
	hob := HighAvailabilityOptionsBuilder{}
	hob.ApicastProductionRedisURL("${APICAST_PRODUCTION_REDIS_URL}")
	hob.ApicastStagingRedisURL("${APICAST_STAGING_REDIS_URL}")
	hob.BackendRedisQueuesEndpoint("${BACKEND_REDIS_QUEUES_ENDPOINT}")
	hob.BackendRedisStorageEndpoint("${BACKEND_REDIS_STORAGE_ENDPOINT}")
	hob.SystemDatabaseURL("${SYSTEM_DATABASE_URL}")
	hob.SystemRedisURL("${SYSTEM_REDIS_URL}")
	res, err := hob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create High Availability Options - %s", err)
	}
	return res, nil
}

func (ha *HighAvailability) setHAOptions() {
	// TODO move this outside this specific method
	optionsProvider := CLIHighAvailabilityOptionsProvider{}
	haOpts, err := optionsProvider.GetHighAvailabilityOptions()
	_ = err
	ha.Options = haOpts
}

func (ha *HighAvailability) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	ha.setHAOptions() // TODO move this outside
	ha.buildParameters(template)
}

// TODO check how to postprocess independently of templates
func (ha *HighAvailability) PostProcess(template *templatev1.Template, otherComponents []Component) {
	res := template.Objects
	ha.setHAOptions() // TODO move this outside
	ha.increaseReplicasNumber(res)
	res = ha.deleteInternalDatabasesObjects(res)
	ha.updateDatabasesURLS(res)
	ha.deleteDBRelatedParameters(template)

	template.Objects = res
}

func (ha *HighAvailability) PostProcessObjects(objects []runtime.RawExtension) []runtime.RawExtension {
	res := objects
	ha.increaseReplicasNumber(res)
	res = ha.deleteInternalDatabasesObjects(res)
	ha.updateDatabasesURLS(res)

	return res
}

func (ha *HighAvailability) increaseReplicasNumber(objects []runtime.RawExtension) {
	// We do not increase the number of replicas in database DeploymentConfigs
	excludedDeploymentConfigs := map[string]bool{
		"system-memcache": true,
		"system-sphinx":   true,
		"zync-database":   true,
	}

	for _, rawExtension := range objects {
		obj := rawExtension.Object
		dc, ok := obj.(*appsv1.DeploymentConfig)
		if ok {
			if _, isExcluded := excludedDeploymentConfigs[dc.Name]; !isExcluded {
				dc.Spec.Replicas = HighlyAvailableReplicas
			}
		}
	}
}

func (ha *HighAvailability) deleteInternalDatabasesObjects(objects []runtime.RawExtension) []runtime.RawExtension {
	keepObjects := []runtime.RawExtension{}

	for rawExtIdx, rawExtension := range objects {
		switch obj := (rawExtension.Object).(type) {
		case *appsv1.DeploymentConfig:
			if _, ok := highlyAvailableExternalDatabases[obj.ObjectMeta.Name]; !ok {
				// We create a new array and add to it the elements that will
				//NOT have to be deleted
				keepObjects = append(keepObjects, objects[rawExtIdx])
			}
		case *v1.Service:
			if _, ok := highlyAvailableExternalDatabases[obj.ObjectMeta.Name]; !ok {
				keepObjects = append(keepObjects, objects[rawExtIdx])
			}
		case *v1.PersistentVolumeClaim:
			if obj.ObjectMeta.Name != "backend-redis-storage" && obj.ObjectMeta.Name != "system-redis-storage" &&
				obj.ObjectMeta.Name != "mysql-storage" {
				keepObjects = append(keepObjects, objects[rawExtIdx])
			}
		case *v1.ConfigMap:
			if obj.ObjectMeta.Name != "mysql-main-conf" && obj.ObjectMeta.Name != "mysql-extra-conf" &&
				obj.ObjectMeta.Name != "redis-config" {
				keepObjects = append(keepObjects, objects[rawExtIdx])
			}
		default:
			keepObjects = append(keepObjects, objects[rawExtIdx])
		}
	}

	return keepObjects
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

func (ha *HighAvailability) updateDatabasesURLS(objects []runtime.RawExtension) {
	for rawExtIdx := range objects {
		obj := objects[rawExtIdx].Object
		secret, ok := obj.(*v1.Secret)
		if ok {
			switch secret.Name {
			case "system-redis":
				secret.StringData["URL"] = ha.Options.systemRedisURL

			// TODO delete mysql-standalone specific parameters
			case "system-database":
				secret.StringData["URL"] = ha.Options.systemDatabaseURL
			case "apicast-redis":
				secret.StringData["PRODUCTION_URL"] = ha.Options.apicastProductionRedisURL
				secret.StringData["STAGING_URL"] = ha.Options.apicastStagingRedisURL
			case "backend-redis":
				secret.StringData["REDIS_STORAGE_URL"] = ha.Options.backendRedisStorageEndpoint
				secret.StringData["REDIS_QUEUES_URL"] = ha.Options.backendRedisQueuesEndpoint
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

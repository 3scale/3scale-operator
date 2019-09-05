package component

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type BackendOptions struct {
	// Non required Options
	serviceEndpoint              *string
	routeEndpoint                *string
	storageURL                   *string
	queuesURL                    *string
	storageSentinelHosts         *string
	storageSentinelRole          *string
	queuesSentinelHosts          *string
	queuesSentinelRole           *string
	listenerResourceRequirements *v1.ResourceRequirements
	workerResourceRequirements   *v1.ResourceRequirements
	cronResourceRequirements     *v1.ResourceRequirements
	listenerReplicas             *int32
	workerReplicas               *int32
	cronReplicas                 *int32

	// required Options
	appLabel              string
	systemBackendUsername string
	systemBackendPassword string
	tenantName            string
	wildcardDomain        string
}

type BackendOptionsBuilder struct {
	options BackendOptions
}

func (m *BackendOptionsBuilder) AppLabel(appLabel string) {
	m.options.appLabel = appLabel
}

func (m *BackendOptionsBuilder) SystemBackendUsername(systemBackendUsername string) {
	m.options.systemBackendUsername = systemBackendUsername
}

func (m *BackendOptionsBuilder) SystemBackendPassword(systemBackendPassword string) {
	m.options.systemBackendPassword = systemBackendPassword
}

func (m *BackendOptionsBuilder) TenantName(tenantName string) {
	m.options.tenantName = tenantName
}

func (m *BackendOptionsBuilder) WildcardDomain(wildcardDomain string) {
	m.options.wildcardDomain = wildcardDomain
}

func (m *BackendOptionsBuilder) ListenerServiceEndpoint(serviceEndpoint *string) {
	m.options.serviceEndpoint = serviceEndpoint
}

func (m *BackendOptionsBuilder) ListenerRouteEndpoint(routeEndpoint *string) {
	m.options.routeEndpoint = routeEndpoint
}

func (m *BackendOptionsBuilder) RedisStorageURL(url *string) {
	m.options.storageURL = url
}

func (m *BackendOptionsBuilder) RedisQueuesURL(url *string) {
	m.options.queuesURL = url
}

func (m *BackendOptionsBuilder) RedisStorageSentinelHosts(hosts *string) {
	m.options.storageSentinelHosts = hosts
}

func (m *BackendOptionsBuilder) RedisStorageSentinelRole(role *string) {
	m.options.storageSentinelRole = role
}

func (m *BackendOptionsBuilder) RedisQueuesSentinelHosts(hosts *string) {
	m.options.queuesSentinelHosts = hosts
}

func (m *BackendOptionsBuilder) RedisQueuesSentinelRole(role *string) {
	m.options.queuesSentinelRole = role
}

func (m *BackendOptionsBuilder) ListenerResourceRequirements(resourceRequirements v1.ResourceRequirements) {
	m.options.listenerResourceRequirements = &resourceRequirements
}

func (m *BackendOptionsBuilder) WorkerResourceRequirements(resourceRequirements v1.ResourceRequirements) {
	m.options.workerResourceRequirements = &resourceRequirements
}

func (m *BackendOptionsBuilder) CronResourceRequirements(resourceRequirements v1.ResourceRequirements) {
	m.options.cronResourceRequirements = &resourceRequirements
}

func (m *BackendOptionsBuilder) ListenerReplicas(replicas int32) {
	m.options.listenerReplicas = &replicas
}

func (m *BackendOptionsBuilder) WorkerReplicas(replicas int32) {
	m.options.workerReplicas = &replicas
}

func (m *BackendOptionsBuilder) CronReplicas(replicas int32) {
	m.options.cronReplicas = &replicas
}

func (m *BackendOptionsBuilder) Build() (*BackendOptions, error) {
	err := m.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	m.setNonRequiredOptions()

	return &m.options, nil
}

func (m *BackendOptionsBuilder) setRequiredOptions() error {
	if m.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if m.options.systemBackendUsername == "" {
		return fmt.Errorf("no System's Backend Username has been provided")
	}
	if m.options.systemBackendPassword == "" {
		return fmt.Errorf("no System's Backend Password has been provided")
	}
	if m.options.tenantName == "" {
		return fmt.Errorf("no tenant name has been provided")
	}
	if m.options.wildcardDomain == "" {
		return fmt.Errorf("no wildcard domain has been provided")
	}

	return nil
}

func (m *BackendOptionsBuilder) setNonRequiredOptions() {

	defaultServiceEndpoint := "http://backend-listener:3000"
	defaultRouteEndpoint := "https://backend-" + m.options.tenantName + "." + m.options.wildcardDomain

	defaultStorageURL := "redis://backend-redis:6379/0"
	defaultQueuesURL := "redis://backend-redis:6379/1"
	defaultStorageSentinelHosts := ""
	defaultStorageSentinelRole := ""
	defaultQueuesSentinelHosts := ""
	defaultQueuesSentinelRole := ""

	if m.options.serviceEndpoint == nil {
		m.options.serviceEndpoint = &defaultServiceEndpoint
	}
	if m.options.routeEndpoint == nil {
		m.options.routeEndpoint = &defaultRouteEndpoint
	}
	if m.options.storageURL == nil {
		m.options.storageURL = &defaultStorageURL
	}
	if m.options.queuesURL == nil {
		m.options.queuesURL = &defaultQueuesURL
	}
	if m.options.storageSentinelHosts == nil {
		m.options.storageSentinelHosts = &defaultStorageSentinelHosts
	}
	if m.options.storageSentinelRole == nil {
		m.options.storageSentinelRole = &defaultStorageSentinelRole
	}
	if m.options.queuesSentinelHosts == nil {
		m.options.queuesSentinelHosts = &defaultQueuesSentinelHosts
	}
	if m.options.queuesSentinelRole == nil {
		m.options.queuesSentinelRole = &defaultQueuesSentinelRole
	}

	if m.options.listenerResourceRequirements == nil {
		m.options.listenerResourceRequirements = m.defaultListenerResourceRequirements()
	}

	if m.options.workerResourceRequirements == nil {
		m.options.workerResourceRequirements = m.defaultWorkerResourceRequirements()
	}

	if m.options.cronResourceRequirements == nil {
		m.options.cronResourceRequirements = m.defaultCronResourceRequirements()
	}

	if m.options.listenerReplicas == nil {
		var listenerDefaultReplicas int32 = 1
		m.options.listenerReplicas = &listenerDefaultReplicas
	}

	if m.options.workerReplicas == nil {
		var workerDefaultReplicas int32 = 1
		m.options.workerReplicas = &workerDefaultReplicas
	}

	if m.options.cronReplicas == nil {
		var cronDefaultReplicas int32 = 1
		m.options.cronReplicas = &cronDefaultReplicas
	}
}

func (m *BackendOptionsBuilder) defaultListenerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("700Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("550Mi"),
		},
	}
}

func (m *BackendOptionsBuilder) defaultWorkerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("300Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("150m"),
			v1.ResourceMemory: resource.MustParse("50Mi"),
		},
	}
}

func (m *BackendOptionsBuilder) defaultCronResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("150m"),
			v1.ResourceMemory: resource.MustParse("80Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("50m"),
			v1.ResourceMemory: resource.MustParse("40Mi"),
		},
	}
}

package component

import "fmt"

type BackendOptions struct {
	// Non required Options
	serviceEndpoint      *string
	routeEndpoint        *string
	storageURL           *string
	queuesURL            *string
	storageSentinelHosts *string
	storageSentinelRole  *string
	queuesSentinelHosts  *string
	queuesSentinelRole   *string
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

func (m *BackendOptionsBuilder) ListenerServiceEndpoint(serviceEndpoint string) {
	m.options.serviceEndpoint = &serviceEndpoint
}

func (m *BackendOptionsBuilder) ListenerRouteEndpoint(routeEndpoint string) {
	m.options.routeEndpoint = &routeEndpoint
}

func (m *BackendOptionsBuilder) RedisStorageURL(url string) {
	m.options.storageURL = &url
}

func (m *BackendOptionsBuilder) RedisQueuesURL(url string) {
	m.options.queuesURL = &url
}

func (m *BackendOptionsBuilder) RedisStorageSentinelHosts(hosts string) {
	m.options.storageSentinelHosts = &hosts
}

func (m *BackendOptionsBuilder) RedisStorageSentinelRole(role string) {
	m.options.storageSentinelRole = &role
}

func (m *BackendOptionsBuilder) RedisQueuesSentinelHosts(hosts string) {
	m.options.queuesSentinelHosts = &hosts
}

func (m *BackendOptionsBuilder) RedisQueuesSentinelRole(role string) {
	m.options.queuesSentinelRole = &role
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
}

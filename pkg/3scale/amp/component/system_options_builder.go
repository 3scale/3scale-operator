package component

import "fmt"

type SystemOptionsBuilder struct {
	options SystemOptions
}

func (s *SystemOptionsBuilder) AdminAccessToken(adminAccessToken string) {
	s.options.adminAccessToken = adminAccessToken
}

func (s *SystemOptionsBuilder) AdminPassword(adminPassword string) {
	s.options.adminPassword = adminPassword
}

func (s *SystemOptionsBuilder) AdminUsername(adminUsername string) {
	s.options.adminUsername = adminUsername
}

func (s *SystemOptionsBuilder) AdminEmail(adminEmail string) {
	s.options.adminEmail = &adminEmail
}

func (s *SystemOptionsBuilder) AmpRelease(ampRelease string) {
	s.options.ampRelease = ampRelease
}

func (s *SystemOptionsBuilder) ApicastAccessToken(apicastAccessToken string) {
	s.options.apicastAccessToken = apicastAccessToken
}

func (s *SystemOptionsBuilder) ApicastRegistryURL(apicastRegistryURL string) {
	s.options.apicastRegistryURL = apicastRegistryURL
}

func (s *SystemOptionsBuilder) AppLabel(appLabel string) {
	s.options.appLabel = appLabel
}

func (s *SystemOptionsBuilder) MasterAccessToken(masterAccessToken string) {
	s.options.masterAccessToken = masterAccessToken
}

func (s *SystemOptionsBuilder) MasterName(masterName string) {
	s.options.masterName = masterName
}

func (s *SystemOptionsBuilder) MasterUsername(masterUsername string) {
	s.options.masterUsername = masterUsername
}

func (s *SystemOptionsBuilder) MasterPassword(masterPassword string) {
	s.options.masterPassword = masterPassword
}

func (s *SystemOptionsBuilder) RecaptchaPublicKey(recaptchaPublicKey string) {
	s.options.recaptchaPublicKey = recaptchaPublicKey
}

func (s *SystemOptionsBuilder) RecaptchaPrivateKey(recaptchaPrivateKey string) {
	s.options.recaptchaPrivateKey = recaptchaPrivateKey
}

func (s *SystemOptionsBuilder) AppSecretKeyBase(appSecretKeyBase string) {
	s.options.appSecretKeyBase = appSecretKeyBase
}

func (s *SystemOptionsBuilder) BackendSharedSecret(backendSharedSecret string) {
	s.options.backendSharedSecret = backendSharedSecret
}

func (s *SystemOptionsBuilder) TenantName(tenantName string) {
	s.options.tenantName = tenantName
}

func (s *SystemOptionsBuilder) WildcardDomain(wildcardDomain string) {
	s.options.wildcardDomain = wildcardDomain
}

func (s *SystemOptionsBuilder) StorageClassName(storageClassName *string) {
	s.options.storageClassName = storageClassName
}

func (s *SystemOptionsBuilder) MemcachedServers(servers string) {
	s.options.memcachedServers = &servers
}

func (s *SystemOptionsBuilder) EventHooksURL(eventHooksURL string) {
	s.options.eventHooksURL = &eventHooksURL
}

func (s *SystemOptionsBuilder) RedisURL(redisURL string) {
	s.options.redisURL = &redisURL
}

func (s *SystemOptionsBuilder) RedisSentinelHosts(hosts string) {
	s.options.redisSentinelHosts = &hosts
}

func (s *SystemOptionsBuilder) RedisSentinelRole(role string) {
	s.options.redisSentinelRole = &role
}

func (s *SystemOptionsBuilder) MessageBusRedisURL(url string) {
	s.options.messageBusRedisURL = &url
}

func (s *SystemOptionsBuilder) MessageBusRedisSentinelHosts(hosts string) {
	s.options.messageBusRedisSentinelHosts = &hosts
}

func (s *SystemOptionsBuilder) MessageBusRedisSentinelRole(role string) {
	s.options.messageBusRedisSentinelRole = &role
}

func (s *SystemOptionsBuilder) RedisNamespace(namespace string) {
	s.options.redisNamespace = &namespace
}

func (s *SystemOptionsBuilder) MessageBusRedisNamespace(namespace string) {
	s.options.messageBusRedisNamespace = &namespace
}

func (s *SystemOptionsBuilder) ApicastSystemMasterProxyConfigEndpoint(endpoint string) {
	s.options.apicastSystemMasterProxyConfigEndpoint = &endpoint
}

func (s *SystemOptionsBuilder) ApicastSystemMasterBaseURL(url string) {
	s.options.apicastSystemMasterBaseURL = &url
}

func (s *SystemOptionsBuilder) Build() (*SystemOptions, error) {
	err := s.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	s.setNonRequiredOptions()

	return &s.options, nil
}

func (s *SystemOptionsBuilder) setRequiredOptions() error {
	if s.options.adminAccessToken == "" {
		return fmt.Errorf("no admin access token has been provided")
	}
	if s.options.adminPassword == "" {
		return fmt.Errorf("no admin password has been provided")
	}
	if s.options.adminUsername == "" {
		return fmt.Errorf("no admin username has been provided")
	}
	if s.options.ampRelease == "" {
		return fmt.Errorf("no AMP release has been provided")
	}
	if s.options.apicastAccessToken == "" {
		return fmt.Errorf("no apicast access token has been provided")
	}
	if s.options.apicastRegistryURL == "" {
		return fmt.Errorf("no apicast registry url has been provided")
	}
	if s.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if s.options.masterAccessToken == "" {
		return fmt.Errorf("no master access token has been provided")
	}
	if s.options.masterName == "" {
		return fmt.Errorf("no master name has been provided")
	}
	if s.options.masterUsername == "" {
		return fmt.Errorf("no master username has been provided")
	}
	if s.options.masterPassword == "" {
		return fmt.Errorf("no master password has been provided")
	}
	if s.options.appSecretKeyBase == "" {
		return fmt.Errorf("no app secret keybase has been provided")
	}
	if s.options.backendSharedSecret == "" {
		return fmt.Errorf("no backend shared secret has been provided")
	}
	if s.options.tenantName == "" {
		return fmt.Errorf("no tenant name has been provided")
	}
	if s.options.wildcardDomain == "" {
		return fmt.Errorf("no wildcard domain has been provided")
	}

	return nil
}

func (s *SystemOptionsBuilder) setNonRequiredOptions() {
	defaultMemcachedServers := "system-memcache:11211"
	defaultEventHooksURL := "http://system-master:3000/master/events/import"

	defaultApicastSystemMasterProxyConfigEndpoint := "http://" + s.options.apicastAccessToken + "@system-master:3000/master/api/proxy/configs"
	defaultApicastSystemMasterBaseURL := "http://" + s.options.apicastAccessToken + "@system-master:3000"
	defaultAdminEmail := ""

	if s.options.memcachedServers == nil {
		s.options.memcachedServers = &defaultMemcachedServers
	}

	if s.options.eventHooksURL == nil {
		s.options.eventHooksURL = &defaultEventHooksURL
	}

	s.setRedisDefaultsOptions()

	if s.options.apicastSystemMasterProxyConfigEndpoint == nil {
		s.options.apicastSystemMasterProxyConfigEndpoint = &defaultApicastSystemMasterProxyConfigEndpoint
	}

	if s.options.apicastSystemMasterBaseURL == nil {
		s.options.apicastSystemMasterBaseURL = &defaultApicastSystemMasterBaseURL
	}

	if s.options.adminEmail == nil {
		s.options.adminEmail = &defaultAdminEmail
	}
}

func (s *SystemOptionsBuilder) setRedisDefaultsOptions() {
	defaultRedisURL := "redis://system-redis:6379/1"
	defaultMessageBusRedisURL := ""
	defaultRedisNamespace := ""
	defaultMessageBusRedisNamespace := ""
	defaultRedisSentinelHosts := ""
	defaultRedisSentinelRole := ""
	defaultMessageBusRedisSentinelHosts := ""
	defaultMessageBusRedisSentinelRole := ""

	if s.options.redisURL == nil {
		s.options.redisURL = &defaultRedisURL
	}

	if s.options.redisSentinelHosts == nil {
		s.options.redisSentinelHosts = &defaultRedisSentinelHosts
	}

	if s.options.redisSentinelRole == nil {
		s.options.redisSentinelRole = &defaultRedisSentinelRole
	}

	if s.options.messageBusRedisURL == nil {
		s.options.messageBusRedisURL = &defaultMessageBusRedisURL
	}

	if s.options.messageBusRedisSentinelHosts == nil {
		s.options.messageBusRedisSentinelHosts = &defaultMessageBusRedisSentinelHosts
	}

	if s.options.messageBusRedisSentinelRole == nil {
		s.options.messageBusRedisSentinelRole = &defaultMessageBusRedisSentinelRole
	}

	if s.options.redisNamespace == nil {
		s.options.redisNamespace = &defaultRedisNamespace
	}

	if s.options.messageBusRedisNamespace == nil {
		s.options.messageBusRedisNamespace = &defaultMessageBusRedisNamespace
	}
}

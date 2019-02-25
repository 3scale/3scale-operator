package product

type release_2_4 struct{}

func (p *release_2_4) GetApicastImage() string {
	return "registry.access.redhat.com/3scale-amp24/apicast-gateway"
}

func (p *release_2_4) GetBackendImage() string {
	return "registry.access.redhat.com/3scale-amp24/backend"
}

func (p *release_2_4) GetBackendRedisImage() string {
	return "registry.access.redhat.com/rhscl/redis-32-rhel7:3.2"
}

func (p *release_2_4) GetSystemImage() string {
	return "registry.access.redhat.com/3scale-amp24/system"
}

func (p *release_2_4) GetSystemRedisImage() string {
	return "registry.access.redhat.com/rhscl/redis-32-rhel7:3.2"
}

func (p *release_2_4) GetSystemMySQLImage() string {
	return "registry.access.redhat.com/rhscl/mysql-57-rhel7:5.7"
}

func (p *release_2_4) GetSystemPostgreSQLImage() string {
	return "registry.access.redhat.com/rhscl/postgresql-95-rhel7:9.5"
}

func (p *release_2_4) GetSystemMemcachedImage() string {
	return "registry.access.redhat.com/3scale-amp20/memcached"
}

func (p *release_2_4) GetWildcardRouterImage() string {
	return "registry.access.redhat.com/3scale-amp22/wildcard-router"
}

func (p *release_2_4) GetZyncImage() string {
	return "registry.access.redhat.com/3scale-amp24/zync"
}

func (p *release_2_4) GetZyncPostgreSQLImage() string {
	return "registry.access.redhat.com/rhscl/postgresql-95-rhel7:9.5"
}

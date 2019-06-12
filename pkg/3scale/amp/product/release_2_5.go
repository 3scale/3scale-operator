package product

type release_2_5 struct{}

func (p *release_2_5) GetApicastImage() string {
	return "registry.access.redhat.com/3scale-amp25/apicast-gateway"
}

func (p *release_2_5) GetBackendImage() string {
	return "registry.access.redhat.com/3scale-amp25/backend"
}

func (p *release_2_5) GetBackendRedisImage() string {
	return "registry.access.redhat.com/rhscl/redis-32-rhel7:3.2"
}

func (p *release_2_5) GetSystemImage() string {
	return "registry.access.redhat.com/3scale-amp25/system"
}

func (p *release_2_5) GetSystemRedisImage() string {
	return "registry.access.redhat.com/rhscl/redis-32-rhel7:3.2"
}

func (p *release_2_5) GetSystemMySQLImage() string {
	return "registry.access.redhat.com/rhscl/mysql-57-rhel7:5.7"
}

func (p *release_2_5) GetSystemPostgreSQLImage() string {
	return "registry.access.redhat.com/rhscl/postgresql-10-rhel7"
}

func (p *release_2_5) GetSystemMemcachedImage() string {
	return "registry.access.redhat.com/3scale-amp20/memcached"
}

func (p *release_2_5) GetZyncImage() string {
	return "registry.access.redhat.com/3scale-amp25/zync"
}

func (p *release_2_5) GetZyncPostgreSQLImage() string {
	return "registry.access.redhat.com/rhscl/postgresql-10-rhel7"
}

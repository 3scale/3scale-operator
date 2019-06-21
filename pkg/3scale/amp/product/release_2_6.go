package product

type release_2_6 struct{}

func (p *release_2_6) GetApicastImage() string {
	return "registry.redhat.io/3scale-amp26/apicast-gateway"
}

func (p *release_2_6) GetBackendImage() string {
	return "registry.redhat.io/3scale-amp26/backend"
}

func (p *release_2_6) GetBackendRedisImage() string {
	return "registry.redhat.io/rhscl/redis-32-rhel7:3.2"
}

func (p *release_2_6) GetSystemImage() string {
	return "registry.redhat.io/3scale-amp26/system"
}

func (p *release_2_6) GetSystemRedisImage() string {
	return "registry.redhat.io/rhscl/redis-32-rhel7:3.2"
}

func (p *release_2_6) GetSystemMySQLImage() string {
	return "registry.redhat.io/rhscl/mysql-57-rhel7:5.7"
}

func (p *release_2_6) GetSystemPostgreSQLImage() string {
	return "registry.redhat.io/rhscl/postgresql-10-rhel7"
}

func (p *release_2_6) GetSystemMemcachedImage() string {
	return "registry.redhat.io/3scale-amp20/memcached"
}

func (p *release_2_6) GetZyncImage() string {
	return "registry.redhat.io/3scale-amp26/zync"
}

func (p *release_2_6) GetZyncPostgreSQLImage() string {
	return "registry.redhat.io/rhscl/postgresql-10-rhel7"
}

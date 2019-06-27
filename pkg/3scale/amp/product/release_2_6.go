package product

type release_2_6 struct{}

func (p *release_2_6) GetApicastImage() string {
	return "quay.io/3scale/3scale26:apicast-3scale-2.6.0-ER1"
}

func (p *release_2_6) GetBackendImage() string {
	return "quay.io/3scale/3scale26:apisonator-3scale-2.6.0-ER1"
}

func (p *release_2_6) GetBackendRedisImage() string {
	return "centos/redis-32-centos7"
}

func (p *release_2_6) GetSystemImage() string {
	return "quay.io/3scale/3scale26:porta-3scale-2.6.0-ER1"
}

func (p *release_2_6) GetSystemRedisImage() string {
	return "centos/redis-32-centos7"
}

func (p *release_2_6) GetSystemMySQLImage() string {
	return "centos/mysql-57-centos7"
}

func (p *release_2_6) GetSystemPostgreSQLImage() string {
	return "centos/postgresql-10-centos7"
}

func (p *release_2_6) GetSystemMemcachedImage() string {
	return "memcached:1.5"
}

func (p *release_2_6) GetZyncImage() string {
	return "quay.io/3scale/3scale26:zync-3scale-2.6.0-ER1"
}

func (p *release_2_6) GetZyncPostgreSQLImage() string {
	return "centos/postgresql-10-centos7"
}

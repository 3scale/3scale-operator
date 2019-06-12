package product

type upstream struct{}

func (u *upstream) GetApicastImage() string {
	return "quay.io/3scale/apicast:nightly"
}

func (u *upstream) GetBackendImage() string {
	return "quay.io/3scale/apisonator:nightly"
}

func (u *upstream) GetBackendRedisImage() string {
	return "registry.access.redhat.com/rhscl/redis-32-rhel7:3.2"
}

func (u *upstream) GetSystemImage() string {
	return "quay.io/3scale/porta:nightly"
}

func (u *upstream) GetSystemRedisImage() string {
	return "registry.access.redhat.com/rhscl/redis-32-rhel7:3.2"
}

func (u *upstream) GetSystemMySQLImage() string {
	return "registry.access.redhat.com/rhscl/mysql-57-rhel7:5.7"
}

func (u *upstream) GetSystemPostgreSQLImage() string {
	return "registry.access.redhat.com/rhscl/postgresql-10-rhel7"
}

func (u *upstream) GetSystemMemcachedImage() string {
	return "registry.access.redhat.com/3scale-amp20/memcached"
}

func (u *upstream) GetZyncImage() string {
	return "quay.io/3scale/zync:nightly"
}

func (u *upstream) GetZyncPostgreSQLImage() string {
	return "registry.access.redhat.com/rhscl/postgresql-10-rhel7"
}

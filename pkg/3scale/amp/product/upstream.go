package product

type upstream struct{}

func (u *upstream) GetApicastImage() string {
	return "quay.io/3scale/apicast:nightly"
}

func (u *upstream) GetBackendImage() string {
	return "quay.io/3scale/apisonator:nightly"
}

func (u *upstream) GetBackendRedisImage() string {
	return "centos/redis-32-centos7"
}

func (u *upstream) GetSystemImage() string {
	return "quay.io/3scale/porta:nightly"
}

func (u *upstream) GetSystemRedisImage() string {
	return "centos/redis-32-centos7"
}

func (u *upstream) GetSystemMySQLImage() string {
	return "centos/mysql-57-centos7"
}

func (u *upstream) GetSystemPostgreSQLImage() string {
	return "centos/postgresql-10-centos7"
}

func (u *upstream) GetSystemMemcachedImage() string {
	return "memcached:1.5"
}

func (u *upstream) GetZyncImage() string {
	return "quay.io/3scale/zync:nightly"
}

func (u *upstream) GetZyncPostgreSQLImage() string {
	return "centos/postgresql-10-centos7"
}

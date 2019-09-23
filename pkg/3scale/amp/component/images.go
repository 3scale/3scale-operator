package component

func ApicastImageURL() string {
	return "quay.io/3scale/apicast:nightly"
}

func ApisonatorImageURL() string {
	return "quay.io/3scale/apisonator:nightly"
}

func PortaImageURL() string {
	return "quay.io/3scale/porta:nightly"
}

func ZyncImageURL() string {
	return "quay.io/3scale/zync:nightly"
}

func BackendRedisImageURL() string {
	return "centos/redis-32-centos7"
}

func SystemRedisImageURL() string {
	return "centos/redis-32-centos7"
}

func SystemMySQLImageURL() string {
	return "centos/mysql-57-centos7"
}

func SystemPostgreSQLImageURL() string {
	return "centos/postgresql-10-centos7"
}

func PortaMemcachedImageURL() string {
	return "memcached:1.5"
}

func ZyncPostgreSQLImageURL() string {
	return "centos/postgresql-10-centos7"
}

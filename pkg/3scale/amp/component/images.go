package component

func ApicastImageURL() string {
	return "quay.io/3scale/3scale27:apicast-3scale-2.7.0-GA"
}

func BackendImageURL() string {
	return "quay.io/3scale/3scale27:apisonator-3scale-2.7.0-GA"
}

func SystemImageURL() string {
	return "quay.io/3scale/3scale27:porta-3scale-2.7.0-GA"
}

func ZyncImageURL() string {
	return "quay.io/3scale/3scale27:zync-3scale-2.7.0-GA"
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

func SystemMemcachedImageURL() string {
	return "memcached:1.5"
}

func ZyncPostgreSQLImageURL() string {
	return "centos/postgresql-10-centos7"
}

package component

func ApicastImageURL() string {
	return "quay.io/3scale/3scale210:apicast-3scale-2.10.0-GA"
}

func BackendImageURL() string {
	return "quay.io/3scale/3scale210:apisonator-3scale-2.10.0-GA"
}

func SystemImageURL() string {
	return "quay.io/3scale/3scale210:porta-3scale-2.10.0-GA"
}

func ZyncImageURL() string {
	return "quay.io/3scale/3scale210:zync-3scale-2.10.0-GA"
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

func OCCLIImageURL() string {
	return "quay.io/openshift/origin-cli:4.2"
}

package component

func ApicastImageURL() string {
	return "quay.io/3scale/3scale214:apicast-3scale-2.14.1-GA"
}

func BackendImageURL() string {
	return "quay.io/3scale/3scale214:apisonator-3scale-2.14.1-GA"
}

func SystemImageURL() string {
	return "quay.io/3scale/3scale214:porta-3scale-2.14.1-GA"
}

func SystemSearchdImageURL() string {
	return "quay.io/3scale/searchd:latest"
}

func ZyncImageURL() string {
	return "quay.io/3scale/3scale214:zync-3scale-2.14.1-GA"
}

func BackendRedisImageURL() string {
	return "quay.io/fedora/redis-6"
}

func SystemRedisImageURL() string {
	return "quay.io/fedora/redis-6"
}

func SystemMySQLImageURL() string {
	return "centos/mysql-80-centos7"
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
	return "quay.io/openshift/origin-cli:4.7"
}

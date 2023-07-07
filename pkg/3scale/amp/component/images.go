package component

func ApicastImageURL() string {
	return "quay.io/3scale/apicast:latest"
}

func BackendImageURL() string {
	return "quay.io/3scale/apisonator:latest"
}

func SystemImageURL() string {
	return "quay.io/3scale/porta:latest"
}

func SystemSearchdImageURL() string {
	return "quay.io/3scale/searchd:latest"
}

func ZyncImageURL() string {
	return "quay.io/3scale/zync:latest"
}

func BackendRedisImageURL() string {
	return "quay.io/centos7/redis-6-centos7:latest"
}

func SystemRedisImageURL() string {
	return "quay.io/centos7/redis-6-centos7:latest"
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

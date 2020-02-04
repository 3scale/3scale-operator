package component

func ApicastImageURL() string {
	return "registry.redhat.io/3scale-amp2/apicast-gateway-rhel7:3scale2.8"
}

func BackendImageURL() string {
	return "registry.redhat.io/3scale-amp2/backend-rhel7:3scale2.8"
}

func SystemImageURL() string {
	return "registry.redhat.io/3scale-amp2/system-rhel7:3scale2.8"
}

func ZyncImageURL() string {
	return "registry.redhat.io/3scale-amp2/zync-rhel7:3scale2.8"
}

func BackendRedisImageURL() string {
	return "registry.redhat.io/rhscl/redis-32-rhel7:3.2"
}

func SystemRedisImageURL() string {
	return "registry.redhat.io/rhscl/redis-32-rhel7:3.2"
}

func SystemMySQLImageURL() string {
	return "registry.redhat.io/rhscl/mysql-57-rhel7:5.7"
}

func SystemPostgreSQLImageURL() string {
	return "registry.redhat.io/rhscl/postgresql-10-rhel7"
}

func SystemMemcachedImageURL() string {
	return "registry.redhat.io/3scale-amp2/memcached-rhel7:3scale2.8"
}

func ZyncPostgreSQLImageURL() string {
	return "registry.redhat.io/rhscl/postgresql-10-rhel7"
}

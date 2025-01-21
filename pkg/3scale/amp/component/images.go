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

func SystemMemcachedImageURL() string {
	return "mirror.gcr.io/library/memcached:1.5"
}

func ZyncPostgreSQLImageURL() string {
	return "mirror.gcr.io/library/postgres:13"
}

func OCCLIImageURL() string {
	return "quay.io/openshift/origin-cli:4.7"
}

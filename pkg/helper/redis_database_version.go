package helper

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/go-logr/logr"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	systemRedisUrl              = "URL"
	systemRedisUsername         = "REDIS_USERNAME"
	systemRedisPassword         = "REDIS_PASSWORD"
	systemRedisSentinelHosts    = "SENTINEL_HOSTS"
	systemRedisSentinelUsername = "REDIS_SENTINEL_USERNAME"
	systemRedisSentinelPassword = "REDIS_SENTINEL_PASSWORD"
	systemRedisCA               = "REDIS_SSL_CA"
	systemRedisCertificate      = "REDIS_SSL_CERT"
	systemRedisKey              = "REDIS_SSL_KEY"

	backendRedisQueuesURL              = "REDIS_QUEUES_URL"
	backendRedisQueuesUsername         = "REDIS_QUEUES_USERNAME"
	backendRedisQueuesPassword         = "REDIS_QUEUES_PASSWORD"
	backendRedisQueuesSentinelHosts    = "REDIS_QUEUES_SENTINEL_HOSTS"
	backendRedisQueuesSentinelRole     = "REDIS_QUEUES_SENTINEL_ROLE"
	backendRedisQueuesSentinelUsername = "REDIS_QUEUES_SENTINEL_USERNAME"
	backendRedisQueuesSentinelPassword = "REDIS_QUEUES_SENTINEL_PASSWORD"
	backendRedisQueuesCA               = "REDIS_SSL_QUEUES_CA"
	backendRedisQueuesCertificate      = "REDIS_SSL_QUEUES_CERT"
	backendRedisQueuesKey              = "REDIS_SSL_QUEUES_KEY"

	backendRedisStorageURL              = "REDIS_STORAGE_URL"
	backendRedisStorageUsername         = "REDIS_STORAGE_USERNAME"
	backendRedisStoragePassword         = "REDIS_STORAGE_PASSWORD"
	backendRedisStorageSentinelHosts    = "REDIS_STORAGE_SENTINEL_HOSTS"
	backendRedisStorageSentinelRole     = "REDIS_STORAGE_SENTINEL_ROLE"
	backendRedisStorageSentinelUsername = "REDIS_STORAGE_SENTINEL_USERNAME"
	backendRedisStorageSentinelPassword = "REDIS_STORAGE_SENTINEL_PASSWORD"
	backendRedisStorageCA               = "REDIS_SSL_CA"
	backendRedisStorageCertificate      = "REDIS_SSL_CERT"
	backendRedisStorageKey              = "REDIS_SSL_KEY"
)

func VerifySystemRedis(k8sclient client.Client, reqConfigMap *v1.ConfigMap, systemRedisRequirement string, apimInstance *appsv1alpha1.APIManager, logger logr.Logger) (bool, error) {
	logger.Info("Verifying system redis version")
	var systemRedisVerified bool
	connSecret, err := fetchSecret(k8sclient, "system-redis", apimInstance.Namespace)
	if err != nil {
		logger.Error(err, "System redis secret not found")
		return false, err
	}

	enabledTLS := apimInstance.IsSystemRedisTLSEnabled()

	systemRedisVerified, err = verifySystemRedisVersion(*connSecret, systemRedisRequirement, enabledTLS)
	if err != nil {
		logger.Error(err, "Encountered error during version verification of system Redis")
		return false, err
	}
	if systemRedisVerified {
		logger.Info("System redis version verified")
	} else {
		logger.Info("System redis version not matching the required version")
	}

	return systemRedisVerified, nil
}

func VerifyBackendRedis(k8sclient client.Client, reqConfigMap *v1.ConfigMap, backendRedisRequirement string, apimInstance *appsv1alpha1.APIManager, logger logr.Logger) (bool, error) {
	logger.Info("Verifying backend redis version")
	var backendRedisVerified bool
	connSecret, err := fetchSecret(k8sclient, "backend-redis", apimInstance.Namespace)
	if err != nil {
		logger.Error(err, "Backend redis secret not found")
		return false, err
	}

	redisStorageEnabledTLS := apimInstance.IsBackendRedisTLSEnabled()
	redisQueueEnabledTLS := apimInstance.IsQueuesRedisTLSEnabled()

	backendRedisVerified, err = verifyBackendRedisVersion(*connSecret, backendRedisRequirement, redisStorageEnabledTLS, redisQueueEnabledTLS)
	if err != nil {
		logger.Error(err, "Encountered error during version verification of backend Redis")
		return false, err
	}
	if backendRedisVerified {
		logger.Info("Backend redis version verified")
	} else {
		logger.Info("Backend redis version not matching the required version")
	}

	return backendRedisVerified, nil
}

func verifySystemRedisVersion(connSecret v1.Secret, requiredVersion string, enabledTLS bool) (bool, error) {
	redisOpts := reconcileSystemRedisSecret(connSecret)

	if enabledTLS {
		redisOpts.TLS.Enabled = true
	}

	rdb, err := Configure(redisOpts)
	if err != nil {
		return false, err
	}

	return verifyRedisVersion(rdb, requiredVersion)
}

func verifyBackendRedisVersion(connSecret v1.Secret, requiredVersion string, enabledRedisStorageTLS bool, enabledRedisQueueTLS bool) (bool, error) {
	redisQueueOpts := reconcileQueuesRedisSecret(connSecret)

	if enabledRedisQueueTLS {
		redisQueueOpts.TLS.Enabled = true
	}

	qrdb, err := Configure(redisQueueOpts)
	if err != nil {
		return false, err
	}

	redisQueuesVersionConfirmed, err := verifyRedisVersion(qrdb, requiredVersion)
	if err != nil {
		return false, err
	}

	redisStorageOpts := reconcileStorageRedisSecret(connSecret)

	if enabledRedisStorageTLS {
		redisStorageOpts.TLS.Enabled = true
	}

	srdb, err := Configure(redisStorageOpts)
	if err != nil {
		return false, err
	}

	redisStorageVersionConfirmed, err := verifyRedisVersion(srdb, requiredVersion)
	if err != nil {
		return false, err
	}

	return redisQueuesVersionConfirmed && redisStorageVersionConfirmed, nil
}

package helper

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/go-logr/logr"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	systemRedisSentinelHosts         = "SENTINEL_HOSTS"
	systemRedisUrl                   = "URL"
	backendRedisQueuesSentinelHosts  = "REDIS_QUEUES_SENTINEL_HOSTS"
	backendRedisStorageSentinelHosts = "REDIS_STORAGE_SENTINEL_HOSTS"
	backendRedisQueuesURL            = "REDIS_QUEUES_URL"
	backendRedisStorageURL           = "REDIS_STORAGE_URL"
)

func VerifySystemRedis(k8sclient client.Client, reqConfigMap *v1.ConfigMap, systemRedisRequirement string, apimInstance *appsv1alpha1.APIManager, logger logr.Logger) (bool, error) {
	logger.Info("Verifying system redis version")
	systemRedisVerified := false
	connSecret, err := fetchSecret(k8sclient, "system-redis", apimInstance.Namespace)
	if err != nil {
		logger.Info("System redis secret not found")
		return false, err
	}

	systemRedisVerified, err = verifySystemRedisVersion(*connSecret, apimInstance.Namespace, systemRedisRequirement, logger)
	if err != nil {
		logger.Info("Encountered error during version verification of system Redis")
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
	backendRedisVerified := false
	connSecret, err := fetchSecret(k8sclient, "backend-redis", apimInstance.Namespace)
	if err != nil {
		logger.Info("Backend redis secret not found")
		return false, err
	}

	backendRedisVerified, err = verifyBackendRedisVersion(*connSecret, apimInstance.Namespace, backendRedisRequirement, logger)
	if err != nil {
		logger.Info("Encountered error during version verification of backend Redis")
		return false, err
	}
	if backendRedisVerified {
		logger.Info("Backend redis version verified")
	} else {
		logger.Info("Backend redis version not matching the required version")
	}

	return backendRedisVerified, nil
}

func verifySystemRedisVersion(connSecret v1.Secret, namespace string, requiredVersion string, logger logr.Logger) (bool, error) {
	redisOpts := reconcileSystemRedisSecret(connSecret)

	rdb, err := Configure(redisOpts)
	if err != nil {
		logger.Info("Failed to setup Redis connection")
		return false, err
	}

	return verifyRedisVersion(rdb, requiredVersion)
}

func verifyBackendRedisVersion(connSecret v1.Secret, namespace string, requiredVersion string, logger logr.Logger) (bool, error) {
	redisQueueOpts := reconcileQueuesRedisSecret(connSecret)

	qrdb, err := Configure(redisQueueOpts)
	if err != nil {
		logger.Info("Failed to setup Redis connection")
		return false, err
	}

	redisQueuesVersionConfirmed, err := verifyRedisVersion(qrdb, requiredVersion)
	if err != nil {
		logger.Info("Failed to verify Redis version")
		return false, err
	}

	redisStorageOpts := reconcileStorageRedisSecret(connSecret)
	srdb, err := Configure(redisStorageOpts)
	if err != nil {
		logger.Info("Failed to setup Redis connection")
		return false, err
	}

	redisStorageVersionConfirmed, err := verifyRedisVersion(srdb, requiredVersion)
	if err != nil {
		logger.Info("Failed to verify Redis version")
		return false, err
	}

	return redisQueuesVersionConfirmed && redisStorageVersionConfirmed, nil
}

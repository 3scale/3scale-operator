package helper

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/go-logr/logr"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	redisSystemSecretName            = "system-redis"
	redisBackendSecretName           = "backend-redis"
	systemRedisSentinelHosts         = "SENTINEL_HOSTS"
	systemRedisUrl                   = "URL"
	backendRedisQueuesSentinelHosts  = "REDIS_QUEUES_SENTINEL_HOSTS"
	backendRedisStorageSentinelHosts = "REDIS_STORAGE_SENTINEL_HOSTS"
	backendRedisQueuesURL            = "REDIS_QUEUES_URL"
	backendRedisStorageURL           = "REDIS_STORAGE_URL"
)

type Redis struct {
	sentinelHost     string
	sentinelPassword string
	sentinelPort     string
	sentinelGroup    string
	redisHost        string
	redisPassword    string
	redisPort        string
}

func VerifySystemRedis(k8sclient client.Client, reqConfigMap *v1.ConfigMap, systemRedisRequirement string, apimInstance *appsv1alpha1.APIManager, logger logr.Logger) (bool, error) {
	logger.Info("Verifying system redis version")
	systemRedisVerified := false
	connSecret, err := fetchSecret(k8sclient, "system-redis", apimInstance.Namespace)
	if err != nil {
		logger.Info("System redis secret not found")
		return false, err
	}

	systemRedisVerified, err = verifySystemRedisVersion(k8sclient, *connSecret, apimInstance.Namespace, systemRedisRequirement, logger)
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

	backendRedisVerified, err = verifyBackendRedisVersion(k8sclient, *connSecret, apimInstance.Namespace, backendRedisRequirement, logger)
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

func verifySystemRedisVersion(k8sclient client.Client, connSecret v1.Secret, namespace string, requiredVersion string, logger logr.Logger) (bool, error) {
	redisCliCommand := ""

	redisPod, err := CreateRedisThrowAwayPod(k8sclient, namespace)
	if err != nil {
		return false, err
	}

	redis := reconcileSystemRedisSecret(connSecret)
	if redis.sentinelHost != "" {
		redisCliCommand = retrieveRedisSentinelCommand(redis.sentinelHost, redis.sentinelPort, redis.sentinelPassword, redis.sentinelGroup)
		stdout, err := executeRedisCliCommand(*redisPod, redisCliCommand, logger)
		if err != nil {
			return false, err
		}
		redis.redisHost, redis.redisPort = parseRedisSentinelResponse(stdout)
	}

	redisCliCommand = retrieveRedisCliCommand(redis.redisHost, redis.redisPort, redis.redisPassword)

	stdout, err := executeRedisCliCommand(*redisPod, redisCliCommand, logger)
	if err != nil {
		return false, err
	}

	currentRedisVersion, err := retrieveCurrentVersionOfRedis(stdout)
	if err != nil {
		return false, err
	}

	redisSystemVersionConfirmed := CompareVersions(requiredVersion, currentRedisVersion)

	return redisSystemVersionConfirmed, nil
}

func verifyBackendRedisVersion(k8sclient client.Client, connSecret v1.Secret, namespace string, requiredVersion string, logger logr.Logger) (bool, error) {
	redisCliCommand := ""

	redisPod, err := CreateRedisThrowAwayPod(k8sclient, namespace)
	if err != nil {
		return false, err
	}

	redis := reconcileQueuesRedisSecret(connSecret)
	if redis.sentinelHost != "" {
		redisCliCommand = retrieveRedisSentinelCommand(redis.sentinelHost, redis.sentinelPort, redis.sentinelPassword, redis.sentinelGroup)
		stdout, err := executeRedisCliCommand(*redisPod, redisCliCommand, logger)
		if err != nil {
			return false, err
		}
		redis.redisHost, redis.redisPort = parseRedisSentinelResponse(stdout)
	}

	redisCliCommand = retrieveRedisCliCommand(redis.redisHost, redis.redisPort, redis.redisPassword)

	stdout, err := executeRedisCliCommand(*redisPod, redisCliCommand, logger)
	if err != nil {
		return false, err
	}

	currentRedisVersion, err := retrieveCurrentVersionOfRedis(stdout)
	if err != nil {
		return false, err
	}

	redisQueuesVersionConfirmed := CompareVersions(requiredVersion, currentRedisVersion)

	redis = reconcileStorageRedisSecret(connSecret)
	if redis.sentinelHost != "" {
		redisCliCommand = retrieveRedisSentinelCommand(redis.sentinelHost, redis.sentinelPort, redis.sentinelPassword, redis.sentinelGroup)
		stdout, err := executeRedisCliCommand(*redisPod, redisCliCommand, logger)
		if err != nil {
			return false, err
		}
		redis.redisHost, redis.redisPort = parseRedisSentinelResponse(stdout)
	}

	redisCliCommand = retrieveRedisCliCommand(redis.redisHost, redis.redisPort, redis.redisPassword)

	stdout, err = executeRedisCliCommand(*redisPod, redisCliCommand, logger)
	if err != nil {
		return false, err
	}

	currentRedisVersion, err = retrieveCurrentVersionOfRedis(stdout)
	if err != nil {
		return false, err
	}

	redisStorageVersionConfirmed := CompareVersions(requiredVersion, currentRedisVersion)

	if redisQueuesVersionConfirmed && redisStorageVersionConfirmed {
		err := DeletePod(k8sclient, redisPod)
		if err != nil {
			return false, nil
		}
	}

	return redisQueuesVersionConfirmed && redisStorageVersionConfirmed, nil
}

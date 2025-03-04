package helper

import (
	"context"

	"github.com/blang/semver/v4"
	"github.com/go-logr/logr"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	OperatorRequirementsConfigMapName     = "3scale-api-management-operator-requirements"
	RHTThreescaleVersion                  = "rht_threescale_version_requirements"
	RHTThreescaleMysqlRequirements        = "rht_mysql_requirements"
	RHTThreescalePostgresRequirements     = "rht_postgres_requirements"
	RHTThreescaleSystemRedisRequirements  = "rht_system_redis_requirements"
	RHTThreescaleBackendRedisRequirements = "rht_backend_redis_requirements"
)

func CompareVersions(required, current string) (bool, error) {
	currentVersion, err := semver.ParseTolerant(current)
	if err != nil {
		return false, err
	}
	requiredVersion, err := semver.ParseTolerant(required)
	if err != nil {
		return false, err
	}

	if requiredVersion.LTE(currentVersion) {
		return true, nil
	}
	return false, nil
}

func fetchSecret(k8sclient client.Client, secretName, namespace string) (*v1.Secret, error) {
	secret := &v1.Secret{}

	err := k8sclient.Get(context.TODO(), client.ObjectKey{Name: secretName, Namespace: namespace}, secret)
	if err != nil {
		return secret, err
	}

	return secret, nil
}

// Check if backend Redis, system Redis or, system db are internal.
func InternalDatabases(apimInstance appsv1alpha1.APIManager, logger logr.Logger) (bool, bool, bool) {
	backendRedisInternal := false
	systemRedisInternal := false
	systemDatabaseInternal := false

	if !apimInstance.IsExternal(appsv1alpha1.BackendRedis) {
		logger.Info("Backend Redis database must be set to external in APIManager custom resource")
		backendRedisInternal = true
	}
	if !apimInstance.IsExternal(appsv1alpha1.SystemRedis) {
		logger.Info("System Redis database must be set to external in APIManager custom resource")
		systemRedisInternal = true
	}
	if !apimInstance.IsExternal(appsv1alpha1.SystemDatabase) {
		logger.Info("System Database must be set to external in APIManager custom resource")
		systemDatabaseInternal = true
	}

	return backendRedisInternal, systemRedisInternal, systemDatabaseInternal
}

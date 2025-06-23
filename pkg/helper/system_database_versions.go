package helper

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	systemDatabaseName = "system-database"
	postgresScheme     = "postgresql"
	mysqlScheme        = "mysql2"
)

func VerifySystemDatabase(k8sclient client.Client, reqConfigMap *v1.ConfigMap, apimInstance *appsv1alpha1.APIManager, logger logr.Logger) (bool, error) {
	databaseVersionVerified := false
	logger.Info("Verifying system database version")
	connSecret, err := fetchSecret(k8sclient, systemDatabaseName, apimInstance.Namespace)
	if err != nil {
		logger.Info("System database secret not found")
		return databaseVersionVerified, err
	}

	dbConfig := reconcileSystemDBSecret(*connSecret)
	if apimInstance.IsSystemDatabaseTLSEnabled() {
		dbConfig.TLS.Enabled = true
	}

	var databaseRequirement string

	if strings.HasPrefix(dbConfig.URL, "postgres://") || strings.HasPrefix(dbConfig.URL, "postgresql://") {
		databaseRequirement = reqConfigMap.Data[RHTThreescalePostgresRequirements]
		if databaseRequirement == "" {
			return true, nil
		}

		databaseVersionVerified, err = verifyPostgresVersion(dbConfig, databaseRequirement)
		if err != nil {
			logger.Info("Failed to verify Postgres database version", "err", err)
			return false, err
		}
	} else if strings.HasPrefix(dbConfig.URL, "mysql://") || strings.HasPrefix(dbConfig.URL, "mysql2://") {
		databaseRequirement = reqConfigMap.Data[RHTThreescaleMysqlRequirements]
		if databaseRequirement == "" {
			return true, nil
		}

		databaseVersionVerified, err = verifyMySQLVersion(dbConfig, databaseRequirement)
		if err != nil {
			logger.Info("Failed to verify MySQL database version", "err", err)
			return false, err
		}
	} else if strings.HasPrefix(dbConfig.URL, "oracle-enhanced://") {
		logger.Info("Oracle system database discovered, bypassing version check")
		return true, nil
	} else {
		return false, fmt.Errorf("unsupported database")
	}

	if databaseVersionVerified {
		logger.Info("System database version verified")
	} else {
		logger.Info("System database version not matching the required version")
	}

	return databaseVersionVerified, nil
}

package helper

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-logr/logr"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SystemDatabase struct {
	scheme   string
	host     string
	port     string
	password string
	user     string
	path     string
}

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

	// validate secret URL
	databaseUrl, bypass, err := systemDatabaseURLIsValid(string(connSecret.Data["URL"]))
	if err != nil {
		logger.Info("System database secret is invalid")
		return false, err
	}
	if bypass {
		logger.Info("Oracle system database discovered, bypassing version check")
		return true, nil
	}
	systemDatabase := databaseObject(databaseUrl)

	if systemDatabase.scheme == postgresScheme {
		postgresDatabaseRequirement := reqConfigMap.Data[RHTThreescalePostgresRequirements]
		if postgresDatabaseRequirement == "" {
			databaseVersionVerified = true
		} else {
			databaseVersionVerified, err = verifySystemPostgresDatabaseVersion(k8sclient, apimInstance.Namespace, postgresDatabaseRequirement, systemDatabase, logger)
			if err != nil {
				logger.Info("Encountered error during version verification of system Postgres")
				return false, err
			}
		}
	}

	if systemDatabase.scheme == mysqlScheme {
		mysqlDatabaseRequirement := reqConfigMap.Data[RHTThreescaleMysqlRequirements]
		if mysqlDatabaseRequirement == "" {
			databaseVersionVerified = true
		} else {
			databaseVersionVerified, err = verifySystemMysqlDatabaseVersion(k8sclient, apimInstance.Namespace, mysqlDatabaseRequirement, systemDatabase, logger)
			if err != nil {
				logger.Info("Encountered error during version verification of system MySQL")
				return false, err
			}
		}
	}

	if databaseVersionVerified {
		logger.Info("System database version verified")
	} else {
		logger.Info("System database version not matching the required version")
	}

	return databaseVersionVerified, nil
}

func databaseObject(url *url.URL) SystemDatabase {
	password, _ := url.User.Password()
	return SystemDatabase{
		scheme:   url.Scheme,
		host:     url.Hostname(),
		port:     url.Port(),
		password: password,
		user:     url.User.Username(),
		path:     url.Path,
	}
}

func verifySystemPostgresDatabaseVersion(k8sclient client.Client, namespace, requiredVersion string, databaseObject SystemDatabase, logger logr.Logger) (bool, error) {
	databasePod, err := CreateDatabaseThrowAwayPod(k8sclient, namespace, "postgres")
	if err != nil {
		logger.Info("Failed to create system database throwaway pod")
		return false, err
	}

	if databaseObject.port == "" {
		databaseObject.port = "5432"
	}

	postgresqlCommand := fmt.Sprintf("PGPASSWORD=\"%s\" psql -h \"%s\" -U \"%s\" -d \"%s\" -p\"%s\" -t -A -c \"SELECT version();\"", databaseObject.password, databaseObject.host, databaseObject.user, strings.TrimLeft(databaseObject.path, "/"), databaseObject.port)
	command := []string{"/bin/bash", "-c", postgresqlCommand}
	podExecutor := NewPodExecutor(logger)
	stdout, stderr, err := podExecutor.ExecuteRemoteCommand(databasePod.Namespace, databasePod.Name, command)
	if err != nil {
		return false, fmt.Errorf("Failed to execute command to retrieve database version. Error %s", err)
	}
	if stderr != "" {
		return false, fmt.Errorf("Failed to execute command to retrieve database version. Error %s", err)
	}

	currentPostgresVersion, err := retrievePostgresVersion(stdout)
	if err != nil {
		logger.Info("Failed to retrieve postgres version from the cli command")
		return false, err
	}

	requirementsMet := CompareVersions(requiredVersion, currentPostgresVersion)
	if requirementsMet {
		err := DeletePod(k8sclient, databasePod)
		if err != nil {
			return false, nil
		}
	}

	return requirementsMet, nil
}

func verifySystemMysqlDatabaseVersion(k8sclient client.Client, namespace, requiredVersion string, databaseObject SystemDatabase, logger logr.Logger) (bool, error) {
	databasePod, err := CreateDatabaseThrowAwayPod(k8sclient, namespace, "mysql")
	if err != nil {
		return false, err
	}

	mysqlCommand := fmt.Sprintf("mysql -sN -h \"%s\" -u \"%s\" -p\"%s\" -e \"SELECT VERSION();\"", databaseObject.host, databaseObject.user, databaseObject.password)
	command := []string{"/bin/bash", "-c", mysqlCommand}
	podExecutor := NewPodExecutor(logger)
	stdout, stderr, err := podExecutor.ExecuteRemoteCommand(databasePod.Namespace, databasePod.Name, command)
	if err != nil {
		return false, fmt.Errorf("Failed to execute command to retrieve database version. Error %s", err)
	}
	if stderr != "" {
		return false, fmt.Errorf("Failed to execute command to retrieve database version. Error %s", err)
	}

	currentMysqlVersion, err := retrieveMysqlVersion(stdout)
	if err != nil {
		logger.Info("Failed to retrieve postgres version from the cli command")
		return false, err
	}

	requirementsMet := CompareVersions(requiredVersion, currentMysqlVersion)

	if requirementsMet {
		err := DeletePod(k8sclient, databasePod)
		if err != nil {
			return false, nil
		}
	}

	return requirementsMet, nil
}

package helper

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5"

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
			databaseVersionVerified, err = verifySystemPostgresDatabaseVersion(postgresDatabaseRequirement, systemDatabase, logger)
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
			databaseVersionVerified, err = verifySystemMysqlDatabaseVersion(mysqlDatabaseRequirement, systemDatabase, logger)
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

func verifySystemPostgresDatabaseVersion(requiredVersion string, databaseObject SystemDatabase, logger logr.Logger) (bool, error) {
	var version string

	if databaseObject.port == "" {
		databaseObject.port = "5432"
	}

	dsn := fmt.Sprintf("%s://%s:%s@%s:%p/%s", databaseObject.scheme, databaseObject.user, databaseObject.password, databaseObject.host, &databaseObject.port, strings.TrimLeft(databaseObject.path, "/"))
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return false, fmt.Errorf("failed to connect to Postgres database. Error %s", err)
	}
	defer conn.Close(context.Background())

	err = conn.QueryRow(context.Background(), "SELECT version()").Scan(&version)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve Postgres database version. Error %s", err)
	}

	currentPostgresVersion, err := retrievePostgresVersion(version)
	if err != nil {
		logger.Info("failed to retrieve postgres version from the cli command")
		return false, err
	}

	requirementsMet := CompareVersions(requiredVersion, currentPostgresVersion)

	return requirementsMet, nil
}

func verifySystemMysqlDatabaseVersion(requiredVersion string, databaseObject SystemDatabase, logger logr.Logger) (bool, error) {
	var version string

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%p)", databaseObject.user, databaseObject.password, databaseObject.host, &databaseObject.port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, fmt.Errorf("failed to connect to MySQL database. Error %s", err)
	}

	err = db.QueryRowContext(context.Background(), "SELECT VERSION()").Scan(&version)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve MySQL database version. Error %s", err)
	}

	currentMysqlVersion, err := retrieveMysqlVersion(version)
	if err != nil {
		logger.Info("Failed to retrieve postgres version from the cli command")
		return false, err
	}

	requirementsMet := CompareVersions(requiredVersion, currentMysqlVersion)

	return requirementsMet, nil
}


package helper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-logr/logr"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	systemDatabaseName = "system-database"
)

func VerifySystemDatabase(k8sclient client.Client, reqConfigMap *v1.ConfigMap, apimInstance *appsv1alpha1.APIManager, logger logr.Logger) (bool, error) {
	databaseVersionVerified := false

	logger.Info("Verifying system database version")
	// In upgrade scenario, system database secret must always be present, if it's not, we are dealing with already broken installation.
	// In fresh installation with external databases, it must be present as well otherwise installation will not succeed.
	connSecret, err := FetchSecret(k8sclient, systemDatabaseName, apimInstance.Namespace)
	if err != nil {
		return databaseVersionVerified, err
	}

	// Retrieve database type from the URL connection string
	databaseType := detectDatabaseType(string(connSecret.Data["URL"]))
	// For now, only mysql and postgresql version are/can be confirmed.
	if databaseType == "oracle" || databaseType == "unknown" {
		// by-pass the check if the database type is unknown or is oracle.
		databaseVersionVerified = true
	}

	if databaseType == "postgres" {
		postgresDatabaseRequirement := reqConfigMap.Data[RHTThreescalePostgresRequirements]
		// If there are no requirements for postgres, and postgres is used, bypass the requirement check
		if postgresDatabaseRequirement == "" {
			databaseVersionVerified = true
		} else {
			databaseVersionVerified, err = verifySystemPostgresDatabaseVersion(k8sclient, *reqConfigMap, *connSecret, *apimInstance, postgresDatabaseRequirement, logger)
			if err != nil {
				return false, err
			}
		}
	}

	if databaseType == "mysql" {
		mysqlDatabaseRequirement := reqConfigMap.Data[RHTThreescaleMysqlRequirements]
		// If there are no requirements for mysql, and mysql is used, bypass the requirement check
		if mysqlDatabaseRequirement == "" {
			databaseVersionVerified = true
		} else {
			databaseVersionVerified, err = verifySystemMysqlDatabaseVersion(k8sclient, *reqConfigMap, *connSecret, *apimInstance, mysqlDatabaseRequirement, logger)
			if err != nil {
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

func verifySystemPostgresDatabaseVersion(k8sclient client.Client, configMap v1.ConfigMap, connSecret v1.Secret, instance appsv1alpha1.APIManager, requiredVersion string, logger logr.Logger) (bool, error) {
	databasePod, err := CreateDatabaseThrowAwayPod(k8sclient, "postgres", instance, connSecret)
	if err != nil {
		return false, err
	}

	postgresHost := retrieveSystemHost(connSecret)
	postgresUser := string(connSecret.Data["DB_USER"])
	postgresPassword := string(connSecret.Data["DB_PASSWORD"])
	postgresDatabaseName, err := retrieveDatabaseName(string(connSecret.Data["URL"]))
	if err != nil {
		return false, err
	}

	if postgresHost == "" || postgresPassword == "" || postgresUser == "" {
		return false, fmt.Errorf("error unable to read credentials from system-database secret")
	}

	postgresqlCommand := fmt.Sprintf("PGPASSWORD=\"%s\" psql -h \"%s\" -U \"%s\" -d \"%s\" -p\"5432\" -t -A -c \"SELECT version();\"", postgresPassword, postgresHost, postgresUser, postgresDatabaseName)
	command := []string{"/bin/bash", "-c", postgresqlCommand}
	podExecutor := NewPodExecutor(logger)
	stdout, stderr, err := podExecutor.ExecuteRemoteCommand(databasePod.Namespace, databasePod.Name, command)
	if err != nil {
		return false, fmt.Errorf("failed to confirm database version")
	}
	if stderr != "" {
		return false, fmt.Errorf("error when executing pod exec command to retrieve database version")
	}

	var currentPostgresVersion string

	if stdout != "" {
		pattern := `PostgreSQL (\d+(\.\d+)*)`
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(stdout)
		if len(match) > 1 {
			// The version number is captured by the first group
			currentPostgresVersion = match[1]
		} else {
			return false, fmt.Errorf("redis version not found in stdout")
		}
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

func detectDatabaseType(secret string) string {
	if strings.HasPrefix(secret, "mysql") {
		return "mysql"
	} else if strings.HasPrefix(secret, "postgres") {
		return "postgres"
	} else if strings.HasPrefix(secret, "oracle") {
		return "oracle"
	} else {
		return "unknown"
	}
}

func retrieveSystemHost(secret v1.Secret) string {
	var systemhost string
	re := regexp.MustCompile(`@(.*?)(:\d+)?/`)
	matches := re.FindStringSubmatch(string(secret.Data["URL"]))
	if len(matches) > 1 {
		systemhost = matches[1]
	}

	return systemhost
}

func retrieveDatabaseName(connectionString string) (string, error) {
	regexPattern := `postgresql://\w+:\w+@[\w.-]+:\d+/(?P<dbname>.+)`
	regex := regexp.MustCompile(regexPattern)
	match := regex.FindStringSubmatch(connectionString)

	if len(match) < 2 {
		return "", fmt.Errorf("unable to extract dbname from the connection string")
	}

	// The dbname will be in the captured group named "dbname"
	dbname := match[regex.SubexpIndex("dbname")]
	return dbname, nil
}

func verifySystemMysqlDatabaseVersion(k8sclient client.Client, configMap v1.ConfigMap, connSecret v1.Secret, instance appsv1alpha1.APIManager, requiredVersion string, logger logr.Logger) (bool, error) {
	databasePod, err := CreateDatabaseThrowAwayPod(k8sclient, "mysql", instance, connSecret)
	if err != nil {
		return false, err
	}

	mysqlHost := retrieveSystemHost(connSecret)
	password := string(connSecret.Data["DB_PASSWORD"])
	user := string(connSecret.Data["DB_USER"])
	if mysqlHost == "" || password == "" || user == "" {
		return false, fmt.Errorf("error unable to read credentials from system-database secret")
	}

	mysqlCommand := fmt.Sprintf("mysql -sN -h \"%s\" -u \"%s\" -p\"%s\" -e \"SELECT VERSION();\"", mysqlHost, user, password)
	command := []string{"/bin/bash", "-c", mysqlCommand}
	podExecutor := NewPodExecutor(logger)
	stdout, stderr, err := podExecutor.ExecuteRemoteCommand(databasePod.Namespace, databasePod.Name, command)
	if err != nil {
		return false, fmt.Errorf("failed to confirm database version")
	}
	if stderr != "" {
		return false, fmt.Errorf("error when executing pod exec command to retrieve database version")
	}

	var currentMysqlVersion string

	if stdout != "" {
		pattern := `[0-9]+\.[0-9]+\.[0-9]+`
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(stdout)
		if len(match) > 0 {
			// The version number is captured by the first group
			currentMysqlVersion = match[0]
		} else {
			return false, fmt.Errorf("redis version not found in stdout")
		}
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

func InternalDatabases(apimInstance appsv1alpha1.APIManager, logger logr.Logger) (bool, bool, bool) {
	backendRedisVerified := false
	systemRedisVerified := false
	systemDatabaseVerified := false
	if !apimInstance.IsExternal(appsv1alpha1.BackendRedis) {
		logger.Info("Backend Redis requirements confirmed")
		backendRedisVerified = true
	}
	if !apimInstance.IsExternal(appsv1alpha1.SystemRedis) {
		logger.Info("System Redis requirements confirmed")
		systemRedisVerified = true
	}
	if !apimInstance.IsExternal(appsv1alpha1.SystemDatabase) {
		logger.Info("System Database requirements confirmed")
		systemDatabaseVerified = true
	}

	return backendRedisVerified, systemRedisVerified, systemDatabaseVerified
}

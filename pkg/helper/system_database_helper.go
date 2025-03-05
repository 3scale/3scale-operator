package helper

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"regexp"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	pgxstd "github.com/jackc/pgx/v5/stdlib"
	v1 "k8s.io/api/core/v1"
)

const (
	systemDatabaseURL = "URL"
	secretName        = "system-database"
)

type DatabaseConfig struct {
	URL string
}

func reconcileSystemDBSecret(secret v1.Secret) *DatabaseConfig {
	return &DatabaseConfig{
		URL: string(secret.Data[systemDatabaseURL]),
	}
}

func verifyMySQLVersion(cfg *DatabaseConfig, requiredVersion string) (bool, error) {
	url, err := url.Parse(cfg.URL)
	if err != nil {
		return false, err
	}
	password, _ := url.User.Password()
	port := url.Port()
	if port == "" {
		port = "3306"
	}

	dbConfig := mysql.NewConfig()

	dbConfig.Net = "tcp"
	dbConfig.Addr = net.JoinHostPort(url.Hostname(), port)
	dbConfig.User = url.User.Username()
	dbConfig.Passwd = password

	// Append params
	params := make(map[string]string)
	q := url.Query()
	for k := range q {
		params[k] = q.Get(k)
	}
	dbConfig.Params = params

	connector, err := mysql.NewConnector(dbConfig)
	if err != nil {
		return false, err
	}

	db := sql.OpenDB(connector)

	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve database version. Error %s", err)
	}

	databaseCurrentVersion, err := retrieveMysqlVersion(version)
	if err != nil {
		return false, err
	}

	return CompareVersions(requiredVersion, databaseCurrentVersion)
}

func verifyPostgresVersion(cfg *DatabaseConfig, requiredVersion string) (bool, error) {
	dbConfig, err := pgx.ParseConfig(cfg.URL)
	if err != nil {
		return false, err
	}

	db := pgxstd.OpenDB(*dbConfig)
	var version string

	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve database version. Error %s", err)
	}

	databaseCurrentVersion, err := retrievePostgresVersion(version)
	if err != nil {
		return false, err
	}

	return CompareVersions(requiredVersion, databaseCurrentVersion)
}

func retrievePostgresVersion(stdout string) (string, error) {
	currentPostgresVersion := ""
	pattern := `PostgreSQL (\d+(\.\d+)*)`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(stdout)
	if len(match) > 1 {
		currentPostgresVersion = match[1]
	} else {
		return currentPostgresVersion, fmt.Errorf("postgres version not found in stdout")
	}

	return currentPostgresVersion, nil
}

func retrieveMysqlVersion(stdout string) (string, error) {
	currentMysqlVersion := ""
	pattern := `[0-9]+\.[0-9]+\.[0-9]+`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(stdout)
	if len(match) > 0 {
		// The version number is captured by the first group
		currentMysqlVersion = match[0]
	} else {
		return currentMysqlVersion, fmt.Errorf("redis version not found in stdout")
	}

	return currentMysqlVersion, nil
}

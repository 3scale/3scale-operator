package helper

import (
	"fmt"
	"net/url"
	"regexp"
)

const (
	urlKey     = "URL"
	secretName = "system-database"
)

func systemDatabaseURLIsValid(rawURL string) (*url.URL, error) {
	resultURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("'%s' field of '%s' secret must have 'scheme://user:password@host/path' format", urlKey, secretName)
	}
	if resultURL.Scheme != "mysql2" {
		return nil, fmt.Errorf("'%s' field of '%s' secret must contain 'mysql2' as the scheme part", urlKey, secretName)
	}

	if resultURL.User == nil {
		return nil, fmt.Errorf("authentication information in '%s' field of '%s' secret must be provided", urlKey, secretName)
	}
	if resultURL.User.Username() == "" {
		return nil, fmt.Errorf("authentication information in '%s' field of '%s' secret must contain a username", urlKey, secretName)
	}
	if resultURL.User.Username() != "root" {
		return nil, fmt.Errorf("authentication information in '%s' field of '%s' secret must contain 'root' as the username", urlKey, secretName)
	}
	if _, set := resultURL.User.Password(); !set {
		return nil, fmt.Errorf("authentication information in '%s' field of '%s' secret must contain a password", urlKey, secretName)
	}

	if resultURL.Host == "" {
		return nil, fmt.Errorf("host information in '%s' field of '%s' secret must be provided", urlKey, secretName)
	}
	if resultURL.Path == "" {
		return nil, fmt.Errorf("database name in '%s' field of '%s' secret must be provided", urlKey, secretName)
	}

	return resultURL, nil
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

package helper

import (
	"fmt"
	"github.com/go-logr/logr"
	"regexp"
	"strings"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifySystemRedis(k8sclient client.Client, reqConfigMap *v1.ConfigMap, apimInstance *appsv1alpha1.APIManager, logger logr.Logger) (bool, error) {
	logger.Info("Verifying system redis version")
	systemRedisVerified := false
	systemRedisRequirement := reqConfigMap.Data[RHTThreescaleSystemRedisRequirements]
	connSecret, err := FetchSecret(k8sclient, "system-redis", apimInstance.Namespace)
	if err != nil {
		return false, err
	}

	systemRedisVerified, err = verifySystemRedisVersion(k8sclient, *connSecret, apimInstance.Namespace, systemRedisRequirement, logger)
	if err != nil {
		return false, err
	}
	if systemRedisVerified {
		logger.Info("System redis version verified")
	} else {
		logger.Info("System redis version not matching the required version")
	}

	return systemRedisVerified, nil
}

func VerifyBackendRedis(k8sclient client.Client, reqConfigMap *v1.ConfigMap, apimInstance *appsv1alpha1.APIManager, logger logr.Logger) (bool, error) {
	logger.Info("Verifying backend redis version")
	backendRedisVerified := false
	backendRedisRequirement := reqConfigMap.Data[RHTThreescaleBackendRedisRequirements]
	connSecret, err := FetchSecret(k8sclient, "backend-redis", apimInstance.Namespace)
	if err != nil {
		return false, err
	}

	backendRedisVerified, err = verifyBackendRedisVersion(k8sclient, *connSecret, apimInstance.Namespace, backendRedisRequirement, logger)
	if err != nil {
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
	redisPod, err := CreateRedisThrowAwayPod(k8sclient, namespace)
	if err != nil {
		return false, err
	}

	var redisHost, redisPort string

	if string(connSecret.Data["SENTINEL_HOSTS"]) != "" {
		sentinelPassword, sentinelHost, sentinelPort := retrieveSentinelRedisHostPortAndPassword(string(connSecret.Data["SENTINEL_HOSTS"]))
		redisGroup := retrieveSentinelRedisGroup(string(connSecret.Data["URL"]))
		redisCliCommand := ""
		if sentinelPassword != "" {
			redisCliCommand = retrieveRedisSentinelCommand(sentinelHost, sentinelPort, sentinelPassword, redisGroup)
		} else {
			redisCliCommand = retrieveRedisSentinelCommand(sentinelHost, sentinelPort, "", redisGroup)
		}

		stdout, err := executeRedisCliCommand(*redisPod, redisCliCommand, logger)
		if err != nil {
			return false, err
		}
		redisHost, redisPort = parseRedisSentinelResponse(stdout)
	} else {
		redisHost, redisPort = retrieveRedisHostAndPort(connSecret, "URL")
	}

	redisInstancePass := extractPassword(string(connSecret.Data["URL"]))
	redisCliCommand := retrieveRedisCliCommand(redisHost, redisPort, redisInstancePass)

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
	redisPod, err := CreateRedisThrowAwayPod(k8sclient, namespace)
	if err != nil {
		return false, err
	}

	var redisHost, redisPort string

	if string(connSecret.Data["REDIS_QUEUES_SENTINEL_HOSTS"]) != "" {
		sentinelPassword, sentinelHost, sentinelPort := retrieveSentinelRedisHostPortAndPassword(string(connSecret.Data["REDIS_QUEUES_SENTINEL_HOSTS"]))
		redisGroup := retrieveSentinelRedisGroup(string(connSecret.Data["REDIS_QUEUES_URL"]))
		redisCliCommand := ""
		if sentinelPassword != "" {
			redisCliCommand = retrieveRedisSentinelCommand(sentinelHost, sentinelPort, sentinelPassword, redisGroup)
		} else {
			redisCliCommand = retrieveRedisSentinelCommand(sentinelHost, sentinelPort, "", redisGroup)
		}

		stdout, err := executeRedisCliCommand(*redisPod, redisCliCommand, logger)
		if err != nil {
			return false, err
		}
		redisHost, redisPort = parseRedisSentinelResponse(stdout)
	} else {
		redisHost, redisPort = retrieveRedisHostAndPort(connSecret, "REDIS_QUEUES_URL")
	}

	redisInstancePass := extractPassword(string(connSecret.Data["REDIS_QUEUES_URL"]))
	redisCliCommand := retrieveRedisCliCommand(redisHost, redisPort, redisInstancePass)
	stdout, err := executeRedisCliCommand(*redisPod, redisCliCommand, logger)
	if err != nil {
		return false, err
	}

	currentRedisVersion, err := retrieveCurrentVersionOfRedis(stdout)
	if err != nil {
		return false, err
	}

	redisQueuesVersionConfirmed := CompareVersions(requiredVersion, currentRedisVersion)

	if string(connSecret.Data["REDIS_STORAGE_SENTINEL_HOSTS"]) != "" {
		sentinelPassword, sentinelHost, sentinelPort := retrieveSentinelRedisHostPortAndPassword(string(connSecret.Data["REDIS_STORAGE_SENTINEL_HOSTS"]))
		redisGroup := retrieveSentinelRedisGroup(string(connSecret.Data["REDIS_STORAGE_URL"]))
		redisCliCommand := ""
		if sentinelPassword != "" {
			redisCliCommand = retrieveRedisSentinelCommand(sentinelHost, sentinelPort, sentinelPassword, redisGroup)
		} else {
			redisCliCommand = retrieveRedisSentinelCommand(sentinelHost, sentinelPort, "", redisGroup)
		}

		stdout, err := executeRedisCliCommand(*redisPod, redisCliCommand, logger)
		if err != nil {
			return false, err
		}
		redisHost, redisPort = parseRedisSentinelResponse(stdout)
	} else {
		redisHost, redisPort = retrieveRedisHostAndPort(connSecret, "REDIS_STORAGE_URL")
	}

	redisInstancePass = extractPassword(string(connSecret.Data["REDIS_STORAGE_URL"]))
	redisCliCommand = retrieveRedisCliCommand(redisHost, redisPort, redisInstancePass)
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

func extractPassword(input string) string {
	pattern := `^redis://(?:[^:@]*:)?([^@]*)@`
	re := regexp.MustCompile(pattern)

	// Find the matches
	matches := re.FindStringSubmatch(input)
	var password string
	if len(matches) > 1 {
		password = matches[1]
	}

	return password
}

func retrieveSentinelRedisGroup(input string) string {
	// Split the input string by "//" and "/"
	parts := strings.Split(input, "/")

	// Check if there are at least two parts (indicating presence of "redis://")
	if len(parts) >= 2 {
		// Extract the part just before the first "/"
		mymasterPart := parts[len(parts)-2]

		// Split the mymaster part by "@" (if present) and return the last segment
		mymasterSegments := strings.Split(mymasterPart, "@")
		return mymasterSegments[len(mymasterSegments)-1]
	}

	// If no match found, return empty string
	return ""
}

func retrieveCurrentVersionOfRedis(stdString string) (string, error) {
	var currentRedisVersion string

	if stdString != "" {
		pattern := `redis_version:(\d+\.\d+\.\d+)`
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(stdString)
		if len(match) >= 2 {
			currentRedisVersion = match[1]
		} else {
			return "", fmt.Errorf("redis version not found in stdout")
		}
	}

	return currentRedisVersion, nil
}

func retrieveRedisSentinelCommand(host, port, password, redisGroup string) string {
	if password != "" {
		return fmt.Sprintf("redis-cli -h %s -p %s -a %s sentinel get-master-addr-by-name %s", host, port, password, redisGroup)
	} else {
		return fmt.Sprintf("redis-cli -h %s -p %s sentinel get-master-addr-by-name %s", host, port, redisGroup)
	}
}

func retrieveRedisCliCommand(host, port, redisInstancePass string) string {
	if redisInstancePass != "" {
		return fmt.Sprintf("redis-cli -h %s -p %s -a %s INFO SERVER | grep redis_version", host, port, redisInstancePass)
	} else {
		return fmt.Sprintf("redis-cli -h %s -p %s INFO SERVER | grep redis_version", host, port)
	}
}

func executeRedisCliCommand(pod v1.Pod, command string, logger logr.Logger) (string, error) {
	redisCommand := []string{"/bin/bash", "-c", command}
	podExecutor := NewPodExecutor(logger)
	stdout, stderr, err := podExecutor.ExecuteRemoteCommand(pod.Namespace, pod.Name, redisCommand)
	if err != nil {
		return stdout, err
	}
	if stderr != "" {
		return "", fmt.Errorf("error when executing pod exec command against redis")
	}

	return stdout, nil
}

func retrieveRedisHostAndPort(secret v1.Secret, secretDataString string) (string, string) {
	var host string
	var port string
	re := regexp.MustCompile(`redis://([^:]+):(\d+)/`)

	matches := re.FindStringSubmatch(string(secret.Data[secretDataString]))

	if len(matches) == 3 {
		host = matches[1]
		port = matches[2]
	} else {
		fmt.Println("URL doesn't match expected pattern.")
	}

	return host, port
}

func retrieveSentinelRedisHostPortAndPassword(secretDataString string) (string, string, string) {
	sentinelHost, sentinelPort, sentinelPassword := retrieveFirstSentinelHost(string(secretDataString))
	return sentinelHost, sentinelPort, sentinelPassword
}

func retrieveFirstSentinelHost(hosts string) (string, string, string) {
	addresses := strings.Split(hosts, ",")
	if len(addresses) == 0 {
		return "", "", ""
	}

	for i := range addresses {
		addresses[i] = strings.TrimSpace(addresses[i])
	}

	parts := strings.SplitN(addresses[0], "://", 2)
	if len(parts) != 2 {
		return "", "", ""
	}

	hostParts := strings.Split(parts[1], "@")
	var password string
	if len(hostParts) == 1 {
		hostParts = append([]string{""}, hostParts[0])
	} else {
		credentials := strings.Split(hostParts[0], ":")
		if len(credentials) != 2 {
			return "", "", ""
		}
		password = credentials[1]
	}

	hostPort := strings.Split(hostParts[len(hostParts)-1], ":")
	if len(hostPort) != 2 {
		return "", "", ""
	}

	return password, hostPort[0], hostPort[1]
}

func parseRedisSentinelResponse(response string) (string, string) {
	lines := strings.Split(strings.TrimSpace(response), "\r\n")

	var ipAddress, port string

	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) >= 2 {
			switch parts[0] {
			case "1)":
				ipAddress = strings.Trim(parts[1], "\"")
			case "2)":
				port = strings.Trim(parts[1], "\"")
			}
		}
	}

	return ipAddress, port
}

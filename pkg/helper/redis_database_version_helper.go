package helper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
)

func reconcileSystemRedisSecret(secret v1.Secret) Redis {
	redis := Redis{}

	sentinelHosts, url := retrieveSystemHostAndUrl(secret)
	if sentinelHosts != "" {
		redis.sentinelHost, redis.sentinelPort, redis.sentinelPassword = retrieveSentinelRedisHostPortAndPassword(sentinelHosts)
		redis.sentinelGroup = retrieveSentinelRedisGroup(url)
		if redis.sentinelPort == "" {
			redis.sentinelPort = "6379"
		}
	}
	redis.redisHost, redis.redisPort, redis.redisPassword = retrieveRedisHostPortPassword(url)
	if redis.redisPort == "" {
		redis.redisPort = "6379"
	}

	return redis
}

func reconcileQueuesRedisSecret(secret v1.Secret) Redis {
	redis := Redis{}

	sentinelHosts, url := retrieveQueuesHostAndUrl(secret)
	if sentinelHosts != "" {
		redis.sentinelHost, redis.sentinelPort, redis.sentinelPassword = retrieveSentinelRedisHostPortAndPassword(sentinelHosts)
		redis.sentinelGroup = retrieveSentinelRedisGroup(url)
		if redis.sentinelPort == "" {
			redis.sentinelPort = "6379"
		}
	}
	redis.redisHost, redis.redisPort, redis.redisPassword = retrieveRedisHostPortPassword(url)
	if redis.redisPort == "" {
		redis.redisPort = "6379"
	}

	return redis
}

func reconcileStorageRedisSecret(secret v1.Secret) Redis {
	redis := Redis{}

	sentinelHosts, url := retrieveStorageHostAndUrl(secret)
	if sentinelHosts != "" {
		redis.sentinelHost, redis.sentinelPort, redis.sentinelPassword = retrieveSentinelRedisHostPortAndPassword(sentinelHosts)
		redis.sentinelGroup = retrieveSentinelRedisGroup(url)
		if redis.sentinelPort == "" {
			redis.sentinelPort = "6379"
		}
	}
	redis.redisHost, redis.redisPort, redis.redisPassword = retrieveRedisHostPortPassword(url)
	if redis.redisPort == "" {
		redis.redisPort = "6379"
	}

	return redis
}

func retrieveQueuesHostAndUrl(secret v1.Secret) (string, string) {
	return string(secret.Data[backendRedisQueuesSentinelHosts]), string(secret.Data[backendRedisQueuesURL])
}

func retrieveStorageHostAndUrl(secret v1.Secret) (string, string) {
	return string(secret.Data[backendRedisStorageSentinelHosts]), string(secret.Data[backendRedisStorageURL])
}

func retrieveSystemHostAndUrl(secret v1.Secret) (string, string) {
	return string(secret.Data[systemRedisSentinelHosts]), string(secret.Data[systemRedisUrl])
}

func extractPassword(input string) string {
	pattern := `^redis://(?:[^:@]*:)?([^@]*)@`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(input)
	var password string
	if len(matches) > 1 {
		password = matches[1]
	}

	return password
}

func retrieveSentinelRedisGroup(input string) string {
	if strings.HasPrefix(input, "redis://") {
		input = strings.TrimPrefix(input, "redis://")
		if strings.Contains(input, "@") {
			input = input[strings.Index(input, "@")+1:]
		}
		parts := strings.Split(input, "/")
		hostParts := strings.Split(parts[0], ":")
		return hostParts[0]
	}

	if strings.Contains(input, ":") && strings.Contains(input, "@") {
		input = input[strings.Index(input, "@")+1:]
	}

	parts := strings.Split(input, "/")
	hostParts := strings.Split(parts[0], ":")
	return hostParts[0]
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

func retrieveRedisHostPortPassword(input string) (string, string, string) {
	var host string
	var port string
	var password string
	re := regexp.MustCompile(`^(?:redis://)?(?::([^@]+)@)?([^:/]+)(?::(\d+))?`)

	matches := re.FindStringSubmatch(input)

	if len(matches) >= 3 {
		password = matches[1]
		host = matches[2]
	}
	if len(matches) >= 4 {
		port = matches[3]
	}

	return host, port, password
}

func retrieveSentinelRedisHostPortAndPassword(secretDataString string) (string, string, string) {
	sentinelHost, sentinelPort, sentinelPassword := retrieveFirstSentinelHostPortPassword(string(secretDataString))
	return sentinelHost, sentinelPort, sentinelPassword
}

func retrieveFirstSentinelHostPortPassword(hosts string) (string, string, string) {
	addresses := strings.Split(hosts, ",")
	if len(addresses) == 0 {
		return "", "", ""
	}

	for i := range addresses {
		addresses[i] = strings.TrimSpace(addresses[i])
	}

	var hostParts []string
	if strings.Contains(addresses[0], "://") {
		parts := strings.SplitN(addresses[0], "://", 2)
		if len(parts) != 2 {
			return "", "", ""
		}
		hostParts = strings.Split(parts[1], "@")
	} else {
		hostParts = strings.Split(addresses[0], "@")
	}

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
	if len(hostPort) == 2 {
		return hostPort[0], hostPort[1], password
	} else if len(hostPort) == 1 {
		return hostPort[0], "", password
	} else {
		return "", "", password
	}
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

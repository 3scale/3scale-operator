package helper

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	goredis "github.com/redis/go-redis/v9"
	v1 "k8s.io/api/core/v1"
)

type Redis struct {
	sentinelOptions goredis.Options
	sentinelGroup   string
	redisOptions    goredis.Options
}

func reconcileSystemRedisSecret(secret v1.Secret) (Redis, error) {
	sentinelHosts, url := retrieveSystemHostAndUrl(secret)
	return parseRedisURL(sentinelHosts, url)
}

func reconcileQueuesRedisSecret(secret v1.Secret) (Redis, error) {
	sentinelHosts, url := retrieveQueuesHostAndUrl(secret)
	return parseRedisURL(sentinelHosts, url)
}

func reconcileStorageRedisSecret(secret v1.Secret) (Redis, error) {
	sentinelHosts, url := retrieveStorageHostAndUrl(secret)
	return parseRedisURL(sentinelHosts, url)
}

func parseRedisURL(sentinelHosts, url string) (Redis, error) {
	redisOpts := Redis{}
	if sentinelHosts != "" {
		redisOpts.sentinelOptions.Addr, redisOpts.sentinelOptions.Password = retrieveSentinelRedisHostPortAndPassword(sentinelHosts)
		redisOpts.sentinelGroup = retrieveSentinelRedisGroup(url)
	}
	opts, err := goredis.ParseURL(url)
	if err != nil {
		return redisOpts, err
	}
	redisOpts.redisOptions.Addr = opts.Addr
	redisOpts.redisOptions.Password = opts.Password

	return redisOpts, nil
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

func retrieveSentinelRedisHostPortAndPassword(secretDataString string) (string, string) {
	sentinelHost, sentinelPort, sentinelPassword := retrieveFirstSentinelHostPortPassword(string(secretDataString))
	if sentinelPort == "" {
		sentinelPort = "6379"
	}
	return fmt.Sprintf("%s:%s", sentinelHost, sentinelPort), sentinelPassword
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

func verifyRedisVersion(redisOpts Redis, requiredVersion string) (bool, error) {
	if redisOpts.sentinelOptions.Addr != "" {
		sentinelClient := goredis.NewSentinelClient(&redisOpts.sentinelOptions)
		addr, err := sentinelClient.GetMasterAddrByName(context.Background(), redisOpts.sentinelGroup).Result()
		if err != nil {
			return false, fmt.Errorf("failed to execute command to retrieve the Redis sentinal master address - error: %s", err)
		}
		host, port := parseRedisSentinelResponse(addr[0])
		redisOpts.redisOptions.Addr = fmt.Sprintf("%s:%s", host, port)
	}

	rdb := goredis.NewClient(&redisOpts.redisOptions)

	info, err := rdb.Info(context.Background(), "server").Result()

	if err != nil {
		return false, fmt.Errorf("failed to execute command to retrieve the Redis version - error: %w", err)
	}

	currentRedisVersion, err := retrieveCurrentVersionOfRedis(info)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve current version of system Redis from the cli command - error: %w", err)
	}

	redisVersionConfirmed := CompareVersions(requiredVersion, currentRedisVersion)

	return redisVersionConfirmed, nil
}

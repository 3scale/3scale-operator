package helper

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	v1 "k8s.io/api/core/v1"
)

const (
	// Timeout for Read operations.
	//  it's just for sanity.
	defaultReadTimeout = 3 * time.Second
	// Timeout for Write operations.
	defaultWriteTimeout = 3 * time.Second
	// Timeout before killing Idle connections in the pool.
	defaultIdleTimeout = 3 * time.Minute
)

type RedisSecretKey struct {
	SentinelURL      string
	SentinelUsername string
	SentinelPassword string
	URL              string
	Username         string
	Password         string
	CACertificate    string
	Cert             string
	Key              string
}

// Convert secret values to RedisConfig
func reconcileRedisSecret(secret v1.Secret, config RedisSecretKey) *RedisConfig {
	return &RedisConfig{
		SentinelURL:      string(secret.Data[config.SentinelURL]),
		SentinelUsername: string(secret.Data[config.SentinelUsername]),
		SentinelPassword: string(secret.Data[config.SentinelPassword]),
		URL:              string(secret.Data[config.URL]),
		Username:         string(secret.Data[config.Username]),
		Password:         string(secret.Data[config.Password]),
		TLS: &TLSConfig{
			CACertificate: string(secret.Data[config.CACertificate]),
			Certificate:   string(secret.Data[config.Cert]),
			Key:           string(secret.Data[config.Key]),
		},
	}
}

func reconcileSystemRedisSecret(secret v1.Secret) *RedisConfig {
	config := RedisSecretKey{
		SentinelURL:      systemRedisSentinelHosts,
		SentinelUsername: systemRedisSentinelUsername,
		SentinelPassword: systemRedisSentinelPassword,
		URL:              systemRedisUrl,
		Username:         systemRedisUsername,
		Password:         systemRedisPassword,
		CACertificate:    systemRedisCA,
		Cert:             systemRedisCertificate,
		Key:              systemRedisKey,
	}
	return reconcileRedisSecret(secret, config)
}

func reconcileStorageRedisSecret(secret v1.Secret) *RedisConfig {
	config := RedisSecretKey{
		SentinelURL:      backendRedisStorageSentinelHosts,
		SentinelUsername: backendRedisStorageSentinelUsername,
		SentinelPassword: backendRedisStorageSentinelPassword,
		URL:              backendRedisStorageURL,
		Username:         backendRedisStorageUsername,
		Password:         backendRedisStoragePassword,
		CACertificate:    backendRedisStorageCA,
		Cert:             backendRedisStorageCertificate,
		Key:              backendRedisStorageKey,
	}
	return reconcileRedisSecret(secret, config)
}

func reconcileQueuesRedisSecret(secret v1.Secret) *RedisConfig {
	config := RedisSecretKey{
		SentinelURL:      backendRedisQueuesSentinelHosts,
		SentinelUsername: backendRedisQueuesSentinelUsername,
		SentinelPassword: backendRedisQueuesSentinelPassword,
		URL:              backendRedisQueuesURL,
		Username:         backendRedisQueuesUsername,
		Password:         backendRedisQueuesPassword,
		CACertificate:    backendRedisQueuesCA,
		Cert:             backendRedisQueuesCertificate,
		Key:              backendRedisQueuesKey,
	}
	return reconcileRedisSecret(secret, config)
}

type RedisConfig struct {
	URL              string
	Username         string
	Password         string
	SentinelURL      string
	SentinelMaster   string
	SentinelUsername string
	SentinelPassword string
	TLS              *TLSConfig
}

func Configure(cfg *RedisConfig) (*goredis.Client, error) {
	if cfg == nil {
		return nil, nil
	}

	var rdb *goredis.Client
	var err error

	if cfg.SentinelURL != "" {
		rdb, err = configureRedisSentinel(cfg)
	} else {
		rdb, err = configureRedis(cfg)
	}
	return rdb, err
}

func configureRedis(cfg *RedisConfig) (*goredis.Client, error) {
	opts, err := goredis.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	opts.ReadTimeout = defaultReadTimeout
	opts.WriteTimeout = defaultWriteTimeout
	opts.ConnMaxIdleTime = defaultIdleTimeout

	if cfg.Username != "" {
		opts.Username = cfg.Username
	}

	if cfg.Password != "" {
		opts.Password = cfg.Password
	}

	if cfg.TLS != nil && cfg.TLS.Enabled {
		tlsConfig, err := LoadCerts(cfg.TLS)

		if err != nil {
			return nil, err
		}

		opts.TLSConfig = tlsConfig
	}

	return goredis.NewClient(opts), nil
}

func configureRedisSentinel(cfg *RedisConfig) (*goredis.Client, error) {
	opts, err := sentinelOptions(cfg)
	if err != nil {
		return nil, err
	}

	client := goredis.NewFailoverClient(opts)

	return client, nil
}

func sentinelOptions(cfg *RedisConfig) (*goredis.FailoverOptions, error) {
	master_url, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}

	username := master_url.User.Username()
	password, _ := master_url.User.Password()

	urls := strings.Split(cfg.SentinelURL, ",")
	if len(urls) == 0 {
		return nil, fmt.Errorf("invalid sentinel URLs")
	}

	sentinels := make([]string, len(urls))

	// 3scale system does not support username/password in the sentinel URL.
	for i := range urls {
		url := strings.TrimSpace(urls[i])
		opt, err := goredis.ParseURL(url)
		if err != nil {
			return nil, err
		}

		sentinels[i] = opt.Addr
	}

	opts := &goredis.FailoverOptions{
		MasterName:       master_url.Hostname(),
		SentinelAddrs:    sentinels,
		Username:         username,
		Password:         password,
		SentinelUsername: cfg.SentinelUsername,
		SentinelPassword: cfg.SentinelPassword,
		ConnMaxIdleTime:  defaultIdleTimeout,
		ReadTimeout:      defaultReadTimeout,
		WriteTimeout:     defaultWriteTimeout,
	}

	if cfg.TLS != nil && cfg.TLS.Enabled {
		var tlsConfig *tls.Config

		tlsConfig, err := LoadCerts(cfg.TLS)
		if err != nil {
			return nil, err
		}

		opts.TLSConfig = tlsConfig
	}

	return opts, nil
}

func verifyRedisVersion(client *goredis.Client, requiredVersion string) (bool, error) {
	info, err := client.Info(context.Background(), "server").Result()

	if err != nil {
		return false, fmt.Errorf("failed to execute command to retrieve the Redis version - error: %w", err)
	}

	currentRedisVersion, err := retrieveCurrentVersionOfRedis(info)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve current version of system Redis from the cli command - error: %w", err)
	}

	redisVersionConfirmed, err := CompareVersions(requiredVersion, currentRedisVersion)
	if err != nil {
		return false, err
	}

	return redisVersionConfirmed, nil
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

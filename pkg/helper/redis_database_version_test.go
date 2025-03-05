package helper

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	caCert   = "./testdata/rootCA.pem"
	certFile = "./testdata/client.crt"
	keyFile  = "./testdata/client.key"
)

type mockRedisServer struct {
	clientConnected bool
}

func (m *mockRedisServer) Listen(t *testing.T) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")

	require.NoError(t, err)

	go func() {
		defer ln.Close()
		conn, err := ln.Accept()
		assert.NoError(t, err)
		m.clientConnected = true
		conn.Write([]byte("OK\n"))
	}()

	return ln.Addr().String()
}

func (m *mockRedisServer) Connected() bool {
	return m.clientConnected
}

// Secret reconcile test
func TestReconcileRedisSecrets(t *testing.T) {

	testCases := []struct {
		testName            string
		redisSecretFunction func(v1.Secret) *RedisConfig
		secret              v1.Secret
		redisConfig         *RedisConfig
	}{
		{
			"system-redis",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte("redis://username:password@sentinel-1:5000, redis://username:password@sentinel-2:5000, redis://username:password@sentinel-3:5000"),
					"URL":            []byte("redis://username:password@my-redis:5000/1"),
				},
			},
			&RedisConfig{
				URL:         "redis://username:password@my-redis:5000/1",
				SentinelURL: "redis://username:password@sentinel-1:5000, redis://username:password@sentinel-2:5000, redis://username:password@sentinel-3:5000",
			},
		},
		{
			"backend-redis (queues)",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte("redis://username:password@sentinel-1:5000, redis://username:password@sentinel-2:5000, redis://username:password@sentinel-3:5000"),
					"REDIS_QUEUES_URL":            []byte("redis://username:password@my-redis:5000/1"),
				},
			},
			&RedisConfig{
				URL:         "redis://username:password@my-redis:5000/1",
				SentinelURL: "redis://username:password@sentinel-1:5000, redis://username:password@sentinel-2:5000, redis://username:password@sentinel-3:5000",
			},
		},
		{
			"backend-redis (storage)",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte("redis://username:password@sentinel-1:5000, redis://username:password@sentinel-2:5000, redis://username:password@sentinel-3:5000"),
					"REDIS_STORAGE_URL":            []byte("redis://username:password@my-redis:5000/1"),
				},
			},
			&RedisConfig{
				URL:         "redis://username:password@my-redis:5000/1",
				SentinelURL: "redis://username:password@sentinel-1:5000, redis://username:password@sentinel-2:5000, redis://username:password@sentinel-3:5000",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(subT *testing.T) {
			function := tc.redisSecretFunction
			redisRetrieved := function(tc.secret)

			require.Equal(t, tc.redisConfig.URL, redisRetrieved.URL)
			require.Equal(t, tc.redisConfig.SentinelURL, redisRetrieved.SentinelURL)
			require.Equal(t, tc.redisConfig.SentinelMaster, redisRetrieved.SentinelMaster)
		})
	}
}

func TestConfigureNoConfig(t *testing.T) {
	rdb, err := Configure(nil)
	require.NoError(t, err)
	require.Nil(t, rdb, "rdb client should be nil")
}

func TestConfigureWithEmptyConfig(t *testing.T) {
	rdb, err := Configure(&RedisConfig{})
	require.EqualError(t, err, "redis: invalid URL scheme: ")
	require.Nil(t, rdb, "rdb client should be nil")
}

func TestConfigureWithInvalidConfig(t *testing.T) {

	testCases := []struct {
		testName    string
		redisConfig *RedisConfig
		err         string
	}{
		{
			"invalid url scheme",
			&RedisConfig{
				URL: "invalid://redis:5000",
			},
			"redis: invalid URL scheme: invalid",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(subT *testing.T) {
			rdb, err := Configure(tc.redisConfig)
			require.EqualError(t, err, tc.err)
			require.Nil(t, rdb, "rdb client should be nil")
		})
	}
}

func TestConnectToRedis(t *testing.T) {
	testCases := []struct {
		testName         string
		scheme           string
		username         string
		password         string
		overridepassword string
		expected         *redis.Options
	}{
		{
			"With redis scheme",
			"redis",
			"",
			"",
			"",
			&redis.Options{
				Username: "",
				Password: "",
			},
		},
		{
			"With rediss scheme",
			"rediss",
			"",
			"",
			"",
			&redis.Options{
				Username: "",
				Password: "",
			},
		},
		{
			"With username and password from url",
			"redis",
			"foo",
			"bar",
			"",
			&redis.Options{
				Username: "foo",
				Password: "bar",
			},
		},
		{
			"With password from url only",
			"redis",
			"",
			"bar",
			"",
			&redis.Options{
				Username: "",
				Password: "bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(subT *testing.T) {
			mockServer := &mockRedisServer{}
			a := mockServer.Listen(t)

			var u string
			if tc.username != "" || tc.password != "" {
				u = fmt.Sprintf("%s://%s:%s@%s", tc.scheme, tc.username, tc.password, a)
			} else {
				u = fmt.Sprintf("%s://%s", tc.scheme, a)
			}

			config := &RedisConfig{
				URL: u,
			}
			rdb, err := Configure(config)
			require.NoError(t, err)
			defer rdb.Close()

			require.NotNil(t, rdb.Conn(), "Pool should not be nil")

			opt := rdb.Options()
			require.Equal(t, tc.expected.Username, opt.Username)
			require.Equal(t, tc.expected.Password, opt.Password)

			rdb.Ping(context.Background())
			require.True(t, mockServer.Connected())
		})
	}
}

func TestSentinelOptions(t *testing.T) {
	testCases := []struct {
		testName    string
		redisConfig *RedisConfig
		expected    *redis.FailoverOptions
	}{
		{
			"no sentiel master",
			&RedisConfig{
				SentinelURL: "redis://sentinel1:5000",
			},
			&redis.FailoverOptions{
				SentinelAddrs: []string{"sentinel1:5000"},
			},
		},
		{
			"with sentiel master",
			&RedisConfig{
				URL:         "redis://master:3000/1",
				SentinelURL: "redis://sentinel1:5000",
			},
			&redis.FailoverOptions{
				MasterName:    "master",
				SentinelAddrs: []string{"sentinel1:5000"},
			},
		},
		{
			"with username/password in master url",
			&RedisConfig{
				URL:         "redis://username:password@master:3000/1",
				SentinelURL: "redis://sentinel:sentinelpass@sentinel1:5000",
			},
			&redis.FailoverOptions{
				MasterName:    "master",
				SentinelAddrs: []string{"sentinel1:5000"},
				Username:      "username",
				Password:      "password",
			},
		},
		{
			"with multiple sentiels",
			&RedisConfig{
				URL:         "redis://master:3000/1",
				SentinelURL: "redis://sentinel:sentinelpass@sentinel1:5000, redis://sentinel2:5000, redis://sentinel3:5000",
			},
			&redis.FailoverOptions{
				MasterName:    "master",
				SentinelAddrs: []string{"sentinel1:5000", "sentinel2:5000", "sentinel3:5000"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(subT *testing.T) {
			opt, err := sentinelOptions(tc.redisConfig)
			require.NoError(t, err)

			require.Equal(t, tc.expected.SentinelAddrs, opt.SentinelAddrs)
			require.Equal(t, tc.expected.MasterName, opt.MasterName)
			require.Equal(t, tc.expected.Username, opt.Username)
			require.Equal(t, tc.expected.Password, opt.Password)
		})
	}
}

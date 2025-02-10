package helper

import (
	"testing"

	goredis "github.com/redis/go-redis/v9"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReconcileRedisSecrets(t *testing.T) {
	cases := []struct {
		testName            string
		redisSecretFunction func(v1.Secret) (Redis, error)
		secret              v1.Secret
		redis               Redis
	}{
		{
			"SystemRedisSecretNoPasswordNoSentinels1",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte(""),
					"URL":            []byte("redis://my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "",
				},
			},
		},
		{
			"SystemRedisSecretNoPasswordNoSentinels2",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte(""),
					"URL":            []byte("redis://my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "",
				},
			},
		},
		{
			"SystemRedisSecretNoPasswordNoSentinels3",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte(""),
					"URL":            []byte("redis://my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "",
				},
			},
		},
		{
			"SystemRedisSecretPasswordNoSentinels1",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte(""),
					"URL":            []byte("redis://:password@my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password",
				},
			},
		},
		{
			"SystemRedisSecretPasswordNoSentinels2",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte(""),
					"URL":            []byte("redis://:password@my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password",
				},
			},
		},
		{
			"SystemRedisSecretPasswordNoSentinels3",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte(""),
					"URL":            []byte("redis://:password@my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password",
				},
			},
		},
		{
			"SystemRedisSecretPasswordNoSentinels3",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte(""),
					"URL":            []byte("redis://:password@my-redis"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:6379",
					Password: "password",
				},
			},
		},
		{
			"SystemRedisSecretNoPasswordSentinels1",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte("redis://sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"URL":            []byte("redis://my-redis"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:6379",
					Password: "",
				},
			},
		},
		{
			"SystemRedisSecretPasswordSentinels1",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte("redis://:password@sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"URL":            []byte("redis://my-redis"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "password",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:6379",
					Password: "",
				},
			},
		},
		{
			"SystemRedisSecretPasswordSentinels2",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte("redis://:password@sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"URL":            []byte("redis://:password1@my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "password",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password1",
				},
			},
		},
		{
			"SystemRedisSecretPasswordSentinels3",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte("redis://:password@sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"URL":            []byte("redis://:password1@my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "password",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password1",
				},
			},
		},
		{
			"SystemRedisSecretPasswordSentinels3",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte(":password@sentinel.cloud-resource-operator.svc.cluster.local"),
					"URL":            []byte("redis://:asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:6379",
					Password: "password",
				},
				sentinelGroup: "redisgrp",
				redisOptions: goredis.Options{
					Addr:     "redisgrp:6379",
					Password: "asdsada121252112sdag21123",
				},
			},
		},
		{
			"SystemRedisSecretPasswordSentinels3",
			reconcileSystemRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-redis",
				},
				Data: map[string][]byte{
					"SENTINEL_HOSTS": []byte("redis//:password@sentinel.cloud-resource-operator.svc.cluster.local, redis//:password@sentinel.cloud-resource-operator.svc.cluster.local,redis//:password@sentinel.cloud-resource-operator.svc.cluster.local"),
					"URL":            []byte("redis://:asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:6379",
					Password: "password",
				},
				sentinelGroup: "redisgrp",
				redisOptions: goredis.Options{
					Addr:     "redisgrp:6379",
					Password: "asdsada121252112sdag21123",
				},
			},
		},
		{
			"QueuesRedisSecretNoPasswordNoSentinels1",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte(""),
					"REDIS_QUEUES_URL":            []byte("redis://my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "",
				},
			},
		},
		{
			"QueuesRedisSecretNoPasswordNoSentinels2",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte(""),
					"REDIS_QUEUES_URL":            []byte("redis://my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "",
				},
			},
		},
		{
			"QueuesRedisSecretNoPasswordNoSentinels3",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte(""),
					"REDIS_QUEUES_URL":            []byte("redis://my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "",
				},
			},
		},
		{
			"QueuesRedisSecretPasswordNoSentinels1",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte(""),
					"REDIS_QUEUES_URL":            []byte("redis://:password@my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password",
				},
			},
		},
		{
			"QueuesRedisSecretPasswordNoSentinels2",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte(""),
					"REDIS_QUEUES_URL":            []byte("redis://:password@my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password",
				},
			},
		},
		{
			"QueuesRedisSecretPasswordNoSentinels3",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte(""),
					"REDIS_QUEUES_URL":            []byte("redis://:password@my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password",
				},
			},
		},
		{
			"QueuesRedisSecretPasswordNoSentinels3",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte(""),
					"REDIS_QUEUES_URL":            []byte("redis://:password@my-redis"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:6379",
					Password: "password",
				},
			},
		},
		{
			"QueuesRedisSecretNoPasswordSentinels1",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte("redis://sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"REDIS_QUEUES_URL":            []byte("redis://my-redis"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:6379",
					Password: "",
				},
			},
		},
		{
			"QueuesRedisSecretPasswordSentinels1",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte("redis://:password@sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"REDIS_QUEUES_URL":            []byte("redis://my-redis"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "password",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:6379",
					Password: "",
				},
			},
		},
		{
			"QueuesRedisSecretPasswordSentinels2",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte("redis://:password@sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"REDIS_QUEUES_URL":            []byte("redis://:password1@my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "password",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password1",
				},
			},
		},
		{
			"QueuesRedisSecretPasswordSentinels3",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte("redis://:password@sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"REDIS_QUEUES_URL":            []byte("redis://:password1@my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "password",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password1",
				},
			},
		},
		{
			"QueuesRedisSecretPasswordSentinels3",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte(":password@sentinel.cloud-resource-operator.svc.cluster.local"),
					"REDIS_QUEUES_URL":            []byte("redis://:asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:6379",
					Password: "password",
				},
				sentinelGroup: "redisgrp",
				redisOptions: goredis.Options{
					Addr:     "redisgrp:6379",
					Password: "asdsada121252112sdag21123",
				},
			},
		},
		{
			"QueuesRedisSecretPasswordSentinels3",
			reconcileQueuesRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_QUEUES_SENTINEL_HOSTS": []byte("redis//:password@sentinel.cloud-resource-operator.svc.cluster.local, redis//:password@sentinel.cloud-resource-operator.svc.cluster.local,redis//:password@sentinel.cloud-resource-operator.svc.cluster.local"),
					"REDIS_QUEUES_URL":            []byte("redis://:asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:6379",
					Password: "password",
				},
				sentinelGroup: "redisgrp",
				redisOptions: goredis.Options{
					Addr:     "redisgrp:6379",
					Password: "asdsada121252112sdag21123",
				},
			},
		},
		{
			"StorageRedisSecretNoPasswordNoSentinels1",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte(""),
					"REDIS_STORAGE_URL":            []byte("redis://my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "",
				},
			},
		},
		{
			"StorageRedisSecretNoPasswordNoSentinels2",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte(""),
					"REDIS_STORAGE_URL":            []byte("redis://my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "",
				},
			},
		},
		{
			"StorageRedisSecretNoPasswordNoSentinels3",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte(""),
					"REDIS_STORAGE_URL":            []byte("redis://my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "",
				},
			},
		},
		{
			"StorageRedisSecretPasswordNoSentinels1",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte(""),
					"REDIS_STORAGE_URL":            []byte("redis://:password@my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password",
				},
			},
		},
		{
			"StorageRedisSecretPasswordNoSentinels2",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte(""),
					"REDIS_STORAGE_URL":            []byte("redis://:password@my-redis:5000/1"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password",
				},
			},
		},
		{
			"StorageRedisSecretPasswordNoSentinels3",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte(""),
					"REDIS_STORAGE_URL":            []byte("redis://:password@my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password",
				},
			},
		},
		{
			"StorageRedisSecretPasswordNoSentinels3",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte(""),
					"REDIS_STORAGE_URL":            []byte("redis://:password@my-redis"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "",
					Password: "",
				},
				sentinelGroup: "",
				redisOptions: goredis.Options{
					Addr:     "my-redis:6379",
					Password: "password",
				},
			},
		},
		{
			"StorageRedisSecretNoPasswordSentinels1",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte("redis://sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"REDIS_STORAGE_URL":            []byte("redis://my-redis"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:6379",
					Password: "",
				},
			},
		},
		{
			"StorageRedisSecretPasswordSentinels1",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte("redis://:password@sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"REDIS_STORAGE_URL":            []byte("redis://my-redis"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "password",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:6379",
					Password: "",
				},
			},
		},
		{
			"StorageRedisSecretPasswordSentinels2",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte("redis://:password@sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"REDIS_STORAGE_URL":            []byte("redis://:password1@my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "password",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password1",
				},
			},
		},
		{
			"StorageRedisSecretPasswordSentinels3",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte("redis://:password@sentinel.cloud-resource-operator.svc.cluster.local:5000"),
					"REDIS_STORAGE_URL":            []byte("redis://:password1@my-redis:5000"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:5000",
					Password: "password",
				},
				sentinelGroup: "my-redis",
				redisOptions: goredis.Options{
					Addr:     "my-redis:5000",
					Password: "password1",
				},
			},
		},
		{
			"StorageRedisSecretPasswordSentinels3",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte(":password@sentinel.cloud-resource-operator.svc.cluster.local"),
					"REDIS_STORAGE_URL":            []byte("redis://:asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:6379",
					Password: "password",
				},
				sentinelGroup: "redisgrp",
				redisOptions: goredis.Options{
					Addr:     "redisgrp:6379",
					Password: "asdsada121252112sdag21123",
				},
			},
		},
		{
			"StorageRedisSecretPasswordSentinels3",
			reconcileStorageRedisSecret,
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backend-redis",
				},
				Data: map[string][]byte{
					"REDIS_STORAGE_SENTINEL_HOSTS": []byte("redis//:password@sentinel.cloud-resource-operator.svc.cluster.local, redis//:password@sentinel.cloud-resource-operator.svc.cluster.local,redis//:password@sentinel.cloud-resource-operator.svc.cluster.local"),
					"REDIS_STORAGE_URL":            []byte("redis://:asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelOptions: goredis.Options{
					Addr:     "sentinel.cloud-resource-operator.svc.cluster.local:6379",
					Password: "password",
				},
				sentinelGroup: "redisgrp",
				redisOptions: goredis.Options{
					Addr:     "redisgrp:6379",
					Password: "asdsada121252112sdag21123",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			function := tc.redisSecretFunction
			redisRetrieved, err := function(tc.secret)

			if err != nil {
				subT.Fatalf("test failed for test case %s, err: %v ", tc.testName, err)
			}
			if redisRetrieved.redisOptions.Addr != tc.redis.redisOptions.Addr {
				subT.Fatalf("test failed for test case %s, expected redis address %v but got %v", tc.testName, tc.redis.redisOptions.Addr, redisRetrieved.redisOptions.Addr)
			}
			if redisRetrieved.redisOptions.Password != tc.redis.redisOptions.Password {
				subT.Fatalf("test failed for test case %s, expected redis password %v but got %v", tc.testName, tc.redis.redisOptions.Password, redisRetrieved.redisOptions.Password)
			}
			if redisRetrieved.sentinelOptions.Addr != tc.redis.sentinelOptions.Addr {
				subT.Fatalf("test failed for test case %s, expected redis sentinel host %v but got %v", tc.testName, tc.redis.sentinelOptions.Addr, redisRetrieved.sentinelOptions.Addr)
			}
			if redisRetrieved.sentinelOptions.Password != tc.redis.sentinelOptions.Password {
				subT.Fatalf("test failed for test case %s, expected redis sentinel password %v but got %v", tc.testName, tc.redis.sentinelOptions.Password, redisRetrieved.sentinelOptions.Password)
			}
			if redisRetrieved.sentinelGroup != tc.redis.sentinelGroup {
				subT.Fatalf("test failed for test case %s, expected redis group %v but got %v", tc.testName, tc.redis.sentinelGroup, redisRetrieved.sentinelGroup)
			}
		})
	}
}


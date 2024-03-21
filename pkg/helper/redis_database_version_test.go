package helper

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReconcileRedisSecrets(t *testing.T) {
	cases := []struct {
		testName            string
		redisSecretFunction func(v1.Secret) Redis
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
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "5000",
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
					"URL":            []byte("my-redis:5000"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "5000",
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
					"URL":            []byte("my-redis:5000/1"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "5000",
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
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "5000",
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
					"URL":            []byte(":password@my-redis:5000/1"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "5000",
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
					"URL":            []byte(":password@my-redis:5000"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "5000",
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
					"URL":            []byte(":password@my-redis"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "6379",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "6379",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "6379",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "password1",
				redisPort:        "5000",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "password1",
				redisPort:        "5000",
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
					"URL":            []byte(":asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "6379",
				sentinelGroup:    "redisgrp",
				redisHost:        "redisgrp",
				redisPassword:    "asdsada121252112sdag21123",
				redisPort:        "6379",
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
					"URL":            []byte(":asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "6379",
				sentinelGroup:    "redisgrp",
				redisHost:        "redisgrp",
				redisPassword:    "asdsada121252112sdag21123",
				redisPort:        "6379",
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
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "5000",
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
					"REDIS_QUEUES_URL":            []byte("my-redis:5000"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "5000",
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
					"REDIS_QUEUES_URL":            []byte("my-redis:5000/1"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "5000",
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
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "5000",
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
					"REDIS_QUEUES_URL":            []byte(":password@my-redis:5000/1"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "5000",
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
					"REDIS_QUEUES_URL":            []byte(":password@my-redis:5000"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "5000",
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
					"REDIS_QUEUES_URL":            []byte(":password@my-redis"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "6379",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "6379",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "6379",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "password1",
				redisPort:        "5000",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "password1",
				redisPort:        "5000",
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
					"REDIS_QUEUES_URL":            []byte(":asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "6379",
				sentinelGroup:    "redisgrp",
				redisHost:        "redisgrp",
				redisPassword:    "asdsada121252112sdag21123",
				redisPort:        "6379",
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
					"REDIS_QUEUES_URL":            []byte(":asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "6379",
				sentinelGroup:    "redisgrp",
				redisHost:        "redisgrp",
				redisPassword:    "asdsada121252112sdag21123",
				redisPort:        "6379",
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
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "5000",
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
					"REDIS_STORAGE_URL":            []byte("my-redis:5000"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "5000",
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
					"REDIS_STORAGE_URL":            []byte("my-redis:5000/1"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "5000",
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
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "5000",
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
					"REDIS_STORAGE_URL":            []byte(":password@my-redis:5000/1"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "5000",
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
					"REDIS_STORAGE_URL":            []byte(":password@my-redis:5000"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "5000",
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
					"REDIS_STORAGE_URL":            []byte(":password@my-redis"),
				},
			},
			Redis{
				sentinelHost:     "",
				sentinelPassword: "",
				sentinelPort:     "",
				redisHost:        "my-redis",
				redisPassword:    "password",
				redisPort:        "6379",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "6379",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "",
				redisPort:        "6379",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "password1",
				redisPort:        "5000",
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
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "5000",
				sentinelGroup:    "my-redis",
				redisHost:        "my-redis",
				redisPassword:    "password1",
				redisPort:        "5000",
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
					"REDIS_STORAGE_URL":            []byte(":asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "6379",
				sentinelGroup:    "redisgrp",
				redisHost:        "redisgrp",
				redisPassword:    "asdsada121252112sdag21123",
				redisPort:        "6379",
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
					"REDIS_STORAGE_URL":            []byte(":asdsada121252112sdag21123@redisgrp"),
				},
			},
			Redis{
				sentinelHost:     "sentinel.cloud-resource-operator.svc.cluster.local",
				sentinelPassword: "password",
				sentinelPort:     "6379",
				sentinelGroup:    "redisgrp",
				redisHost:        "redisgrp",
				redisPassword:    "asdsada121252112sdag21123",
				redisPort:        "6379",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			function := tc.redisSecretFunction
			redisRetrieved := function(tc.secret)

			if redisRetrieved.redisHost != tc.redis.redisHost {
				subT.Fatalf("test failed for test case %s, expected redis host %v but got %v", tc.testName, tc.redis.redisHost, redisRetrieved.redisHost)
			}
			if redisRetrieved.redisPort != tc.redis.redisPort {
				subT.Fatalf("test failed for test case %s, expected redis port %v but got %v", tc.testName, tc.redis.redisPort, redisRetrieved.redisPort)
			}
			if redisRetrieved.redisPassword != tc.redis.redisPassword {
				subT.Fatalf("test failed for test case %s, expected redis password %v but got %v", tc.testName, tc.redis.redisPassword, redisRetrieved.redisPassword)
			}
			if redisRetrieved.sentinelHost != tc.redis.sentinelHost {
				subT.Fatalf("test failed for test case %s, expected redis sentinel host %v but got %v", tc.testName, tc.redis.sentinelHost, redisRetrieved.sentinelHost)
			}
			if redisRetrieved.sentinelPort != tc.redis.sentinelPort {
				subT.Fatalf("test failed for test case %s, expected redis sentinel port %v but got %v", tc.testName, tc.redis.sentinelPort, redisRetrieved.sentinelPort)
			}
			if redisRetrieved.sentinelPassword != tc.redis.sentinelPassword {
				subT.Fatalf("test failed for test case %s, expected redis sentinel password %v but got %v", tc.testName, tc.redis.sentinelPassword, redisRetrieved.sentinelPassword)
			}
			if redisRetrieved.sentinelGroup != tc.redis.sentinelGroup {
				subT.Fatalf("test failed for test case %s, expected redis group %v but got %v", tc.testName, tc.redis.sentinelGroup, redisRetrieved.sentinelGroup)
			}
		})
	}
}

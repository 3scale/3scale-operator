package component

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CommonEmbeddedRedis struct {
	Options *CommonEmbeddedRedisOptions
}

func NewCommonEmbeddedRedis(options *CommonEmbeddedRedisOptions) *CommonEmbeddedRedis {
	return &CommonEmbeddedRedis{Options: options}
}

func (commonRedis *CommonEmbeddedRedis) ConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: commonRedis.buildConfigMapObjectMeta(),
		TypeMeta:   commonRedis.buildConfigMapTypeMeta(),
		Data:       commonRedis.buildConfigMapData(),
	}
}

func (commonRedis *CommonEmbeddedRedis) buildConfigMapObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   backendRedisConfigVolumeName,
		Labels: commonRedis.Options.ConfigMapLabels,
	}
}

func (commonRedis *CommonEmbeddedRedis) buildConfigMapTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: "v1",
	}
}

func (commonRedis *CommonEmbeddedRedis) buildConfigMapData() map[string]string {
	return map[string]string{
		"redis.conf": commonRedis.getRedisConfData(),
	}
}

func (commonRedis *CommonEmbeddedRedis) getRedisConfData() string { // TODO read this from a real file
	return `protected-mode no

port 6379

timeout 0
tcp-keepalive 300

daemonize no
supervised no

loglevel notice

databases 16

save 900 1
save 300 10
save 60 10000

stop-writes-on-bgsave-error yes

rdbcompression yes
rdbchecksum yes

dbfilename dump.rdb

slave-serve-stale-data yes
slave-read-only yes

repl-diskless-sync no
repl-disable-tcp-nodelay no

appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
aof-load-truncated yes

lua-time-limit 5000

activerehashing no

aof-rewrite-incremental-fsync yes
dir /var/lib/redis/data
`
}

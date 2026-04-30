package component

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"

	"github.com/3scale/3scale-operator/pkg/helper"
)

func TestBackendComponentRedisTLSEnvVars(t *testing.T) {
	cases := []struct {
		name     string
		options  BackendOptions
		expected []v1.EnvVar
	}{
		{
			"TLSDisabled",
			BackendOptions{
				BackendRedisTLS: TLSConfig{Enabled: false},
			},
			nil,
		},
		{
			"OneWayTLS_CAOnly",
			BackendOptions{
				BackendRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "some-ca-cert",
				},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("CONFIG_REDIS_SSL", "1"),
				helper.EnvVarFromValue("CONFIG_REDIS_CA_FILE", ConfigRedisCaFilePath),
			},
		},
		{
			"MutualTLS",
			BackendOptions{
				BackendRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "some-ca-cert",
					Certificate:   "some-cert",
					Key:           "some-key",
				},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("CONFIG_REDIS_SSL", "1"),
				helper.EnvVarFromValue("CONFIG_REDIS_CA_FILE", ConfigRedisCaFilePath),
				helper.EnvVarFromValue("CONFIG_REDIS_CERT", ConfigRedisClientCertPath),
				helper.EnvVarFromValue("CONFIG_REDIS_PRIVATE_KEY", ConfigRedisPrivateKeyPath),
			},
		},
		{
			"TLSEnabled_NoCA_NoCert",
			BackendOptions{
				BackendRedisTLS: TLSConfig{Enabled: true},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("CONFIG_REDIS_SSL", "1"),
			},
		},
		{
			"TLSEnabled_CertWithoutKey",
			BackendOptions{
				BackendRedisTLS: TLSConfig{
					Enabled:     true,
					Certificate: "some-cert",
				},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("CONFIG_REDIS_SSL", "1"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			backend := &Backend{Options: &tc.options}
			result := backend.BackendRedisTLSEnvVars()
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("BackendRedisTLSEnvVars mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBackendComponentQueuesRedisTLSEnvVars(t *testing.T) {
	cases := []struct {
		name     string
		options  BackendOptions
		expected []v1.EnvVar
	}{
		{
			"TLSDisabled",
			BackendOptions{
				BackendRedisQueuesTLS: TLSConfig{Enabled: false},
			},
			nil,
		},
		{
			"OneWayTLS_CAOnly",
			BackendOptions{
				BackendRedisQueuesTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "some-ca-cert",
				},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("CONFIG_QUEUES_SSL", "1"),
				helper.EnvVarFromValue("CONFIG_QUEUES_CA_FILE", ConfigQueuesCaFilePath),
			},
		},
		{
			"MutualTLS",
			BackendOptions{
				BackendRedisQueuesTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "some-ca-cert",
					Certificate:   "some-cert",
					Key:           "some-key",
				},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("CONFIG_QUEUES_SSL", "1"),
				helper.EnvVarFromValue("CONFIG_QUEUES_CA_FILE", ConfigQueuesCaFilePath),
				helper.EnvVarFromValue("CONFIG_QUEUES_CERT", ConfigQueuesClientCertPath),
				helper.EnvVarFromValue("CONFIG_QUEUES_PRIVATE_KEY", ConfigQueuesPrivateKeyPath),
			},
		},
		{
			"TLSEnabled_NoCA_NoCert",
			BackendOptions{
				BackendRedisQueuesTLS: TLSConfig{Enabled: true},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("CONFIG_QUEUES_SSL", "1"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			backend := &Backend{Options: &tc.options}
			result := backend.QueuesRedisTLSEnvVars()
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("QueuesRedisTLSEnvVars mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBackendComponentRedisTLSVolumes(t *testing.T) {
	cases := []struct {
		name     string
		options  BackendOptions
		expected []v1.Volume
	}{
		{
			"BothDisabled",
			BackendOptions{
				BackendRedisTLS:       TLSConfig{Enabled: false},
				BackendRedisQueuesTLS: TLSConfig{Enabled: false},
			},
			[]v1.Volume{},
		},
		{
			"StorageOnly_OneWayTLS",
			BackendOptions{
				BackendRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca-data",
				},
				BackendRedisQueuesTLS: TLSConfig{Enabled: false},
			},
			[]v1.Volume{
				{
					Name: "backend-redis-tls",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: BackendSecretBackendRedisSecretName,
							Items: []v1.KeyToPath{
								{Key: "REDIS_SSL_CA", Path: "backend-redis-ca.crt"},
							},
						},
					},
				},
			},
		},
		{
			"QueuesOnly_OneWayTLS",
			BackendOptions{
				BackendRedisTLS: TLSConfig{Enabled: false},
				BackendRedisQueuesTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca-data",
				},
			},
			[]v1.Volume{
				{
					Name: "queues-redis-tls",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: BackendSecretBackendRedisSecretName,
							Items: []v1.KeyToPath{
								{Key: "REDIS_SSL_QUEUES_CA", Path: "backend-redis-queues-ca.crt"},
							},
						},
					},
				},
			},
		},
		{
			"BothEnabled_MutualTLS",
			BackendOptions{
				BackendRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca",
					Certificate:   "cert",
					Key:           "key",
				},
				BackendRedisQueuesTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca",
					Certificate:   "cert",
					Key:           "key",
				},
			},
			[]v1.Volume{
				{
					Name: "backend-redis-tls",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: BackendSecretBackendRedisSecretName,
							Items: []v1.KeyToPath{
								{Key: "REDIS_SSL_CA", Path: "backend-redis-ca.crt"},
								{Key: "REDIS_SSL_CERT", Path: "backend-redis-client.crt"},
								{Key: "REDIS_SSL_KEY", Path: "backend-redis-private.key"},
							},
						},
					},
				},
				{
					Name: "queues-redis-tls",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: BackendSecretBackendRedisSecretName,
							Items: []v1.KeyToPath{
								{Key: "REDIS_SSL_QUEUES_CA", Path: "backend-redis-queues-ca.crt"},
								{Key: "REDIS_SSL_QUEUES_CERT", Path: "backend-redis-queues-client.crt"},
								{Key: "REDIS_SSL_QUEUES_KEY", Path: "backend-redis-queues-private.key"},
							},
						},
					},
				},
			},
		},
		{
			"StorageMutualTLS_QueuesOneWayTLS",
			BackendOptions{
				BackendRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca",
					Certificate:   "cert",
					Key:           "key",
				},
				BackendRedisQueuesTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca",
				},
			},
			[]v1.Volume{
				{
					Name: "backend-redis-tls",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: BackendSecretBackendRedisSecretName,
							Items: []v1.KeyToPath{
								{Key: "REDIS_SSL_CA", Path: "backend-redis-ca.crt"},
								{Key: "REDIS_SSL_CERT", Path: "backend-redis-client.crt"},
								{Key: "REDIS_SSL_KEY", Path: "backend-redis-private.key"},
							},
						},
					},
				},
				{
					Name: "queues-redis-tls",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: BackendSecretBackendRedisSecretName,
							Items: []v1.KeyToPath{
								{Key: "REDIS_SSL_QUEUES_CA", Path: "backend-redis-queues-ca.crt"},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			backend := &Backend{Options: &tc.options}
			result := backend.backendVolumes()
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("backendVolumes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBackendComponentRedisTLSVolumeMounts(t *testing.T) {
	cases := []struct {
		name     string
		options  BackendOptions
		expected []v1.VolumeMount
	}{
		{
			"BothDisabled",
			BackendOptions{
				BackendRedisTLS:       TLSConfig{Enabled: false},
				BackendRedisQueuesTLS: TLSConfig{Enabled: false},
			},
			[]v1.VolumeMount{},
		},
		{
			"StorageOnly",
			BackendOptions{
				BackendRedisTLS:       TLSConfig{Enabled: true},
				BackendRedisQueuesTLS: TLSConfig{Enabled: false},
			},
			[]v1.VolumeMount{
				{Name: "backend-redis-tls", ReadOnly: false, MountPath: "/tls/backend-redis"},
			},
		},
		{
			"QueuesOnly",
			BackendOptions{
				BackendRedisTLS:       TLSConfig{Enabled: false},
				BackendRedisQueuesTLS: TLSConfig{Enabled: true},
			},
			[]v1.VolumeMount{
				{Name: "queues-redis-tls", ReadOnly: false, MountPath: "/tls/backend-queues"},
			},
		},
		{
			"BothEnabled",
			BackendOptions{
				BackendRedisTLS:       TLSConfig{Enabled: true},
				BackendRedisQueuesTLS: TLSConfig{Enabled: true},
			},
			[]v1.VolumeMount{
				{Name: "backend-redis-tls", ReadOnly: false, MountPath: "/tls/backend-redis"},
				{Name: "queues-redis-tls", ReadOnly: false, MountPath: "/tls/backend-queues"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			backend := &Backend{Options: &tc.options}
			result := backend.backendContainerVolumeMounts()
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("backendContainerVolumeMounts mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

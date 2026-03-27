package component

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"

	"github.com/3scale/3scale-operator/pkg/helper"
)

func TestSystemRedisTLSEnvVars(t *testing.T) {
	cases := []struct {
		name     string
		options  SystemOptions
		expected []v1.EnvVar
	}{
		{
			"TLSDisabled",
			SystemOptions{
				SystemRedisTLS: TLSConfig{Enabled: false},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("REDIS_SSL", "0"),
			},
		},
		{
			"OneWayTLS_CAOnly",
			SystemOptions{
				SystemRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "some-ca-cert",
				},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("REDIS_SSL", "1"),
				helper.EnvVarFromValue("REDIS_CA_FILE", "/tls/system-redis/system-redis-ca.crt"),
			},
		},
		{
			"MutualTLS",
			SystemOptions{
				SystemRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "some-ca-cert",
					Certificate:   "some-cert",
					Key:           "some-key",
				},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("REDIS_SSL", "1"),
				helper.EnvVarFromValue("REDIS_CA_FILE", "/tls/system-redis/system-redis-ca.crt"),
				helper.EnvVarFromValue("REDIS_CLIENT_CERT", "/tls/system-redis/system-redis-client.crt"),
				helper.EnvVarFromValue("REDIS_PRIVATE_KEY", "/tls/system-redis/system-redis-private.key"),
			},
		},
		{
			"TLSEnabled_NoCA_NoCert",
			SystemOptions{
				SystemRedisTLS: TLSConfig{Enabled: true},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("REDIS_SSL", "1"),
			},
		},
		{
			"TLSEnabled_CertWithoutKey",
			SystemOptions{
				SystemRedisTLS: TLSConfig{
					Enabled:     true,
					Certificate: "some-cert",
				},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("REDIS_SSL", "1"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			system := &System{Options: &tc.options}
			result := system.SystemRedisTLSEnvVars()
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("SystemRedisTLSEnvVars mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBackendRedisTLSEnvVars(t *testing.T) {
	cases := []struct {
		name     string
		options  SystemOptions
		expected []v1.EnvVar
	}{
		{
			"TLSDisabled",
			SystemOptions{
				BackendRedisTLS: TLSConfig{Enabled: false},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("BACKEND_REDIS_SSL", "0"),
			},
		},
		{
			"OneWayTLS_CAOnly",
			SystemOptions{
				BackendRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "some-ca-cert",
				},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("BACKEND_REDIS_SSL", "1"),
				helper.EnvVarFromValue("BACKEND_REDIS_CA_FILE", "/tls/backend-redis/backend-redis-ca.crt"),
			},
		},
		{
			"MutualTLS",
			SystemOptions{
				BackendRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "some-ca-cert",
					Certificate:   "some-cert",
					Key:           "some-key",
				},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("BACKEND_REDIS_SSL", "1"),
				helper.EnvVarFromValue("BACKEND_REDIS_CA_FILE", "/tls/backend-redis/backend-redis-ca.crt"),
				helper.EnvVarFromValue("BACKEND_REDIS_CLIENT_CERT", "/tls/backend-redis/backend-redis-client.crt"),
				helper.EnvVarFromValue("BACKEND_REDIS_PRIVATE_KEY", "/tls/backend-redis/backend-redis-private.key"),
			},
		},
		{
			"TLSEnabled_NoCA_NoCert",
			SystemOptions{
				BackendRedisTLS: TLSConfig{Enabled: true},
			},
			[]v1.EnvVar{
				helper.EnvVarFromValue("BACKEND_REDIS_SSL", "1"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			system := &System{Options: &tc.options}
			result := system.BackendRedisTLSEnvVars()
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("BackendRedisTLSEnvVars mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRedisTLSVolumes(t *testing.T) {
	cases := []struct {
		name     string
		options  SystemOptions
		expected []v1.Volume
	}{
		{
			"BothDisabled",
			SystemOptions{
				SystemRedisTLS:  TLSConfig{Enabled: false},
				BackendRedisTLS: TLSConfig{Enabled: false},
			},
			[]v1.Volume{},
		},
		{
			"SystemOnly_OneWayTLS",
			SystemOptions{
				SystemRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca-data",
				},
				BackendRedisTLS: TLSConfig{Enabled: false},
			},
			[]v1.Volume{
				{
					Name: "system-redis-tls",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: SystemSecretSystemRedisSecretName,
							Items: []v1.KeyToPath{
								{Key: "REDIS_SSL_CA", Path: "system-redis-ca.crt"},
							},
						},
					},
				},
			},
		},
		{
			"BackendOnly_OneWayTLS",
			SystemOptions{
				SystemRedisTLS: TLSConfig{Enabled: false},
				BackendRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca-data",
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
							},
						},
					},
				},
			},
		},
		{
			"BothEnabled_MutualTLS",
			SystemOptions{
				SystemRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca",
					Certificate:   "cert",
					Key:           "key",
				},
				BackendRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca",
					Certificate:   "cert",
					Key:           "key",
				},
			},
			[]v1.Volume{
				{
					Name: "system-redis-tls",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: SystemSecretSystemRedisSecretName,
							Items: []v1.KeyToPath{
								{Key: "REDIS_SSL_CA", Path: "system-redis-ca.crt"},
								{Key: "REDIS_SSL_CERT", Path: "system-redis-client.crt"},
								{Key: "REDIS_SSL_KEY", Path: "system-redis-private.key"},
							},
						},
					},
				},
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
			},
		},
		{
			"SystemMutualTLS_BackendOneWayTLS",
			SystemOptions{
				SystemRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca",
					Certificate:   "cert",
					Key:           "key",
				},
				BackendRedisTLS: TLSConfig{
					Enabled:       true,
					CACertificate: "ca",
				},
			},
			[]v1.Volume{
				{
					Name: "system-redis-tls",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: SystemSecretSystemRedisSecretName,
							Items: []v1.KeyToPath{
								{Key: "REDIS_SSL_CA", Path: "system-redis-ca.crt"},
								{Key: "REDIS_SSL_CERT", Path: "system-redis-client.crt"},
								{Key: "REDIS_SSL_KEY", Path: "system-redis-private.key"},
							},
						},
					},
				},
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			system := &System{Options: &tc.options}
			result := system.redisTLSVolumes()
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("redisTLSVolumes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRedisTLSVolumeMounts(t *testing.T) {
	cases := []struct {
		name     string
		options  SystemOptions
		expected []v1.VolumeMount
	}{
		{
			"BothDisabled",
			SystemOptions{
				SystemRedisTLS:  TLSConfig{Enabled: false},
				BackendRedisTLS: TLSConfig{Enabled: false},
			},
			[]v1.VolumeMount{},
		},
		{
			"SystemOnly",
			SystemOptions{
				SystemRedisTLS:  TLSConfig{Enabled: true},
				BackendRedisTLS: TLSConfig{Enabled: false},
			},
			[]v1.VolumeMount{
				{Name: "system-redis-tls", ReadOnly: false, MountPath: "/tls/system-redis"},
			},
		},
		{
			"BackendOnly",
			SystemOptions{
				SystemRedisTLS:  TLSConfig{Enabled: false},
				BackendRedisTLS: TLSConfig{Enabled: true},
			},
			[]v1.VolumeMount{
				{Name: "backend-redis-tls", ReadOnly: false, MountPath: "/tls/backend-redis"},
			},
		},
		{
			"BothEnabled",
			SystemOptions{
				SystemRedisTLS:  TLSConfig{Enabled: true},
				BackendRedisTLS: TLSConfig{Enabled: true},
			},
			[]v1.VolumeMount{
				{Name: "system-redis-tls", ReadOnly: false, MountPath: "/tls/system-redis"},
				{Name: "backend-redis-tls", ReadOnly: false, MountPath: "/tls/backend-redis"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			system := &System{Options: &tc.options}
			result := system.redisTLSVolumeMounts()
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("redisTLSVolumeMounts mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

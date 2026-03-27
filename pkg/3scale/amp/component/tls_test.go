package component

import "testing"

func TestTLSConfig_HasCA(t *testing.T) {
	cases := []struct {
		name     string
		cfg      TLSConfig
		expected bool
	}{
		{"Empty", TLSConfig{}, false},
		{"WithCA", TLSConfig{CACertificate: "ca"}, true},
		{"WithoutCA", TLSConfig{Certificate: "cert", Key: "key"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.HasCA(); got != tc.expected {
				t.Errorf("HasCA() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestTLSConfig_HasCertAuth(t *testing.T) {
	cases := []struct {
		name     string
		cfg      TLSConfig
		expected bool
	}{
		{"Empty", TLSConfig{}, false},
		{"CertOnly", TLSConfig{Certificate: "cert"}, false},
		{"KeyOnly", TLSConfig{Key: "key"}, false},
		{"BothSet", TLSConfig{Certificate: "cert", Key: "key"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.HasCertAuth(); got != tc.expected {
				t.Errorf("HasCertAuth() = %v, want %v", got, tc.expected)
			}
		})
	}
}

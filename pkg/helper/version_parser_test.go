package helper

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	cases := []struct {
		name            string
		image           string
		expectedVersion string
	}{
		{"test01", "3scale-community-operator.v0.4.0", "0.4.0"},
		{"test02", "quay.io/3scale/3scale28:apicast-3scale-2.8.1-GA", "2.8.1"},
		{"test03", "memcached:1.5", "1.5"},
		{"test04", "redis-32-rhel7", "32"},
		{"test05", "quay.io/3scale/apisonator:nightly", "nightly"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			parsedVersion := ParseVersion(tc.image)
			if parsedVersion != tc.expectedVersion {
				subT.Errorf("versions differ: got: %s; expected: %s", parsedVersion, tc.expectedVersion)
			}
		})
	}
}

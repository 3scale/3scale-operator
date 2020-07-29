package helper

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSanitizeBackendSystemName(t *testing.T) {
	cases := []struct {
		name               string
		systemName         string
		expectedSystemName string
	}{
		{"test01", "hits.45498", "hits"},
		{"test02", "hits.something.45498", "hits.something"},
		{"test03", "hits", "hits"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			newName := SanitizeBackendSystemName(tc.systemName)
			if newName != tc.expectedSystemName {
				diff := cmp.Diff(newName, tc.expectedSystemName)
				subT.Errorf("diff %s", diff)
			}
		})
	}
}

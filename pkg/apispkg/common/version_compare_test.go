package common

import (
	"testing"
)

func TestVersionCompare(t *testing.T) {
	cases := []struct {
		testName        string
		currentVersion  string
		incomingVersion string
		expectedResult  bool
	}{
		{"MultiHopDetected", "2.11", "2.15", true},
		{"MultiHopNotDetected", "2.14", "2.15", false},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			value, _ := CompareMinorVersions(tc.currentVersion, tc.incomingVersion)
			if value != tc.expectedResult {
				subT.Fatalf("test failed for test case %s, expected %v but got %v", tc.testName, tc.expectedResult, value)
			}
		})
	}
}

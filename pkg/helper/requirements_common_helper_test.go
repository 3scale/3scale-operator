package helper

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
		{"IncomingMajorRequiredIsHiger", "7.0.0", "6.2", false},
		{"IncomingMinorRequiredIsHiger", "6.3.0", "6.2", false},
		{"IncomingMajorRequiredIsLower", "5.0.0", "6.0.0", true},
		{"IncomingMinorRequiredIsLower", "6.1.0", "6.2", true},
		{"VersionsMatch", "6.2.0", "6.2.0", true},
		{"IncomingPatchVersionIsHigher", "6.2.1", "6.2.0", false},
		{"IncomingPatchVersionIsLower", "6.2.0", "6.2.1", true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			value := CompareVersions(tc.currentVersion, tc.incomingVersion)
			if value != tc.expectedResult {
				subT.Fatalf("test failed for test case %s, expected %v but got %v", tc.testName, tc.expectedResult, value)
			}
		})
	}
}

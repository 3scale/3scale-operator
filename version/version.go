package version

import (
	"strings"
)

var (
	Version           = "0.13.0"
	threescaleRelease = "2.16.0"
)

func ThreescaleVersionMajorMinor() string {
	parts := strings.Split(threescaleRelease, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return ""
}

func ThreescaleVersionMajorMinorPatch() string {
	return threescaleRelease
}

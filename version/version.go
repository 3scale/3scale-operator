package version

import (
	"strings"
)

var (
	Version           = "0.13.2"
	threescaleRelease = "2.16.2"
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

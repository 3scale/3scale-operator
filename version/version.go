package version

import (
	"strings"
)

var (
	Version           = "0.12.3"
	threescaleRelease = "2.15.3"
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

package common

import (
	"fmt"
	"strconv"
	"strings"
)

func CompareMinorVersions(v1, v2 string) (bool, error) {
	second1, err := extractSecondValue(v1)
	if err != nil {
		// handle error
		return false, err
	}

	second2, err := extractSecondValue(v2)
	if err != nil {
		// handle error
		return false, err
	}

	difference := second1 - second2
	result := difference != 0 && difference != 1 && difference != -1
	return result, err
}

func extractSecondValue(version string) (int, error) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid version format: %s", version)
	}

	second, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	return second, nil
}

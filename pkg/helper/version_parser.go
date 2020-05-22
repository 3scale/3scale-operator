package helper

import (
	"regexp"
)

type versionParser func(string) string

var (
	reg3digitRawParser versionParser = func(text string) string { return regexp.MustCompile(`\d+\.\d+\.\d+`).FindString(text) }
	reg2digitRawParser versionParser = func(text string) string { return regexp.MustCompile(`\d+\.\d+`).FindString(text) }
	regColonRawParser  versionParser = func(text string) string {
		matches := regexp.MustCompile(`:(.+)$`).FindStringSubmatch(text)
		if matches == nil || len(matches) < 2 {
			return ""
		}
		return matches[1]
	}
	reg1digitRawParser versionParser = func(text string) string { return regexp.MustCompile(`\d+`).FindString(text) }
)

func ParseVersion(image string) string {
	regExpsFuncs := []versionParser{reg3digitRawParser, reg2digitRawParser, regColonRawParser, reg1digitRawParser}

	for _, regExpRawFunc := range regExpsFuncs {
		version := regExpRawFunc(image)
		if len(version) > 0 {
			return version
		}
	}

	return "unknown"
}

package helper

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
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

func normalizeDockerURL(image string) string {
	// Only for version parsing purposes, Host and scheme is added to have the URL being parsed correctly
	// normalizeDockerURL tries to build a URL string that can be parsed by net/url library
	// docker image URLS can be tricky sometimes.
	// o := url.Parse("registry.redhat.io:443/rhel8/redis-5:1") -> o.Path is empty
	newImage := image

	if !strings.Contains(image, "/") {
		newImage = fmt.Sprintf("docker.io/%s", newImage)
	}

	if !strings.HasPrefix(newImage, "http://") && !strings.HasPrefix(newImage, "https://") {
		newImage = fmt.Sprintf("http://%s", newImage)
	}

	u, err := url.Parse(newImage)
	if err != nil {
		// silently discard the error and return empty
		return ""
	}

	return u.Path
}

func ParseVersion(image string) string {
	regExpsFuncs := []versionParser{reg3digitRawParser, reg2digitRawParser, regColonRawParser, reg1digitRawParser}

	nImage := normalizeDockerURL(image)

	for _, regExpRawFunc := range regExpsFuncs {
		version := regExpRawFunc(nImage)
		if len(version) > 0 {
			return version
		}
	}

	return "unknown"
}

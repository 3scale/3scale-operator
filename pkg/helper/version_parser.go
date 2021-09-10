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

func normalizeDockerImage(image string) string {
	newImage := image

	if strings.Index(image, "/") < 0 {
		newImage = fmt.Sprintf("docker.io/%s", newImage)
	}

	if !strings.HasPrefix(newImage, "http://") && !strings.HasPrefix(newImage, "https://") {
		newImage = fmt.Sprintf("http://%s", newImage)
	}

	return newImage
}

func ParseVersion(image string) string {
	regExpsFuncs := []versionParser{reg3digitRawParser, reg2digitRawParser, regColonRawParser, reg1digitRawParser}

	nImage := normalizeDockerImage(image)

	u, err := url.Parse(nImage)
	if err != nil {
		return "unknown"
	}

	for _, regExpRawFunc := range regExpsFuncs {
		version := regExpRawFunc(u.Path)
		if len(version) > 0 {
			return version
		}
	}

	return "unknown"
}

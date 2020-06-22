package helper

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
)

// UIDBasedJobName returns a Job name that is compromised of the provided prefix,
// a hyphen and the provided uid: "<prefix>-<uid>". The returned name is a
// DNS1123 Label compliant name. Due to UIDs are 36 characters of length this
// means that the maximum prefix lenght that can be provided is of 26
// characters. If the generated name is not DNS1123 compliant an error is
// returned
func UIDBasedJobName(prefix string, uid types.UID) (string, error) {
	uidStr := string(uid)
	jobName := fmt.Sprintf("%s-%s", prefix, uidStr)
	errStrings := validation.IsDNS1123Label(jobName)
	var err error

	if len(errStrings) > 0 {
		err = fmt.Errorf("Error generating UID-based K8s Job Name: '%s'", strings.Join(errStrings, "\n"))
	}

	return jobName, err
}

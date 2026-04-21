package integration

import (
	"fmt"
	"io"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// minTotalDeploymentUpdates is a floor that guards against a silent counter
// failure. A synthetic image change is applied to system-memcache before
// calling verifyNoDeploymentUpdates (see triggerSyntheticDeploymentUpdate),
// which guarantees the operator issues at least one real UPDATE reconciling
// the deployment back to the desired image. A count of 0 therefore means the
// counter is broken (namespace/name fields not captured) rather than the
// install being genuinely update-free.
const minTotalDeploymentUpdates = 1

// maxTotalDeploymentUpdates is the ceiling on the total number of Deployment
// update calls for a single APIManager CR instance over the full test session.
//
// Rationale: during a normal install each deployment receives 1–3 legitimate
// updates as pods roll out and probes settle, giving ~12–36 total for 12
// deployments. The ceiling is set at 50 to absorb timing variance while
// staying well below what the perpetual-reconcile bug produced (~7 deployments
// × many cycles = hundreds of updates per install).
const maxTotalDeploymentUpdates = 50

// verifyNoDeploymentUpdates asserts that the total deployment update count
// for the given APIManager CR over the full test session is within the ceiling.
// On failure it lists each deployment's individual count to aid diagnosis.
func verifyNoDeploymentUpdates(namespace, name string, w io.Writer) {
	updateCounts := reconcileCounter.GetUpdateCounts(namespace, name)
	totalUpdates := reconcileCounter.GetTotalUpdates(namespace, name)

	fmt.Fprintf(w, "\n=== Deployment Update Report (%s/%s, session ceiling %d) ===\n",
		namespace, name, maxTotalDeploymentUpdates)
	fmt.Fprintf(w, "Total: %d\n", totalUpdates)

	names := make([]string, 0, len(updateCounts))
	for n := range updateCounts {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(w, "  %s: %d\n", n, updateCounts[n])
	}
	fmt.Fprintf(w, "=============================================================\n\n")

	Expect(totalUpdates).To(BeNumerically(">=", minTotalDeploymentUpdates),
		fmt.Sprintf("%s/%s: total deployment updates is 0 — counter is likely misconfigured (namespace/name fields not captured)", namespace, name))
	Expect(totalUpdates).To(BeNumerically("<=", maxTotalDeploymentUpdates),
		deploymentUpdateDetail(namespace, name, updateCounts, totalUpdates))
}

// deploymentUpdateDetail builds a human-readable breakdown for use in a Gomega
// failure message so the offending deployments are immediately visible.
func deploymentUpdateDetail(namespace, name string, counts map[string]int, total int) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s/%s: total deployment updates %d exceeded ceiling %d; per-deployment breakdown:\n",
		namespace, name, total, maxTotalDeploymentUpdates)
	names := make([]string, 0, len(counts))
	for n := range counts {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(&sb, "  %s: %d\n", n, counts[n])
	}
	return sb.String()
}

var _ = Describe // suppress unused import lint for ginkgo dot-import

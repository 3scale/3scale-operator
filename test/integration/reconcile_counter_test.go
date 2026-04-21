package integration

import (
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap/zapcore"
)

// counterState is the single shared mutable core of a ReconcileCounter tree.
// All ReconcileCounter instances produced by With() point to the same counterState,
// so counts accumulate correctly regardless of which logger variant does the write.
// mu protects updateCounts; deploymentNames is read-only after construction.
type counterState struct {
	deploymentNames map[string]bool
	// updateCounts is keyed by "namespace/name" (the APIManager CR instance),
	// then by deployment name. This allows parallel specs targeting different
	// CR instances to accumulate counts independently.
	updateCounts map[string]map[string]int
	mu           sync.Mutex
}

// ReconcileCounter wraps a zapcore.Core to count deployment update events per
// APIManager CR instance. It overrides With and Check so the counter survives
// logger.WithName/WithValues chains, and captures the "namespace" and "name"
// fields that controller-runtime adds to the reconciler logger context.
type ReconcileCounter struct {
	zapcore.Core
	state     *counterState
	namespace string // captured from With() chain
	crName    string // captured from With() chain
}

func (rc *ReconcileCounter) crKey() string {
	if rc.namespace == "" || rc.crName == "" {
		return ""
	}
	return rc.namespace + "/" + rc.crName
}

// NewReconcileCounter creates a new ReconcileCounter wrapping the given core.
func NewReconcileCounter(core zapcore.Core, deploymentNames []string) *ReconcileCounter {
	nameMap := make(map[string]bool)
	for _, name := range deploymentNames {
		nameMap[name] = true
	}
	return &ReconcileCounter{
		Core: core,
		state: &counterState{
			deploymentNames: nameMap,
			updateCounts:    make(map[string]map[string]int),
		},
	}
}

// With returns a new ReconcileCounter wrapping the inner core-with-fields,
// sharing the same counter state. It captures "namespace" and "name" fields
// so Write can key counts by CR instance.
func (rc *ReconcileCounter) With(fields []zapcore.Field) zapcore.Core {
	ns := rc.namespace
	name := rc.crName
	for _, f := range fields {
		if f.Type == zapcore.StringType {
			switch f.Key {
			case "namespace":
				ns = f.String
			case "name":
				name = f.String
			}
		}
	}
	return &ReconcileCounter{
		Core:      rc.Core.With(fields),
		state:     rc.state,
		namespace: ns,
		crName:    name,
	}
}

// Check ensures rc.Write is called (not the inner core's Write) for entries
// that pass the level check.
func (rc *ReconcileCounter) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if rc.Core.Enabled(entry.Level) {
		return ce.AddCore(entry, rc)
	}
	return ce
}

// Write intercepts log entries and counts deployment updates keyed by CR instance.
// The key is resolved in priority order:
//  1. With() chain (context logger): controller-runtime injects "namespace"/"name" before Reconcile.
//  2. Explicit fields on the log entry: UpdateResource logs obj.GetNamespace() as "namespace" and
//     the APIManager owner reference name as "name".
func (rc *ReconcileCounter) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if !strings.Contains(entry.Message, "Updated object 'v1.Deployment/") {
		return rc.Core.Write(entry, fields)
	}
	parts := strings.Split(entry.Message, "v1.Deployment/")
	if len(parts) == 2 {
		deploymentName := strings.TrimSuffix(parts[1], "'")
		if rc.state.deploymentNames[deploymentName] {
			key := rc.crKey()
			if key == "" {
				var ns, name string
				for _, f := range fields {
					if f.Type == zapcore.StringType {
						switch f.Key {
						case "namespace":
							ns = f.String
						case "name":
							name = f.String
						}
					}
				}
				if ns != "" && name != "" {
					key = ns + "/" + name
				}
			}
			if key != "" {
				rc.state.mu.Lock()
				if rc.state.updateCounts[key] == nil {
					rc.state.updateCounts[key] = make(map[string]int)
				}
				rc.state.updateCounts[key][deploymentName]++
				rc.state.mu.Unlock()
			}
		}
	}
	return rc.Core.Write(entry, fields)
}

// GetUpdateCounts returns a copy of per-deployment counts for the given CR instance.
func (rc *ReconcileCounter) GetUpdateCounts(namespace, name string) map[string]int {
	key := namespace + "/" + name
	rc.state.mu.Lock()
	defer rc.state.mu.Unlock()
	counts := make(map[string]int)
	for k, v := range rc.state.updateCounts[key] {
		counts[k] = v
	}
	return counts
}

// GetTotalUpdates returns the total deployment update count for the given CR instance.
func (rc *ReconcileCounter) GetTotalUpdates(namespace, name string) int {
	key := namespace + "/" + name
	rc.state.mu.Lock()
	defer rc.state.mu.Unlock()
	total := 0
	for _, count := range rc.state.updateCounts[key] {
		total += count
	}
	return total
}

// GetReport returns a formatted breakdown for the given CR instance.
func (rc *ReconcileCounter) GetReport(namespace, name string) string {
	key := namespace + "/" + name
	rc.state.mu.Lock()
	defer rc.state.mu.Unlock()
	counts := rc.state.updateCounts[key]
	if len(counts) == 0 {
		return fmt.Sprintf("No deployment updates detected for %s", key)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Deployment update counts for %s:\n", key)
	for depName, count := range counts {
		fmt.Fprintf(&sb, "  %s: %d updates\n", depName, count)
	}
	return sb.String()
}

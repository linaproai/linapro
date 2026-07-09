// This file defines normalized execution-source values shared by runtime,
// wasm, and host-side governance layers.

package contract

import "strings"

// ExecutionSource identifies what triggered one plugin execution.
type ExecutionSource string

// Execution source constants enumerate the stable execution origins used in
// audit, governance, and runtime flows.
const (
	// ExecutionSourceRoute marks one request-bound dynamic route execution.
	ExecutionSourceRoute ExecutionSource = "route"
	// ExecutionSourceHook marks one host hook callback execution.
	ExecutionSourceHook ExecutionSource = "hook"
	// ExecutionSourceJobs marks one scheduled job execution.
	ExecutionSourceJobs ExecutionSource = "jobs"
	// ExecutionSourceJobsDiscovery marks dynamic-plugin job declaration discovery.
	ExecutionSourceJobsDiscovery ExecutionSource = "jobs.discovery"
	// ExecutionSourceLifecycle marks one install/enable/disable lifecycle execution.
	ExecutionSourceLifecycle ExecutionSource = "lifecycle"
)

// NormalizeExecutionSource trims and lowercases one execution source value.
func NormalizeExecutionSource(source ExecutionSource) ExecutionSource {
	return ExecutionSource(strings.ToLower(strings.TrimSpace(string(source))))
}

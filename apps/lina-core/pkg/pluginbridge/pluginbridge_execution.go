// This file defines normalized execution-source values shared by runtime,
// wasm, and host-side governance layers.

package pluginbridge

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
	// ExecutionSourceCron marks one scheduled job execution.
	ExecutionSourceCron ExecutionSource = "cron"
	// ExecutionSourceCronDiscovery marks one host-driven cron declaration
	// discovery call against a dynamic plugin runtime.
	ExecutionSourceCronDiscovery ExecutionSource = "cron_discovery"
	// ExecutionSourceLifecycle marks one install/enable/disable lifecycle execution.
	ExecutionSourceLifecycle ExecutionSource = "lifecycle"
)

// NormalizeExecutionSource trims and lowercases one execution source value.
func NormalizeExecutionSource(source ExecutionSource) ExecutionSource {
	return ExecutionSource(strings.ToLower(strings.TrimSpace(string(source))))
}

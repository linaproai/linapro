// This file defines normalized execution-source values shared by runtime,
// wasm, and host-side governance layers.

package pluginbridge

import "strings"

// ExecutionSource identifies what triggered one plugin execution.
type ExecutionSource string

const (
	// ExecutionSourceRoute marks one request-bound dynamic route execution.
	ExecutionSourceRoute ExecutionSource = "route"
	// ExecutionSourceHook marks one host hook callback execution.
	ExecutionSourceHook ExecutionSource = "hook"
	// ExecutionSourceCron marks one scheduled job execution.
	ExecutionSourceCron ExecutionSource = "cron"
	// ExecutionSourceLifecycle marks one install/enable/disable lifecycle execution.
	ExecutionSourceLifecycle ExecutionSource = "lifecycle"
)

// NormalizeExecutionSource trims and lowercases one execution source value.
func NormalizeExecutionSource(source ExecutionSource) ExecutionSource {
	return ExecutionSource(strings.ToLower(strings.TrimSpace(string(source))))
}

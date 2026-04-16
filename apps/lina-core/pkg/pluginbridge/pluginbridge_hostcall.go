// This file defines the low-level host call entrypoint constants, status
// codes, and shared runtime values used by both host and guest code.

package pluginbridge

const (
	// HostModuleName is the wazero host module namespace for Lina host functions.
	HostModuleName = "lina_env"
	// HostCallFunctionName is the single host call dispatch function name.
	HostCallFunctionName = "host_call"

	// DefaultGuestHostCallAllocExport is the guest export used by the host to
	// allocate response buffers during host call processing.
	DefaultGuestHostCallAllocExport = "lina_host_call_alloc"
)

const (
	// HostCallStatusSuccess indicates the host call completed successfully.
	HostCallStatusSuccess uint32 = 0
	// HostCallStatusCapabilityDenied indicates the plugin lacks capability or authorization.
	HostCallStatusCapabilityDenied uint32 = 1
	// HostCallStatusNotFound indicates an unknown opcode, service, or method.
	HostCallStatusNotFound uint32 = 2
	// HostCallStatusInvalidRequest indicates a malformed request payload.
	HostCallStatusInvalidRequest uint32 = 3
	// HostCallStatusInternalError indicates a host-side processing failure.
	HostCallStatusInternalError uint32 = 4
)

const (
	// LogLevelDebug maps to logger.Debug.
	LogLevelDebug int32 = 1
	// LogLevelInfo maps to logger.Info.
	LogLevelInfo int32 = 2
	// LogLevelWarning maps to logger.Warning.
	LogLevelWarning int32 = 3
	// LogLevelError maps to logger.Error.
	LogLevelError int32 = 4
)

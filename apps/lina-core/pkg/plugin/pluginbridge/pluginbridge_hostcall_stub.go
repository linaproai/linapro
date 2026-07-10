//go:build !wasip1

// This file provides the host-build raw guest host-service transport stub.
// Public clients are constructed by pluginbridge_hostcall_clients.go and all
// fail through this single transport entry when not running as a WASI guest.

package pluginbridge

// InvokeHostService reports that generic guest host calls are unavailable.
func InvokeHostService(_ string, _ string, _ string, _ string, _ []byte) ([]byte, error) {
	return nil, ErrHostCallsUnavailable
}

// InvokeOwnerHostService reports that owner-aware guest host calls are
// unavailable outside WASI guest builds.
func InvokeOwnerHostService(_ string, _ string, _ string, _ string, _ string, _ string, _ []byte) ([]byte, error) {
	return nil, ErrHostCallsUnavailable
}

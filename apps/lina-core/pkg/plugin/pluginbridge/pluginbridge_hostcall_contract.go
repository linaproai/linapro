// This file defines the shared guest host-service client contracts used by
// both wasip1 host-call implementations and non-WASI unsupported stubs.

package pluginbridge

import (
	"time"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// RuntimeHostService exposes guest-side helpers for the runtime host service.
type RuntimeHostService interface {
	// Log writes one structured runtime log entry through the host.
	Log(level int, message string, fields map[string]string) error
	// StateGet reads one plugin-scoped runtime state value by key.
	StateGet(key string) (string, bool, error)
	// StateGetMany reads plugin-scoped runtime state values by key.
	StateGetMany(keys []string) (map[string]string, []string, error)
	// StateSet writes one plugin-scoped runtime state value.
	StateSet(key string, value string) error
	// StateSetMany writes plugin-scoped runtime state values.
	StateSetMany(values map[string]string) error
	// StateDelete removes one plugin-scoped runtime state value.
	StateDelete(key string) error
	// StateDeleteMany removes plugin-scoped runtime state values.
	StateDeleteMany(keys []string) error
	// StateGetInt reads one integer runtime state value.
	StateGetInt(key string) (int, bool, error)
	// StateSetInt writes one integer runtime state value.
	StateSetInt(key string, value int) error
	// Now returns the current host time string.
	Now() (string, error)
	// UUID returns one host-generated unique identifier string.
	UUID() (string, error)
	// Node returns the current host node identity string.
	Node() (string, error)
}

// NetworkHostService exposes guest-side helpers for the governed outbound network host service.
type NetworkHostService interface {
	// Request executes one governed outbound HTTP request through the host.
	Request(targetURL string, request *protocol.HostServiceNetworkRequest) (*protocol.HostServiceNetworkResponse, error)
}

// HostConfigHostService exposes guest-side helpers for authorized host config reads.
type HostConfigHostService interface {
	// Get reads one authorized host config value as JSON.
	Get(key string) (string, bool, error)
	// String reads one authorized host config value as a string.
	String(key string) (string, bool, error)
	// Bool reads one authorized host config value as a bool.
	Bool(key string) (bool, bool, error)
	// Int reads one authorized host config value as an int.
	Int(key string) (int, bool, error)
	// Duration reads one authorized host config value as a duration.
	Duration(key string) (time.Duration, bool, error)
}

// ManifestHostService exposes guest-side helpers for plugin-scoped manifest resources.
type ManifestHostService interface {
	// Get reads one manifest resource as bytes.
	Get(path string) ([]byte, bool, error)
	// GetText reads one manifest resource as UTF-8 text.
	GetText(path string) (string, bool, error)
	// Scan decodes a YAML manifest resource or nested key into target.
	Scan(path string, key string, target any) (bool, error)
}

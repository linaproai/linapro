// This file defines the dynamic-only guest host-service client contracts used
// by the Services directory for host-service families that do not have a
// shared capability package contract.

package pluginbridge

import "lina-core/pkg/plugin/pluginbridge/protocol"

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

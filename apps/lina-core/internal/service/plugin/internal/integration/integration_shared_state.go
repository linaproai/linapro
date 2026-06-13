// This file centralizes the in-memory integration runtime state that must stay
// consistent across multiple plugin service instances created inside one host
// process.

package integration

import (
	"sync"

	"lina-core/pkg/plugin/pluginhost"
)

// SharedState stores process-wide integration caches used by source-plugin
// route guards and route-binding projections.
type SharedState struct {
	sourceRouteBindingsMu sync.RWMutex
	sourceRouteBindings   map[string][]pluginhost.SourceRouteBinding

	enabledSnapshotMu     sync.RWMutex
	enabledSnapshot       map[string]bool
	enabledSnapshotLoaded bool
}

// NewSharedState creates an integration state holder for one host composition
// root. Pass the same instance to integration services that must share
// source-plugin route guards and enabled snapshots in the current process.
func NewSharedState() *SharedState {
	return &SharedState{
		sourceRouteBindings: make(map[string][]pluginhost.SourceRouteBinding),
		enabledSnapshot:     make(map[string]bool),
	}
}

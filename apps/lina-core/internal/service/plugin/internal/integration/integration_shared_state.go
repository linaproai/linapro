// This file centralizes the in-memory integration runtime state that must stay
// consistent across multiple plugin service instances created inside one host
// process.

package integration

import (
	"sync"

	"lina-core/pkg/pluginhost"
)

// sharedState stores process-wide integration caches used by source-plugin
// route guards and route-binding projections.
type sharedState struct {
	sourceRouteBindingsMu sync.RWMutex
	sourceRouteBindings   map[string][]pluginhost.SourceRouteBinding

	enabledSnapshotMu     sync.RWMutex
	enabledSnapshot       map[string]bool
	enabledSnapshotLoaded bool
}

// defaultSharedState is reused by all integration service instances so plugin
// lifecycle API calls and HTTP route guards observe the same enablement state.
var defaultSharedState = &sharedState{
	sourceRouteBindings: make(map[string][]pluginhost.SourceRouteBinding),
	enabledSnapshot:     make(map[string]bool),
}

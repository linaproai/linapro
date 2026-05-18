// Package wasm implements the low-level wazero WASM bridge used by dynamic route
// execution. It manages module compilation caching, host call registration, and
// the alloc→write→execute→read ABI protocol shared with guest plugins.
package wasm

import (
	"sync"

	"github.com/tetratelabs/wazero"

	bridgecontract "lina-core/pkg/pluginbridge/contract"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
)

// ExecutionInput carries the minimum manifest data needed to run one bridge call.
type ExecutionInput struct {
	// PluginID identifies the calling plugin for host function context.
	PluginID string
	// ArtifactPath is the filesystem path to the compiled wasm artifact.
	ArtifactPath string
	// BridgeSpec carries the guest-exported function names for the bridge ABI.
	BridgeSpec *bridgecontract.BridgeSpec
	// Capabilities is the set of host capabilities granted to this plugin.
	Capabilities map[string]struct{}
	// HostServices is the structured host service authorization snapshot for this plugin.
	HostServices []*bridgehostservice.HostServiceSpec
	// ExecutionSource identifies what triggered this bridge execution.
	ExecutionSource bridgecontract.ExecutionSource
	// RoutePath is the matched dynamic route path when execution is route-bound.
	RoutePath string
	// RequestID is the host-generated request identifier for this execution.
	RequestID string
	// Identity carries the sanitized user identity snapshot when available.
	Identity *bridgecontract.IdentitySnapshotV1
	// CronCollector receives dynamic-plugin cron registrations during reserved
	// discovery executions.
	CronCollector CronRegistrationCollector
}

// wasmCacheEntry stores one compiled module together with the runtime that owns it.
type wasmCacheEntry struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
}

// wasmModuleCache caches compiled Wasm modules keyed by the archived active
// artifact path. Dynamic release archive paths include the release checksum, so
// same-version refreshes naturally compile a separate module.
var (
	wasmModuleCacheMu sync.RWMutex
	wasmModuleCache   = make(map[string]*wasmCacheEntry)
)

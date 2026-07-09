// Package wasm implements the low-level wazero WASM bridge used by dynamic route
// execution. It manages module compilation caching, host call registration, and
// the alloc→write→execute→read ABI protocol shared with guest plugins.
package wasm

import (
	"context"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"

	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

const (
	// defaultBridgeExecutionTimeout is the host-side fallback for dynamic
	// plugin bridge calls when callers do not provide a tighter deadline.
	defaultBridgeExecutionTimeout = 30 * time.Second
	// defaultWasmMemoryLimitPages caps one dynamic plugin module instance to
	// 256 MiB because each WebAssembly memory page is 64 KiB.
	defaultWasmMemoryLimitPages uint32 = 4096
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
	// ArtifactDefaultConfig is the active-release manifest/config/config.yaml
	// content used as the lowest-priority plugin config source.
	ArtifactDefaultConfig []byte
	// ArtifactManifestResources stores active-release manifest resources keyed
	// relative to manifest/.
	ArtifactManifestResources map[string][]byte
	// ExecutionSource identifies what triggered this bridge execution.
	ExecutionSource bridgecontract.ExecutionSource
	// RoutePath is the matched dynamic route path when execution is route-bound.
	RoutePath string
	// RequestID is the host-generated request identifier for this execution.
	RequestID string
	// Identity carries the sanitized user identity snapshot when available.
	Identity *bridgecontract.IdentitySnapshotV1
	// JobCollector captures dynamic-plugin job declarations during Jobs discovery.
	JobCollector JobRegistrationCollector
}

// Runtime defines the dynamic-plugin WASM execution runtime owned by a plugin
// service instance.
type Runtime interface {
	// ExecuteBridge executes one bridge request against the archived active WASM
	// artifact using this runtime instance.
	ExecuteBridge(
		ctx context.Context,
		input ExecutionInput,
		requestContent []byte,
	) (*bridgecontract.BridgeResponseEnvelopeV1, error)
	// InvalidateCache removes the cached compiled module for the given artifact path.
	InvalidateCache(ctx context.Context, artifactPath string)
	// InvalidateAllCache removes all cached compiled modules owned by this runtime.
	InvalidateAllCache(ctx context.Context)
}

// runtimeImpl owns host-service dependencies and compiled module cache state for
// one plugin service instance.
type runtimeImpl struct {
	hostServices *hostServiceRuntime
	cacheMu      sync.RWMutex
	cache        map[string]*wasmCacheEntry
	inflight     map[string]*wasmCompileInflight
	compileHook  func(string)
}

// wasmCacheEntry stores one compiled module together with the runtime that owns
// it. The entry tracks active execution leases so cache invalidation can remove
// stale entries immediately without closing a runtime that is still instantiating
// or executing a guest module.
type wasmCacheEntry struct {
	mu          sync.Mutex
	idle        *sync.Cond
	runtime     wazero.Runtime
	compiled    wazero.CompiledModule
	active      int
	invalidated bool
	closed      bool
}

// wasmCompileInflight coordinates one artifact compilation outside the global
// cache lock so concurrent requests share the same result.
type wasmCompileInflight struct {
	done  chan struct{}
	entry *wasmCacheEntry
	err   error
}

// wasmModuleLease pins a cached module entry while a bridge execution uses its
// runtime and compiled module.
type wasmModuleLease struct {
	entry    *wasmCacheEntry
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
}

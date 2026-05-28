// This file manages cached compiled Wasm modules keyed by archived runtime
// artifact paths.

package wasm

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"lina-core/pkg/logger"
)

// InvalidateCache removes the cached compiled module for the given artifact path.
// This must be called when a plugin's active release changes (upgrade, rollback,
// uninstall) so subsequent requests recompile from the new artifact.
func InvalidateCache(ctx context.Context, artifactPath string) {
	if ctx == nil {
		ctx = context.Background()
	}
	artifactPath = strings.TrimSpace(artifactPath)
	if artifactPath == "" {
		return
	}

	var entry *wasmCacheEntry
	wasmModuleCacheMu.Lock()
	if cached, ok := wasmModuleCache[artifactPath]; ok {
		entry = cached
		delete(wasmModuleCache, artifactPath)
	}
	wasmModuleCacheMu.Unlock()
	closeInvalidatedWasmCacheEntry(ctx, artifactPath, entry)
}

// InvalidateAllCache removes all cached compiled modules. This is useful during
// full reconciliation passes or shutdown.
func InvalidateAllCache(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	wasmModuleCacheMu.Lock()
	entries := make(map[string]*wasmCacheEntry, len(wasmModuleCache))
	for path, entry := range wasmModuleCache {
		entries[path] = entry
		delete(wasmModuleCache, path)
	}
	wasmModuleCacheMu.Unlock()
	for path, entry := range entries {
		closeInvalidatedWasmCacheEntry(ctx, path, entry)
	}
}

// getOrCompileWasmModule returns a lease for the cached compiled module or
// compiles it from disk.
func getOrCompileWasmModule(ctx context.Context, artifactPath string) (*wasmModuleLease, error) {
	wasmModuleCacheMu.RLock()
	if entry, ok := wasmModuleCache[artifactPath]; ok {
		wasmModuleCacheMu.RUnlock()
		if lease := entry.acquireLease(); lease != nil {
			return lease, nil
		}
	} else {
		wasmModuleCacheMu.RUnlock()
	}

	wasmModuleCacheMu.Lock()
	defer wasmModuleCacheMu.Unlock()

	// Double-check after acquiring write lock.
	if entry, ok := wasmModuleCache[artifactPath]; ok {
		if lease := entry.acquireLease(); lease != nil {
			return lease, nil
		}
		delete(wasmModuleCache, artifactPath)
	}

	rt := newWasmRuntime(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after WASI init failure failed err=%v", closeErr)
		}
		return nil, gerror.Wrap(err, "initialize WASI failed")
	}

	// Register host call module so guest imports are satisfied at compile time.
	if err := registerHostCallModule(ctx, rt); err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after host-call registration failure failed err=%v", closeErr)
		}
		return nil, gerror.Wrap(err, "register host call module failed")
	}

	wasmBytes, err := os.ReadFile(artifactPath)
	if err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after artifact read failure failed err=%v", closeErr)
		}
		return nil, gerror.Wrap(err, "read dynamic plugin Wasm artifact failed")
	}
	compiled, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after compile failure failed err=%v", closeErr)
		}
		return nil, gerror.Wrap(err, "compile dynamic plugin Wasm failed")
	}

	entry := &wasmCacheEntry{
		runtime:  rt,
		compiled: compiled,
	}
	entry.idle = sync.NewCond(&entry.mu)
	wasmModuleCache[artifactPath] = entry
	lease := entry.acquireLease()
	if lease == nil {
		delete(wasmModuleCache, artifactPath)
		return nil, gerror.New("compiled dynamic plugin Wasm cache entry is not available")
	}
	return lease, nil
}

// newWasmRuntime creates one host-governed runtime for dynamic plugin modules.
// Context cancellation is wired into guest execution so bridge deadlines can
// stop non-returning guest code, while memory pages are capped for every module.
func newWasmRuntime(ctx context.Context) wazero.Runtime {
	config := wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).
		WithMemoryLimitPages(defaultWasmMemoryLimitPages)
	return wazero.NewRuntimeWithConfig(ctx, config)
}

// acquireLease pins the entry for one bridge execution.
func (entry *wasmCacheEntry) acquireLease() *wasmModuleLease {
	entry.mu.Lock()
	defer entry.mu.Unlock()
	if entry.invalidated || entry.closed {
		return nil
	}
	entry.active++
	return &wasmModuleLease{
		entry:    entry,
		runtime:  entry.runtime,
		compiled: entry.compiled,
	}
}

// Release unpins the cached runtime after one bridge execution.
func (lease *wasmModuleLease) Release() {
	if lease == nil || lease.entry == nil {
		return
	}
	entry := lease.entry
	entry.mu.Lock()
	if entry.active > 0 {
		entry.active--
	}
	if entry.active == 0 && entry.idle != nil {
		entry.idle.Broadcast()
	}
	entry.mu.Unlock()
	lease.entry = nil
}

// closeInvalidatedWasmCacheEntry waits for active leases before closing one
// stale runtime. New requests can compile a fresh entry while this wait happens.
func closeInvalidatedWasmCacheEntry(ctx context.Context, artifactPath string, entry *wasmCacheEntry) {
	if entry == nil {
		return
	}
	entry.mu.Lock()
	if entry.closed {
		entry.mu.Unlock()
		return
	}
	entry.invalidated = true
	for entry.active > 0 {
		entry.idle.Wait()
	}
	entry.closed = true
	runtime := entry.runtime
	entry.mu.Unlock()

	if err := runtime.Close(ctx); err != nil {
		logger.Warningf(ctx, "close cached wasm runtime failed artifactPath=%s err=%v", artifactPath, err)
	}
}

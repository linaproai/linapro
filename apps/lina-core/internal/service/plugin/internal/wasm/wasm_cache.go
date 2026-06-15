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
func (r *runtimeImpl) InvalidateCache(ctx context.Context, artifactPath string) {
	if r == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	artifactPath = strings.TrimSpace(artifactPath)
	if artifactPath == "" {
		return
	}

	var entry *wasmCacheEntry
	r.cacheMu.Lock()
	if cached, ok := r.cache[artifactPath]; ok {
		entry = cached
		delete(r.cache, artifactPath)
	}
	r.cacheMu.Unlock()
	closeInvalidatedWasmCacheEntry(ctx, artifactPath, entry)
}

// InvalidateAllCache removes all cached compiled modules. This is useful during
// full reconciliation passes or shutdown.
func (r *runtimeImpl) InvalidateAllCache(ctx context.Context) {
	if r == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	r.cacheMu.Lock()
	entries := make(map[string]*wasmCacheEntry, len(r.cache))
	for path, entry := range r.cache {
		entries[path] = entry
		delete(r.cache, path)
	}
	r.cacheMu.Unlock()
	for path, entry := range entries {
		closeInvalidatedWasmCacheEntry(ctx, path, entry)
	}
}

// getOrCompileWasmModule returns a lease for the cached compiled module or
// compiles it from disk.
func (r *runtimeImpl) getOrCompileWasmModule(ctx context.Context, artifactPath string) (*wasmModuleLease, error) {
	if r == nil {
		return nil, gerror.New("wasm runtime is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	r.cacheMu.RLock()
	if entry, ok := r.cache[artifactPath]; ok {
		r.cacheMu.RUnlock()
		if lease := entry.acquireLease(); lease != nil {
			return lease, nil
		}
	} else {
		r.cacheMu.RUnlock()
	}

	for {
		inflight, owner := r.getOrCreateCompileInflight(artifactPath)
		if owner {
			entry, err := r.compileWasmCacheEntry(ctx, artifactPath)
			r.finishCompileInflight(artifactPath, inflight, entry, err)
			if err != nil {
				return nil, err
			}
			if lease := entry.acquireLease(); lease != nil {
				return lease, nil
			}
			return nil, gerror.New("compiled dynamic plugin Wasm cache entry is not available")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-inflight.done:
		}
		if inflight.err != nil {
			return nil, inflight.err
		}
		if inflight.entry != nil {
			if lease := inflight.entry.acquireLease(); lease != nil {
				return lease, nil
			}
		}
	}
}

// getOrCreateCompileInflight returns the in-flight compilation for artifactPath.
// The owner return value identifies the caller responsible for performing the
// compile outside the global cache lock.
func (r *runtimeImpl) getOrCreateCompileInflight(artifactPath string) (*wasmCompileInflight, bool) {
	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()
	if r.cache == nil {
		r.cache = make(map[string]*wasmCacheEntry)
	}
	if entry, ok := r.cache[artifactPath]; ok {
		if lease := entry.acquireLease(); lease != nil {
			lease.Release()
			return &wasmCompileInflight{done: closedCompileInflightDone(), entry: entry}, false
		}
		delete(r.cache, artifactPath)
	}
	if r.inflight == nil {
		r.inflight = make(map[string]*wasmCompileInflight)
	}
	if inflight, ok := r.inflight[artifactPath]; ok {
		return inflight, false
	}
	inflight := &wasmCompileInflight{done: make(chan struct{})}
	r.inflight[artifactPath] = inflight
	return inflight, true
}

// compileWasmCacheEntry reads and compiles one artifact without holding
// runtimeImpl.cacheMu.
func (r *runtimeImpl) compileWasmCacheEntry(ctx context.Context, artifactPath string) (*wasmCacheEntry, error) {
	if r.compileHook != nil {
		r.compileHook(artifactPath)
	}
	rt := newWasmRuntime(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after WASI init failure failed err=%v", closeErr)
		}
		return nil, gerror.Wrap(err, "initialize WASI failed")
	}

	// Register host call module so guest imports are satisfied at compile time.
	if err := r.registerHostCallModule(ctx, rt); err != nil {
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
	return entry, nil
}

// finishCompileInflight stores a successful compile result and wakes waiters.
func (r *runtimeImpl) finishCompileInflight(
	artifactPath string,
	inflight *wasmCompileInflight,
	entry *wasmCacheEntry,
	err error,
) {
	r.cacheMu.Lock()
	if r.cache == nil {
		r.cache = make(map[string]*wasmCacheEntry)
	}
	if err == nil && entry != nil {
		r.cache[artifactPath] = entry
	}
	if inflight != nil {
		inflight.entry = entry
		inflight.err = err
		close(inflight.done)
	}
	delete(r.inflight, artifactPath)
	r.cacheMu.Unlock()
}

// closedCompileInflightDone returns a closed channel for already-cached entries.
func closedCompileInflightDone() chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
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

// This file manages cached compiled Wasm modules keyed by archived runtime
// artifact paths.

package wasm

import (
	"context"
	"os"
	"strings"

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
	wasmModuleCacheMu.Lock()
	defer wasmModuleCacheMu.Unlock()
	if entry, ok := wasmModuleCache[artifactPath]; ok {
		if err := entry.runtime.Close(ctx); err != nil {
			logger.Warningf(ctx, "close cached wasm runtime failed artifactPath=%s err=%v", artifactPath, err)
		}
		delete(wasmModuleCache, artifactPath)
	}
}

// InvalidateAllCache removes all cached compiled modules. This is useful during
// full reconciliation passes or shutdown.
func InvalidateAllCache(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	wasmModuleCacheMu.Lock()
	defer wasmModuleCacheMu.Unlock()
	for path, entry := range wasmModuleCache {
		if err := entry.runtime.Close(ctx); err != nil {
			logger.Warningf(ctx, "close cached wasm runtime failed artifactPath=%s err=%v", path, err)
		}
		delete(wasmModuleCache, path)
	}
}

// getOrCompileWasmModule returns the cached compiled module or compiles it from disk.
func getOrCompileWasmModule(ctx context.Context, artifactPath string) (wazero.Runtime, wazero.CompiledModule, error) {
	wasmModuleCacheMu.RLock()
	if entry, ok := wasmModuleCache[artifactPath]; ok {
		wasmModuleCacheMu.RUnlock()
		return entry.runtime, entry.compiled, nil
	}
	wasmModuleCacheMu.RUnlock()

	wasmModuleCacheMu.Lock()
	defer wasmModuleCacheMu.Unlock()

	// Double-check after acquiring write lock.
	if entry, ok := wasmModuleCache[artifactPath]; ok {
		return entry.runtime, entry.compiled, nil
	}

	rt := wazero.NewRuntime(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after WASI init failure failed err=%v", closeErr)
		}
		return nil, nil, gerror.Wrap(err, "initialize WASI failed")
	}

	// Register host call module so guest imports are satisfied at compile time.
	if err := registerHostCallModule(ctx, rt); err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after host-call registration failure failed err=%v", closeErr)
		}
		return nil, nil, gerror.Wrap(err, "register host call module failed")
	}

	wasmBytes, err := os.ReadFile(artifactPath)
	if err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after artifact read failure failed err=%v", closeErr)
		}
		return nil, nil, gerror.Wrap(err, "read dynamic plugin Wasm artifact failed")
	}
	compiled, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after compile failure failed err=%v", closeErr)
		}
		return nil, nil, gerror.Wrap(err, "compile dynamic plugin Wasm failed")
	}

	wasmModuleCache[artifactPath] = &wasmCacheEntry{
		runtime:  rt,
		compiled: compiled,
	}
	return rt, compiled, nil
}

// Package wasm implements the low-level wazero WASM bridge used by dynamic route
// execution. It manages module compilation caching, host call registration, and
// the alloc→write→execute→read ABI protocol shared with guest plugins.
package wasm

import (
	"context"
	"os"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"lina-core/pkg/logger"
	"lina-core/pkg/pluginbridge"
)

// ExecutionInput carries the minimum manifest data needed to run one bridge call.
type ExecutionInput struct {
	// PluginID identifies the calling plugin for host function context.
	PluginID string
	// ArtifactPath is the filesystem path to the compiled wasm artifact.
	ArtifactPath string
	// BridgeSpec carries the guest-exported function names for the bridge ABI.
	BridgeSpec *pluginbridge.BridgeSpec
	// Capabilities is the set of host capabilities granted to this plugin.
	Capabilities map[string]struct{}
	// HostServices is the structured host service authorization snapshot for this plugin.
	HostServices []*pluginbridge.HostServiceSpec
	// ExecutionSource identifies what triggered this bridge execution.
	ExecutionSource pluginbridge.ExecutionSource
	// RoutePath is the matched dynamic route path when execution is route-bound.
	RoutePath string
	// RequestID is the host-generated request identifier for this execution.
	RequestID string
	// Identity carries the sanitized user identity snapshot when available.
	Identity *pluginbridge.IdentitySnapshotV1
}

// wasmCacheEntry holds a pre-compiled Wasm module bound to its wazero runtime.
// The compiled module can be instantiated multiple times for concurrent requests
// while the runtime manages the underlying compilation cache.
type wasmCacheEntry struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
}

// wasmModuleCache caches compiled Wasm modules keyed by artifact path so that
// repeated bridge invocations against the same active release skip disk I/O and
// compilation. Entries are evicted when the active release changes.
var (
	wasmModuleCacheMu sync.RWMutex
	wasmModuleCache   = make(map[string]*wasmCacheEntry)
)

// InvalidateCache removes the cached compiled module for the given artifact path.
// This must be called when a plugin's active release changes (upgrade, rollback,
// uninstall) so subsequent requests recompile from the new artifact.
func InvalidateCache(artifactPath string) {
	wasmModuleCacheMu.Lock()
	defer wasmModuleCacheMu.Unlock()
	if entry, ok := wasmModuleCache[artifactPath]; ok {
		if err := entry.runtime.Close(context.Background()); err != nil {
			logger.Warningf(context.Background(), "close cached wasm runtime failed artifactPath=%s err=%v", artifactPath, err)
		}
		delete(wasmModuleCache, artifactPath)
	}
}

// InvalidateAllCache removes all cached compiled modules. This is useful during
// full reconciliation passes or shutdown.
func InvalidateAllCache() {
	wasmModuleCacheMu.Lock()
	defer wasmModuleCacheMu.Unlock()
	for path, entry := range wasmModuleCache {
		if err := entry.runtime.Close(context.Background()); err != nil {
			logger.Warningf(context.Background(), "close cached wasm runtime failed artifactPath=%s err=%v", path, err)
		}
		delete(wasmModuleCache, path)
	}
}

// ExecuteBridge runs one bridge invocation against the archived active Wasm
// artifact using the alloc→write→execute→read protocol defined by the shared
// bridge ABI. It reuses cached compiled modules across concurrent requests.
func ExecuteBridge(
	ctx context.Context,
	input ExecutionInput,
	requestContent []byte,
) (response *pluginbridge.BridgeResponseEnvelopeV1, err error) {
	if input.BridgeSpec == nil {
		return nil, gerror.New("动态插件缺少 Wasm bridge 元数据")
	}

	rt, compiled, err := getOrCompileWasmModule(ctx, input.ArtifactPath)
	if err != nil {
		return nil, err
	}

	// Each request gets a fresh module instance because guest globals (request
	// and response buffers) are mutable and must not be shared across requests.
	module, err := rt.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName("").WithStartFunctions())
	if err != nil {
		return nil, gerror.Wrap(err, "实例化动态插件 Wasm 失败")
	}
	defer func() {
		if closeErr := module.Close(ctx); closeErr != nil && err == nil {
			err = gerror.Wrap(closeErr, "关闭动态插件 Wasm 模块失败")
		}
	}()

	// Inject host call context so that host function callbacks can access
	// plugin identity and capabilities.
	ctx = withHostCallContext(ctx, &hostCallContext{
		pluginID:        input.PluginID,
		capabilities:    input.Capabilities,
		hostServices:    input.HostServices,
		executionSource: input.ExecutionSource,
		routePath:       input.RoutePath,
		requestID:       input.RequestID,
		identity:        input.Identity,
	})

	var (
		allocFn      = module.ExportedFunction(input.BridgeSpec.AllocExport)
		executeFn    = module.ExportedFunction(input.BridgeSpec.ExecuteExport)
		initializeFn = module.ExportedFunction("_initialize")
	)
	if allocFn == nil || executeFn == nil {
		return nil, gerror.New("动态插件 Wasm bridge 缺少必需导出函数")
	}
	if initializeFn != nil {
		// `_initialize` is optional and is only invoked when guest toolchains emit
		// it, keeping the host compatible with both reactor and non-reactor builds.
		if _, err := initializeFn.Call(ctx); err != nil {
			return nil, gerror.Wrap(err, "初始化动态插件 Wasm 运行时失败")
		}
	}

	// The bridge ABI protocol is: alloc(size) → host writes to returned pointer →
	// execute(size). The guest's execute reads from the same global buffer that
	// alloc exposed, so only the payload length needs to be passed to execute.
	allocResult, err := allocFn.Call(ctx, uint64(len(requestContent)))
	if err != nil {
		return nil, gerror.Wrap(err, "调用动态插件 alloc 失败")
	}
	if len(allocResult) == 0 {
		return nil, gerror.New("动态插件 alloc 未返回有效指针")
	}
	requestPointer := uint32(allocResult[0])
	if ok := module.Memory().Write(requestPointer, requestContent); !ok {
		return nil, gerror.New("写入动态插件请求内存失败")
	}

	// Execute returns one packed pointer/length pair so the host can read the
	// response bytes without any JSON or text-based marshaling layer.
	executeResult, err := executeFn.Call(ctx, uint64(len(requestContent)))
	if err != nil {
		return nil, gerror.Wrap(err, "调用动态插件 execute 失败")
	}
	if len(executeResult) == 0 {
		return nil, gerror.New("动态插件 execute 未返回有效响应")
	}
	responsePointer, responseLength := decodeDynamicResponsePointer(executeResult[0])
	responseContent, ok := module.Memory().Read(responsePointer, responseLength)
	if !ok {
		return nil, gerror.New("读取动态插件响应内存失败")
	}
	response, err = pluginbridge.DecodeResponseEnvelope(responseContent)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// getOrCompileWasmModule returns a cached compiled module or compiles a new one
// from disk and caches it for future requests.
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
		return nil, nil, gerror.Wrap(err, "初始化 WASI 失败")
	}

	// Register host call module so guest imports are satisfied at compile time.
	if err := registerHostCallModule(ctx, rt); err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after host-call registration failure failed err=%v", closeErr)
		}
		return nil, nil, gerror.Wrap(err, "注册宿主调用模块失败")
	}

	wasmBytes, err := os.ReadFile(artifactPath)
	if err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after artifact read failure failed err=%v", closeErr)
		}
		return nil, nil, gerror.Wrap(err, "读取动态插件 Wasm 产物失败")
	}
	compiled, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		if closeErr := rt.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "close wasm runtime after compile failure failed err=%v", closeErr)
		}
		return nil, nil, gerror.Wrap(err, "编译动态插件 Wasm 失败")
	}

	wasmModuleCache[artifactPath] = &wasmCacheEntry{
		runtime:  rt,
		compiled: compiled,
	}
	return rt, compiled, nil
}

// decodeDynamicResponsePointer unpacks the bridge ABI return value where the
// high 32 bits are the response pointer and the low 32 bits are the byte length.
func decodeDynamicResponsePointer(value uint64) (uint32, uint32) {
	return uint32(value >> 32), uint32(value & 0xffffffff)
}

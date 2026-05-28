// This file executes dynamic plugin bridge requests through the Wasm
// alloc/write/execute/read ABI.

package wasm

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/tetratelabs/wazero"

	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	bridgecodec "lina-core/pkg/plugin/pluginbridge/protocol"
)

// ExecuteBridge executes one bridge request against the archived active wasm
// artifact using the alloc/write/execute/read ABI sequence.
func ExecuteBridge(
	ctx context.Context,
	input ExecutionInput,
	requestContent []byte,
) (response *bridgecontract.BridgeResponseEnvelopeV1, err error) {
	ctx, cancel := bridgeExecutionContext(ctx)
	if cancel != nil {
		defer cancel()
	}

	if input.BridgeSpec == nil {
		return nil, gerror.New("dynamic plugin is missing Wasm bridge metadata")
	}

	lease, err := getOrCompileWasmModule(ctx, input.ArtifactPath)
	if err != nil {
		return nil, err
	}
	defer lease.Release()

	// Each request gets a fresh module instance because guest globals (request
	// and response buffers) are mutable and must not be shared across requests.
	module, err := lease.runtime.InstantiateModule(ctx, lease.compiled, wazero.NewModuleConfig().WithName("").WithStartFunctions())
	if err != nil {
		return nil, gerror.Wrap(err, "instantiate dynamic plugin Wasm failed")
	}
	defer func() {
		closeCtx := context.WithoutCancel(ctx)
		if closeErr := module.Close(closeCtx); closeErr != nil && err == nil {
			err = gerror.Wrap(closeErr, "close dynamic plugin Wasm module failed")
		}
	}()

	// Inject host call context so that host function callbacks can access
	// plugin identity and capabilities.
	ctx = withHostCallContext(ctx, &hostCallContext{
		pluginID:                  input.PluginID,
		capabilities:              input.Capabilities,
		hostServices:              input.HostServices,
		artifactDefaultConfig:     append([]byte(nil), input.ArtifactDefaultConfig...),
		artifactManifestResources: cloneExecutionManifestResources(input.ArtifactManifestResources),
		executionSource:           input.ExecutionSource,
		routePath:                 input.RoutePath,
		requestID:                 input.RequestID,
		identity:                  input.Identity,
		cronCollector:             input.CronCollector,
	})

	var (
		allocFn      = module.ExportedFunction(input.BridgeSpec.AllocExport)
		executeFn    = module.ExportedFunction(input.BridgeSpec.ExecuteExport)
		initializeFn = module.ExportedFunction("_initialize")
	)
	if allocFn == nil || executeFn == nil {
		return nil, gerror.New("dynamic plugin Wasm bridge is missing required exported functions")
	}
	if initializeFn != nil {
		// `_initialize` is optional and is only invoked when guest toolchains emit
		// it, keeping the host compatible with both reactor and non-reactor builds.
		if _, err := initializeFn.Call(ctx); err != nil {
			return nil, gerror.Wrap(err, "initialize dynamic plugin Wasm runtime failed")
		}
	}

	// The bridge ABI protocol is: alloc(size) -> host writes to returned pointer ->
	// execute(size). The guest's execute reads from the same global buffer that
	// alloc exposed, so only the payload length needs to be passed to execute.
	allocResult, err := allocFn.Call(ctx, uint64(len(requestContent)))
	if err != nil {
		return nil, gerror.Wrap(err, "call dynamic plugin alloc failed")
	}
	if len(allocResult) == 0 {
		return nil, gerror.New("dynamic plugin alloc returned no valid pointer")
	}
	requestPointer := uint32(allocResult[0])
	if ok := module.Memory().Write(requestPointer, requestContent); !ok {
		return nil, gerror.New("write dynamic plugin request memory failed")
	}

	// Execute returns one packed pointer/length pair so the host can read the
	// response bytes without any JSON or text-based marshaling layer.
	executeResult, err := executeFn.Call(ctx, uint64(len(requestContent)))
	if err != nil {
		return nil, gerror.Wrap(err, "call dynamic plugin execute failed")
	}
	if len(executeResult) == 0 {
		return nil, gerror.New("dynamic plugin execute returned no valid response")
	}
	responsePointer, responseLength := decodeDynamicResponsePointer(executeResult[0])
	responseContent, ok := module.Memory().Read(responsePointer, responseLength)
	if !ok {
		return nil, gerror.New("read dynamic plugin response memory failed")
	}
	response, err = bridgecodec.DecodeResponseEnvelope(responseContent)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// bridgeExecutionContext returns a deadline-bound context for one dynamic
// plugin bridge execution. Existing caller deadlines remain authoritative.
func bridgeExecutionContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); ok {
		return ctx, nil
	}
	return context.WithTimeout(ctx, defaultBridgeExecutionTimeout)
}

// cloneExecutionManifestResources detaches release-bound manifest resources for
// one execution so host functions cannot mutate caller-owned maps.
func cloneExecutionManifestResources(source map[string][]byte) map[string][]byte {
	if len(source) == 0 {
		return nil
	}
	clone := make(map[string][]byte, len(source))
	for path, content := range source {
		clone[path] = append([]byte(nil), content...)
	}
	return clone
}

// decodeDynamicResponsePointer unpacks the bridge return value into pointer and length.
func decodeDynamicResponsePointer(value uint64) (uint32, uint32) {
	return uint32(value >> 32), uint32(value & 0xffffffff)
}

// This file executes dynamic plugin bridge requests through the Wasm
// alloc/write/execute/read ABI.

package wasm

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/tetratelabs/wazero"

	bridgecodec "lina-core/pkg/pluginbridge/codec"
	bridgecontract "lina-core/pkg/pluginbridge/contract"
)

// ExecuteBridge executes one bridge request against the archived active wasm
// artifact using the alloc/write/execute/read ABI sequence.
func ExecuteBridge(
	ctx context.Context,
	input ExecutionInput,
	requestContent []byte,
) (response *bridgecontract.BridgeResponseEnvelopeV1, err error) {
	if input.BridgeSpec == nil {
		return nil, gerror.New("dynamic plugin is missing Wasm bridge metadata")
	}

	rt, compiled, err := getOrCompileWasmModule(ctx, input.ArtifactPath)
	if err != nil {
		return nil, err
	}

	// Each request gets a fresh module instance because guest globals (request
	// and response buffers) are mutable and must not be shared across requests.
	module, err := rt.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName("").WithStartFunctions())
	if err != nil {
		return nil, gerror.Wrap(err, "instantiate dynamic plugin Wasm failed")
	}
	defer func() {
		if closeErr := module.Close(ctx); closeErr != nil && err == nil {
			err = gerror.Wrap(closeErr, "close dynamic plugin Wasm module failed")
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
		cronCollector:   input.CronCollector,
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

// decodeDynamicResponsePointer unpacks the bridge return value into pointer and length.
func decodeDynamicResponsePointer(value uint64) (uint32, uint32) {
	return uint32(value >> 32), uint32(value & 0xffffffff)
}

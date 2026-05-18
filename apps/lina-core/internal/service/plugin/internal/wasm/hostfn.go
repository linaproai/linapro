// This file registers the lina_env host module on the wazero runtime and
// implements the single host_call dispatch function for structured host services.

package wasm

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	bridgehostcall "lina-core/pkg/pluginbridge/hostcall"
)

// registerHostCallModule registers the lina_env host module with the host_call
// function on the given wazero runtime. This must be called after WASI
// instantiation and before module compilation, because the guest module imports
// from lina_env and wazero validates imports at compile time.
// registerHostCallModule registers the lina_env host module and host_call export.
func registerHostCallModule(ctx context.Context, rt wazero.Runtime) error {
	_, err := rt.NewHostModuleBuilder(bridgehostcall.HostModuleName).
		NewFunctionBuilder().
		WithGoModuleFunction(
			api.GoModuleFunc(hostCallHandler),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
			[]api.ValueType{api.ValueTypeI64},
		).
		Export(bridgehostcall.HostCallFunctionName).
		Instantiate(ctx)
	return err
}

// hostCallHandler is the wazero host function implementation for lina_env.host_call.
// It reads the opcode, request pointer, and request length from the stack,
// dispatches to the appropriate capability handler, writes the response into
// guest memory via the lina_host_call_alloc export, and returns the packed
// (pointer << 32 | length) result.
// hostCallHandler is the wazero callback that reads guest input, dispatches
// the host call, and writes the encoded response back into guest memory.
func hostCallHandler(ctx context.Context, mod api.Module, stack []uint64) {
	var (
		opcode = uint32(stack[0])
		reqPtr = uint32(stack[1])
		reqLen = uint32(stack[2])
	)

	// Extract per-request context.
	hcc := hostCallContextFrom(ctx)
	if hcc == nil {
		stack[0] = writeHostCallError(ctx, mod, bridgehostcall.HostCallStatusInternalError, "host call context not available")
		return
	}

	// Read request bytes from guest memory.
	var reqBytes []byte
	if reqLen > 0 {
		var ok bool
		reqBytes, ok = mod.Memory().Read(reqPtr, reqLen)
		if !ok {
			stack[0] = writeHostCallError(ctx, mod, bridgehostcall.HostCallStatusInternalError, "failed to read host call request from guest memory")
			return
		}
		// Make a copy since guest memory may be invalidated by re-entrant alloc.
		copied := make([]byte, len(reqBytes))
		copy(copied, reqBytes)
		reqBytes = copied
	}

	if opcode != bridgehostcall.OpcodeServiceInvoke {
		stack[0] = writeHostCallError(ctx, mod, bridgehostcall.HostCallStatusNotFound,
			fmt.Sprintf("unknown host call opcode: 0x%04x", opcode))
		return
	}

	// Dispatch to structured host service handler.
	respEnvelope := dispatchHostCall(ctx, hcc, opcode, reqBytes)

	// Encode and write response to guest memory.
	respBytes := bridgehostcall.MarshalHostCallResponse(respEnvelope)
	stack[0] = writeHostCallResponse(ctx, mod, respBytes)
}

// dispatchHostCall routes the opcode to the correct structured host service handler.
// dispatchHostCall routes the opcode to the correct structured host service handler.
func dispatchHostCall(ctx context.Context, hcc *hostCallContext, opcode uint32, reqBytes []byte) *bridgehostcall.HostCallResponseEnvelope {
	switch opcode {
	case bridgehostcall.OpcodeServiceInvoke:
		return handleHostServiceInvoke(ctx, hcc, reqBytes)
	default:
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusNotFound,
			fmt.Sprintf("unhandled host call opcode: 0x%04x", opcode))
	}
}

// writeHostCallResponse writes encoded response bytes into guest memory via
// the lina_host_call_alloc export and returns a packed (pointer << 32 | length).
// writeHostCallResponse writes encoded response bytes into guest memory and
// returns the packed pointer/length pair expected by the bridge ABI.
func writeHostCallResponse(ctx context.Context, mod api.Module, respBytes []byte) uint64 {
	if len(respBytes) == 0 {
		return 0
	}

	allocFn := mod.ExportedFunction(bridgehostcall.DefaultGuestHostCallAllocExport)
	if allocFn == nil {
		// Guest does not export the host call alloc function; cannot write response.
		return 0
	}

	result, err := allocFn.Call(ctx, uint64(len(respBytes)))
	if err != nil || len(result) == 0 {
		return 0
	}
	respPtr := uint32(result[0])
	if !mod.Memory().Write(respPtr, respBytes) {
		return 0
	}

	return uint64(respPtr)<<32 | uint64(len(respBytes))
}

// writeHostCallError is a convenience wrapper that encodes an error response
// and writes it to guest memory.
// writeHostCallError encodes one error envelope and writes it to guest memory.
func writeHostCallError(ctx context.Context, mod api.Module, status uint32, message string) uint64 {
	envelope := bridgehostcall.NewHostCallErrorResponse(status, message)
	return writeHostCallResponse(ctx, mod, bridgehostcall.MarshalHostCallResponse(envelope))
}

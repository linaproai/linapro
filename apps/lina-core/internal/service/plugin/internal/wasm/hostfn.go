// This file registers the lina_env host module on the wazero runtime and
// implements the single host_call dispatch function for structured host services.

package wasm

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"lina-core/pkg/pluginbridge"
)

// registerHostCallModule registers the lina_env host module with the host_call
// function on the given wazero runtime. This must be called after WASI
// instantiation and before module compilation, because the guest module imports
// from lina_env and wazero validates imports at compile time.
func registerHostCallModule(ctx context.Context, rt wazero.Runtime) error {
	_, err := rt.NewHostModuleBuilder(pluginbridge.HostModuleName).
		NewFunctionBuilder().
		WithGoModuleFunction(
			api.GoModuleFunc(hostCallHandler),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
			[]api.ValueType{api.ValueTypeI64},
		).
		Export(pluginbridge.HostCallFunctionName).
		Instantiate(ctx)
	return err
}

// hostCallHandler is the wazero host function implementation for lina_env.host_call.
// It reads the opcode, request pointer, and request length from the stack,
// dispatches to the appropriate capability handler, writes the response into
// guest memory via the lina_host_call_alloc export, and returns the packed
// (pointer << 32 | length) result.
func hostCallHandler(ctx context.Context, mod api.Module, stack []uint64) {
	var (
		opcode = uint32(stack[0])
		reqPtr = uint32(stack[1])
		reqLen = uint32(stack[2])
	)

	// Extract per-request context.
	hcc := hostCallContextFrom(ctx)
	if hcc == nil {
		stack[0] = writeHostCallError(ctx, mod, pluginbridge.HostCallStatusInternalError, "host call context not available")
		return
	}

	// Read request bytes from guest memory.
	var reqBytes []byte
	if reqLen > 0 {
		var ok bool
		reqBytes, ok = mod.Memory().Read(reqPtr, reqLen)
		if !ok {
			stack[0] = writeHostCallError(ctx, mod, pluginbridge.HostCallStatusInternalError, "failed to read host call request from guest memory")
			return
		}
		// Make a copy since guest memory may be invalidated by re-entrant alloc.
		copied := make([]byte, len(reqBytes))
		copy(copied, reqBytes)
		reqBytes = copied
	}

	if opcode != pluginbridge.OpcodeServiceInvoke {
		stack[0] = writeHostCallError(ctx, mod, pluginbridge.HostCallStatusNotFound,
			fmt.Sprintf("unknown host call opcode: 0x%04x", opcode))
		return
	}

	// Dispatch to structured host service handler.
	respEnvelope := dispatchHostCall(ctx, hcc, opcode, reqBytes)

	// Encode and write response to guest memory.
	respBytes := pluginbridge.MarshalHostCallResponse(respEnvelope)
	stack[0] = writeHostCallResponse(ctx, mod, respBytes)
}

// dispatchHostCall routes the opcode to the correct structured host service handler.
func dispatchHostCall(ctx context.Context, hcc *hostCallContext, opcode uint32, reqBytes []byte) *pluginbridge.HostCallResponseEnvelope {
	switch opcode {
	case pluginbridge.OpcodeServiceInvoke:
		return handleHostServiceInvoke(ctx, hcc, reqBytes)
	default:
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusNotFound,
			fmt.Sprintf("unhandled host call opcode: 0x%04x", opcode))
	}
}

// writeHostCallResponse writes encoded response bytes into guest memory via
// the lina_host_call_alloc export and returns a packed (pointer << 32 | length).
func writeHostCallResponse(ctx context.Context, mod api.Module, respBytes []byte) uint64 {
	if len(respBytes) == 0 {
		return 0
	}

	allocFn := mod.ExportedFunction(pluginbridge.DefaultGuestHostCallAllocExport)
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
func writeHostCallError(ctx context.Context, mod api.Module, status uint32, message string) uint64 {
	envelope := pluginbridge.NewHostCallErrorResponse(status, message)
	return writeHostCallResponse(ctx, mod, pluginbridge.MarshalHostCallResponse(envelope))
}

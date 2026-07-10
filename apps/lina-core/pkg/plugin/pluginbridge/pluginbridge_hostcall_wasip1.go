//go:build wasip1

// This file provides guest-side helpers for invoking structured host services
// through the lina_env.host_call import. It is only compiled for wasip1 targets.

package pluginbridge

import (
	"strconv"
	"unsafe"

	"lina-core/pkg/plugin/pluginbridge/protocol"

	"github.com/gogf/gf/v2/errors/gerror"
)

// linaHostCall is the imported host function provided by the lina_env module.
//
//go:wasmimport lina_env host_call
func linaHostCall(opcode uint32, reqPtr uint32, reqLen uint32) uint64

// invokeHostCall sends one host call request and returns the decoded payload.
func invokeHostCall(opcode uint32, reqBytes []byte) ([]byte, error) {
	var reqPtr uint32
	var reqLen uint32
	if len(reqBytes) > 0 {
		reqPtr = uint32(uintptr(unsafe.Pointer(&reqBytes[0])))
		reqLen = uint32(len(reqBytes))
	}

	var (
		packed  = linaHostCall(opcode, reqPtr, reqLen)
		respLen = uint32(packed & 0xffffffff)
	)

	if respLen == 0 {
		return nil, nil
	}

	buf := guestHostCallResponseBuffer
	if uint32(len(buf)) < respLen {
		return nil, gerror.Newf("host call response buffer underflow: have %d, need %d", len(buf), respLen)
	}
	envelope, err := protocol.UnmarshalHostCallResponse(buf[:respLen])
	if err != nil {
		return nil, gerror.Wrap(err, "host call response decode failed")
	}
	if envelope.Status != protocol.HostCallStatusSuccess {
		message := string(envelope.Payload)
		if message == "" {
			message = "host call failed with status " + strconv.FormatInt(int64(envelope.Status), 10)
		}
		return nil, gerror.Newf("host call error (status=%d): %s", envelope.Status, message)
	}
	return envelope.Payload, nil
}

// invokeHostServiceEnvelope dispatches one structured host-service request
// envelope through the shared host call import.
func invokeHostServiceEnvelope(request *protocol.HostServiceRequestEnvelope) ([]byte, error) {
	return invokeHostCall(protocol.OpcodeServiceInvoke, protocol.MarshalHostServiceRequestEnvelope(request))
}

// invokeHostService builds one core-owned structured host-service request
// envelope and dispatches it through the shared host call import.
func invokeHostService(service string, method string, resourceRef string, table string, payload []byte) ([]byte, error) {
	return invokeHostServiceEnvelope(&protocol.HostServiceRequestEnvelope{
		Service:     service,
		Method:      method,
		ResourceRef: resourceRef,
		Table:       table,
		Payload:     payload,
	})
}

// InvokeHostService dispatches one structured host-service request through the
// WASI host call transport.
func InvokeHostService(service string, method string, resourceRef string, table string, payload []byte) ([]byte, error) {
	return invokeHostService(service, method, resourceRef, table, payload)
}

// InvokeOwnerHostService dispatches one plugin-owned structured host-service
// request through the WASI host call transport.
func InvokeOwnerHostService(
	owner string,
	service string,
	version string,
	method string,
	resourceRef string,
	table string,
	payload []byte,
) ([]byte, error) {
	return invokeHostServiceEnvelope(&protocol.HostServiceRequestEnvelope{
		Owner:       owner,
		Service:     service,
		Version:     version,
		Method:      method,
		ResourceRef: resourceRef,
		Table:       table,
		Payload:     payload,
	})
}

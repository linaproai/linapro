//go:build wasip1

// This file provides guest-side helpers for invoking structured host services
// through the lina_env.host_call import. It is only compiled for wasip1 targets.

package pluginbridge

import (
	"strconv"
	"unsafe"

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
	envelope, err := UnmarshalHostCallResponse(buf[:respLen])
	if err != nil {
		return nil, gerror.Wrap(err, "host call response decode failed")
	}
	if envelope.Status != HostCallStatusSuccess {
		message := string(envelope.Payload)
		if message == "" {
			message = "host call failed with status " + strconv.FormatInt(int64(envelope.Status), 10)
		}
		return nil, gerror.Newf("host call error (status=%d): %s", envelope.Status, message)
	}
	return envelope.Payload, nil
}

// invokeHostService builds one structured host-service request envelope and
// dispatches it through the shared host call import.
func invokeHostService(service string, method string, resourceRef string, table string, payload []byte) ([]byte, error) {
	request := &HostServiceRequestEnvelope{
		Service:     service,
		Method:      method,
		ResourceRef: resourceRef,
		Table:       table,
		Payload:     payload,
	}
	return invokeHostCall(OpcodeServiceInvoke, MarshalHostServiceRequestEnvelope(request))
}

// HostDBQueryResult preserves the previous guest-side result shape for callers
// that have not yet migrated to the structured data service SDK.
type HostDBQueryResult struct {
	// Columns lists the result set column names in row order.
	Columns []string
	// Rows stores the tabular result values as string slices per row.
	Rows [][]string
	// RowCount reports the total number of rows returned by the legacy query.
	RowCount int
}

// HostDBQuery is no longer part of the public host service protocol.
func HostDBQuery(_ string, _ []string, _ int) (*HostDBQueryResult, error) {
	return nil, gerror.New("HostDBQuery has been removed; use the structured pluginbridge.Data() service instead")
}

// HostDBExecute is no longer part of the public host service protocol.
func HostDBExecute(_ string, _ []string) (int64, int64, error) {
	return 0, 0, gerror.New("HostDBExecute has been removed; use the structured pluginbridge.Data() service instead")
}

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

// RuntimeHostService exposes guest-side helpers for the runtime host service.
type RuntimeHostService interface {
	// Log writes one structured runtime log entry through the host.
	Log(level int, message string, fields map[string]string) error
	// StateGet reads one plugin-scoped runtime state value by key.
	StateGet(key string) (string, bool, error)
	// StateSet writes one plugin-scoped runtime state value.
	StateSet(key string, value string) error
	// StateDelete removes one plugin-scoped runtime state value.
	StateDelete(key string) error
	// StateGetInt reads one integer runtime state value.
	StateGetInt(key string) (int, bool, error)
	// StateSetInt writes one integer runtime state value.
	StateSetInt(key string, value int) error
	// Now returns the current host time string.
	Now() (string, error)
	// UUID returns one host-generated unique identifier string.
	UUID() (string, error)
	// Node returns the current host node identity string.
	Node() (string, error)
}

type runtimeHostService struct{}

var defaultRuntimeHostService RuntimeHostService = &runtimeHostService{}

// Runtime returns the runtime host service guest client.
func Runtime() RuntimeHostService {
	return defaultRuntimeHostService
}

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

// Log writes one structured runtime log entry through the host.
func (s *runtimeHostService) Log(level int, message string, fields map[string]string) error {
	request := &HostCallLogRequest{
		Level:   int32(level),
		Message: message,
		Fields:  fields,
	}
	_, err := invokeHostService(HostServiceRuntime, HostServiceMethodRuntimeLogWrite, "", "", MarshalHostCallLogRequest(request))
	return err
}

// StateGet reads one plugin-scoped runtime state value by key.
func (s *runtimeHostService) StateGet(key string) (string, bool, error) {
	request := &HostCallStateGetRequest{Key: key}
	payload, err := invokeHostService(HostServiceRuntime, HostServiceMethodRuntimeStateGet, "", "", MarshalHostCallStateGetRequest(request))
	if err != nil {
		return "", false, err
	}
	if len(payload) == 0 {
		return "", false, nil
	}
	response, err := UnmarshalHostCallStateGetResponse(payload)
	if err != nil {
		return "", false, err
	}
	return response.Value, response.Found, nil
}

// StateSet writes one plugin-scoped runtime state value.
func (s *runtimeHostService) StateSet(key string, value string) error {
	request := &HostCallStateSetRequest{Key: key, Value: value}
	_, err := invokeHostService(HostServiceRuntime, HostServiceMethodRuntimeStateSet, "", "", MarshalHostCallStateSetRequest(request))
	return err
}

// StateDelete removes one plugin-scoped runtime state value.
func (s *runtimeHostService) StateDelete(key string) error {
	request := &HostCallStateDeleteRequest{Key: key}
	_, err := invokeHostService(HostServiceRuntime, HostServiceMethodRuntimeStateDelete, "", "", MarshalHostCallStateDeleteRequest(request))
	return err
}

// StateGetInt reads one integer runtime state value.
func (s *runtimeHostService) StateGetInt(key string) (int, bool, error) {
	value, found, err := s.StateGet(key)
	if err != nil || !found {
		return 0, found, err
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return 0, true, gerror.Newf("state value for %q is not an integer: %s", key, value)
	}
	return number, true, nil
}

// StateSetInt writes one integer runtime state value.
func (s *runtimeHostService) StateSetInt(key string, value int) error {
	return s.StateSet(key, strconv.Itoa(value))
}

// Now returns the current host time string.
func (s *runtimeHostService) Now() (string, error) {
	return s.runtimeInfoValue(HostServiceMethodRuntimeInfoNow)
}

// UUID returns one host-generated unique identifier string.
func (s *runtimeHostService) UUID() (string, error) {
	return s.runtimeInfoValue(HostServiceMethodRuntimeInfoUUID)
}

// Node returns the current host node identity string.
func (s *runtimeHostService) Node() (string, error) {
	return s.runtimeInfoValue(HostServiceMethodRuntimeInfoNode)
}

func (s *runtimeHostService) runtimeInfoValue(method string) (string, error) {
	payload, err := invokeHostService(HostServiceRuntime, method, "", "", nil)
	if err != nil {
		return "", err
	}
	if len(payload) == 0 {
		return "", nil
	}
	response, err := UnmarshalHostServiceValueResponse(payload)
	if err != nil {
		return "", err
	}
	return response.Value, nil
}

// HostLog writes one runtime log entry through the host.
func HostLog(level int, message string, fields map[string]string) error {
	return Runtime().Log(level, message, fields)
}

// HostStateGet reads one plugin-scoped runtime state value.
func HostStateGet(key string) (string, bool, error) {
	return Runtime().StateGet(key)
}

// HostStateSet writes one plugin-scoped runtime state value.
func HostStateSet(key string, value string) error {
	return Runtime().StateSet(key, value)
}

// HostStateDelete removes one plugin-scoped runtime state value.
func HostStateDelete(key string) error {
	return Runtime().StateDelete(key)
}

// HostStateGetInt reads one integer plugin-scoped runtime state value.
func HostStateGetInt(key string) (int, bool, error) {
	return Runtime().StateGetInt(key)
}

// HostStateSetInt writes one integer plugin-scoped runtime state value.
func HostStateSetInt(key string, value int) error {
	return Runtime().StateSetInt(key, value)
}

// HostDBQueryResult preserves the previous guest-side result shape for callers
// that have not yet migrated to the structured data service SDK.
type HostDBQueryResult struct {
	Columns  []string
	Rows     [][]string
	RowCount int
}

// HostDBQuery is no longer part of the public host service protocol.
func HostDBQuery(_ string, _ []string, _ int) (*HostDBQueryResult, error) {
	return nil, gerror.New("HostDBQuery 已移除，请改用 pluginbridge.Data() 结构化数据服务")
}

// HostDBExecute is no longer part of the public host service protocol.
func HostDBExecute(_ string, _ []string) (int64, int64, error) {
	return 0, 0, gerror.New("HostDBExecute 已移除，请改用 pluginbridge.Data() 结构化数据服务")
}

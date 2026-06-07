//go:build wasip1

// This file provides guest-side helpers for the runtime host service so
// runtime, data, storage, and network SDKs share the same structured shape.

package guest

import (
	"strconv"

	"lina-core/pkg/plugin/pluginbridge/protocol"

	"github.com/gogf/gf/v2/errors/gerror"
)

// runtimeHostService is the default guest-side runtime host-service client.
type runtimeHostService struct{}

// defaultRuntimeHostService stores the singleton runtime host-service client
// used by package-level helpers.
var defaultRuntimeHostService RuntimeHostService = &runtimeHostService{}

// Runtime returns the runtime host service guest client.
func Runtime() RuntimeHostService {
	return defaultRuntimeHostService
}

// Log writes one structured runtime log entry through the host.
func (s *runtimeHostService) Log(level int, message string, fields map[string]string) error {
	request := &protocol.HostCallLogRequest{
		Level:   int32(level),
		Message: message,
		Fields:  fields,
	}
	_, err := invokeGuestHostService(
		protocol.HostServiceRuntime,
		protocol.HostServiceMethodRuntimeLogWrite,
		"",
		"",
		protocol.MarshalHostCallLogRequest(request),
	)
	return err
}

// StateGet reads one plugin-scoped runtime state value by key.
func (s *runtimeHostService) StateGet(key string) (string, bool, error) {
	request := &protocol.HostCallStateGetRequest{Key: key}
	payload, err := invokeGuestHostService(
		protocol.HostServiceRuntime,
		protocol.HostServiceMethodRuntimeStateGet,
		"",
		"",
		protocol.MarshalHostCallStateGetRequest(request),
	)
	if err != nil {
		return "", false, err
	}
	if len(payload) == 0 {
		return "", false, nil
	}
	response, err := protocol.UnmarshalHostCallStateGetResponse(payload)
	if err != nil {
		return "", false, err
	}
	return response.Value, response.Found, nil
}

// StateSet writes one plugin-scoped runtime state value.
func (s *runtimeHostService) StateSet(key string, value string) error {
	request := &protocol.HostCallStateSetRequest{Key: key, Value: value}
	_, err := invokeGuestHostService(
		protocol.HostServiceRuntime,
		protocol.HostServiceMethodRuntimeStateSet,
		"",
		"",
		protocol.MarshalHostCallStateSetRequest(request),
	)
	return err
}

// StateDelete removes one plugin-scoped runtime state value.
func (s *runtimeHostService) StateDelete(key string) error {
	request := &protocol.HostCallStateDeleteRequest{Key: key}
	_, err := invokeGuestHostService(
		protocol.HostServiceRuntime,
		protocol.HostServiceMethodRuntimeStateDelete,
		"",
		"",
		protocol.MarshalHostCallStateDeleteRequest(request),
	)
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
	return s.runtimeInfoValue(protocol.HostServiceMethodRuntimeInfoNow)
}

// UUID returns one host-generated unique identifier string.
func (s *runtimeHostService) UUID() (string, error) {
	return s.runtimeInfoValue(protocol.HostServiceMethodRuntimeInfoUUID)
}

// Node returns the current host node identity string.
func (s *runtimeHostService) Node() (string, error) {
	return s.runtimeInfoValue(protocol.HostServiceMethodRuntimeInfoNode)
}

// runtimeInfoValue reads one runtime info method response and extracts the
// string value payload.
func (s *runtimeHostService) runtimeInfoValue(method string) (string, error) {
	payload, err := invokeGuestHostService(protocol.HostServiceRuntime, method, "", "", nil)
	if err != nil {
		return "", err
	}
	if len(payload) == 0 {
		return "", nil
	}
	response, err := protocol.UnmarshalHostServiceValueResponse(payload)
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

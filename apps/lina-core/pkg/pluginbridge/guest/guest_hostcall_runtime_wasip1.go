//go:build wasip1

// This file provides guest-side helpers for the runtime host service so
// runtime, data, storage, and network SDKs share the same structured shape.

package guest

import (
	"strconv"

	"github.com/gogf/gf/v2/errors/gerror"
)

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
	request := &HostCallLogRequest{
		Level:   int32(level),
		Message: message,
		Fields:  fields,
	}
	_, err := invokeHostService(
		HostServiceRuntime,
		HostServiceMethodRuntimeLogWrite,
		"",
		"",
		MarshalHostCallLogRequest(request),
	)
	return err
}

// StateGet reads one plugin-scoped runtime state value by key.
func (s *runtimeHostService) StateGet(key string) (string, bool, error) {
	request := &HostCallStateGetRequest{Key: key}
	payload, err := invokeHostService(
		HostServiceRuntime,
		HostServiceMethodRuntimeStateGet,
		"",
		"",
		MarshalHostCallStateGetRequest(request),
	)
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
	_, err := invokeHostService(
		HostServiceRuntime,
		HostServiceMethodRuntimeStateSet,
		"",
		"",
		MarshalHostCallStateSetRequest(request),
	)
	return err
}

// StateDelete removes one plugin-scoped runtime state value.
func (s *runtimeHostService) StateDelete(key string) error {
	request := &HostCallStateDeleteRequest{Key: key}
	_, err := invokeHostService(
		HostServiceRuntime,
		HostServiceMethodRuntimeStateDelete,
		"",
		"",
		MarshalHostCallStateDeleteRequest(request),
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

// runtimeInfoValue reads one runtime info method response and extracts the
// string value payload.
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

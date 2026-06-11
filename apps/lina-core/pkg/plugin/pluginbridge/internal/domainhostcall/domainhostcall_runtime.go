// This file implements the guest-side runtime host-service client using the
// injected raw host-service invoker.

package domainhostcall

import (
	"strconv"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// runtimeService adapts runtime host-service calls to the pluginbridge runtime
// helper contract.
type runtimeService struct{ baseService }

// Runtime creates the runtime host service guest client.
func Runtime(invoker HostServiceInvoker) *runtimeService {
	return &runtimeService{baseService: newBaseServiceWithHostService(nil, invoker)}
}

// Log writes one structured runtime log entry through the host.
func (s *runtimeService) Log(level int, message string, fields map[string]string) error {
	request := &protocol.HostCallLogRequest{
		Level:   int32(level),
		Message: message,
		Fields:  fields,
	}
	_, err := s.callHostService(
		protocol.HostServiceRuntime,
		protocol.HostServiceMethodRuntimeLogWrite,
		"",
		"",
		protocol.MarshalHostCallLogRequest(request),
	)
	return err
}

// StateGet reads one plugin-scoped runtime state value by key.
func (s *runtimeService) StateGet(key string) (string, bool, error) {
	request := &protocol.HostCallStateGetRequest{Key: key}
	payload, err := s.callHostService(
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
func (s *runtimeService) StateSet(key string, value string) error {
	request := &protocol.HostCallStateSetRequest{Key: key, Value: value}
	_, err := s.callHostService(
		protocol.HostServiceRuntime,
		protocol.HostServiceMethodRuntimeStateSet,
		"",
		"",
		protocol.MarshalHostCallStateSetRequest(request),
	)
	return err
}

// StateDelete removes one plugin-scoped runtime state value.
func (s *runtimeService) StateDelete(key string) error {
	request := &protocol.HostCallStateDeleteRequest{Key: key}
	_, err := s.callHostService(
		protocol.HostServiceRuntime,
		protocol.HostServiceMethodRuntimeStateDelete,
		"",
		"",
		protocol.MarshalHostCallStateDeleteRequest(request),
	)
	return err
}

// StateGetInt reads one integer runtime state value.
func (s *runtimeService) StateGetInt(key string) (int, bool, error) {
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
func (s *runtimeService) StateSetInt(key string, value int) error {
	return s.StateSet(key, strconv.Itoa(value))
}

// Now returns the current host time string.
func (s *runtimeService) Now() (string, error) {
	return s.runtimeInfoValue(protocol.HostServiceMethodRuntimeInfoNow)
}

// UUID returns one host-generated unique identifier string.
func (s *runtimeService) UUID() (string, error) {
	return s.runtimeInfoValue(protocol.HostServiceMethodRuntimeInfoUUID)
}

// Node returns the current host node identity string.
func (s *runtimeService) Node() (string, error) {
	return s.runtimeInfoValue(protocol.HostServiceMethodRuntimeInfoNode)
}

func (s *runtimeService) runtimeInfoValue(method string) (string, error) {
	payload, err := s.callHostService(protocol.HostServiceRuntime, method, "", "", nil)
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

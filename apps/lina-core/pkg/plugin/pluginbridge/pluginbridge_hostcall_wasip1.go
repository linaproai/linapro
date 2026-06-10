//go:build wasip1

// This file provides guest-side helpers for invoking structured host services
// through the lina_env.host_call import and exposes generic transport, plugin
// config, and outbound network helpers used by higher level guest SDKs. It is
// only compiled for wasip1 targets.

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

// invokeHostService builds one structured host-service request envelope and
// dispatches it through the shared host call import.
func invokeHostService(service string, method string, resourceRef string, table string, payload []byte) ([]byte, error) {
	request := &protocol.HostServiceRequestEnvelope{
		Service:     service,
		Method:      method,
		ResourceRef: resourceRef,
		Table:       table,
		Payload:     payload,
	}
	return invokeHostCall(protocol.OpcodeServiceInvoke, protocol.MarshalHostServiceRequestEnvelope(request))
}

// InvokeHostService dispatches one structured host-service request through the
// WASI host call transport.
func InvokeHostService(service string, method string, resourceRef string, table string, payload []byte) ([]byte, error) {
	return invokeHostService(service, method, resourceRef, table, payload)
}

// configValue invokes one config host-service method and decodes the common
// response.
func configValue(key string) (string, bool, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsConfigGet,
		"",
		"",
		protocol.MarshalHostServiceConfigKeyRequest(&protocol.HostServiceConfigKeyRequest{Key: key}),
	)
	if err != nil {
		return "", false, err
	}
	if len(payload) == 0 {
		return "", false, nil
	}
	response, err := protocol.UnmarshalHostServiceConfigValueResponse(payload)
	if err != nil {
		return "", false, err
	}
	return response.Value, response.Found, nil
}

// networkHostService is the default guest-side outbound network host-service
// client.
type networkHostService struct{}

// defaultNetworkHostService stores the singleton outbound network host-service
// client used by package-level helpers.
var defaultNetworkHostService NetworkHostService = &networkHostService{}

// Network returns the outbound network host service guest client.
func Network() NetworkHostService {
	return defaultNetworkHostService
}

// Request executes one governed outbound HTTP request through the host.
func (s *networkHostService) Request(
	targetURL string,
	request *protocol.HostServiceNetworkRequest,
) (*protocol.HostServiceNetworkResponse, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceNetwork,
		protocol.HostServiceMethodNetworkRequest,
		targetURL,
		"",
		protocol.MarshalHostServiceNetworkRequest(request),
	)
	if err != nil {
		return nil, err
	}
	return protocol.UnmarshalHostServiceNetworkResponse(payload)
}

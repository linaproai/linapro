// This file defines the guest-side domain hostcall client constructors and
// shared request helpers. The package is intentionally internal so dynamic
// plugin authors continue to depend only on pkg/plugin/pluginbridge.

package domainhostcall

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

const hostCallsUnavailableMessage = "pluginbridge guest host-call transport is only available for wasip1 builds"

// errHostCallsUnavailable reports that a domain client method is not published
// through the dynamic-plugin host-service transport.
var errHostCallsUnavailable error = unavailableError{}

// unavailableError mirrors the public pluginbridge unavailable sentinel without
// importing the parent package from this internal implementation package.
type unavailableError struct{}

func (unavailableError) Error() string {
	return hostCallsUnavailableMessage
}

func (unavailableError) Is(target error) bool {
	return target != nil && target.Error() == hostCallsUnavailableMessage
}

// Invoker dispatches one already-encoded capability host-service request and
// decodes the host response into out when supplied.
type Invoker func(service string, method string, request []byte, out any) error

// ResourceInvoker dispatches one resource-scoped capability host-service
// request and decodes the host response into out when supplied.
type ResourceInvoker func(service string, method string, resourceRef string, request []byte, out any) error

// HostServiceInvoker dispatches one raw host-service request and returns the
// encoded response payload.
type HostServiceInvoker func(service string, method string, resourceRef string, table string, request []byte) ([]byte, error)

// baseService stores the hostcall invoker shared by one concrete client.
type baseService struct {
	invoke         Invoker
	invokeResource ResourceInvoker
	invokeHost     HostServiceInvoker
}

// newBaseService wraps the injected hostcall invoker for concrete clients.
func newBaseService(invoker Invoker) baseService {
	return baseService{invoke: invoker}
}

// newBaseServiceWithResource wraps the injected resource-aware hostcall
// invoker for concrete clients that require method-level resource references.
func newBaseServiceWithResource(invoker ResourceInvoker) baseService {
	return baseService{invokeResource: invoker}
}

// newBaseServiceWithHostService wraps both JSON and raw host-service invokers.
func newBaseServiceWithHostService(invoker Invoker, hostInvoker HostServiceInvoker) baseService {
	return baseService{invoke: invoker, invokeHost: hostInvoker}
}

// call dispatches one encoded request through the injected invoker.
func (s baseService) call(service string, method string, request []byte, out any) error {
	if s.invoke == nil {
		if s.invokeResource != nil {
			return s.invokeResource(service, method, "", request, out)
		}
		return gerror.New("domain hostcall invoker is nil")
	}
	return s.invoke(service, method, request, out)
}

// callResource dispatches one resource-scoped encoded request.
func (s baseService) callResource(service string, method string, resourceRef string, request []byte, out any) error {
	if s.invokeResource == nil {
		if s.invoke != nil && resourceRef == "" {
			return s.invoke(service, method, request, out)
		}
		return gerror.New("domain hostcall resource invoker is nil")
	}
	return s.invokeResource(service, method, resourceRef, request, out)
}

// callHostService dispatches one raw host-service request.
func (s baseService) callHostService(service string, method string, resourceRef string, table string, request []byte) ([]byte, error) {
	if s.invokeHost == nil {
		return nil, gerror.New("domain raw hostcall invoker is nil")
	}
	return s.invokeHost(service, method, resourceRef, table, request)
}

// callHostServiceJSONRequest encodes a JSON input envelope for raw host-service
// methods and decodes a JSON response envelope when out is supplied.
func (s baseService) callHostServiceJSONRequest(service string, method string, resourceRef string, table string, input any, out any) error {
	var payload []byte
	if input != nil {
		content, err := json.Marshal(input)
		if err != nil {
			return err
		}
		payload = protocol.MarshalHostServiceCapabilityJSONRequest(&protocol.HostServiceCapabilityJSONRequest{Value: content})
	}
	responsePayload, err := s.callHostService(service, method, resourceRef, table, payload)
	if err != nil || out == nil {
		return err
	}
	response, err := protocol.UnmarshalHostServiceCapabilityJSONResponse(responsePayload)
	if err != nil {
		return err
	}
	if response == nil || len(response.Value) == 0 {
		return nil
	}
	return json.Unmarshal(response.Value, out)
}

// callJSONRequest encodes one JSON input envelope and dispatches it.
func (s baseService) callJSONRequest(service string, method string, input any, out any) error {
	var payload []byte
	if input != nil {
		content, err := json.Marshal(input)
		if err != nil {
			return err
		}
		payload = protocol.MarshalHostServiceCapabilityJSONRequest(&protocol.HostServiceCapabilityJSONRequest{Value: content})
	}
	return s.call(service, method, payload, out)
}

// idsRequest carries a string identifier batch for JSON capability methods.
type idsRequest struct {
	IDs []string `json:"ids"`
}

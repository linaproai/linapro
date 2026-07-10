// This file provides shared guest host-call helpers used by capability clients,
// domain adapters, and ordinary Go tests. Build-specific raw transports remain
// in the wasip1 and non-WASI files.

package pluginbridge

import (
	"encoding/json"

	"lina-core/pkg/plugin/pluginbridge/protocol"

	"github.com/gogf/gf/v2/errors/gerror"
)

// ErrHostCallsUnavailable reports that guest host calls are unavailable in
// non-WASI builds.
var ErrHostCallsUnavailable = gerror.New(
	"pluginbridge guest host-call transport is only available for wasip1 builds",
)

// invokeGuestHostService dispatches one structured host-service request through
// the raw pluginbridge guest transport.
func invokeGuestHostService(service string, method string, resourceRef string, table string, payload []byte) ([]byte, error) {
	return InvokeHostService(service, method, resourceRef, table, payload)
}

// invokeCapabilityJSON invokes one capability host-service method and decodes
// the JSON response value into out when supplied.
func invokeCapabilityJSON(service string, method string, request []byte, out any) error {
	return invokeCapabilityJSONWithResource(service, method, "", request, out)
}

// invokeCapabilityJSONWithResource invokes one resource-scoped capability host
// service method and decodes the JSON response value into out when supplied.
func invokeCapabilityJSONWithResource(service string, method string, resourceRef string, request []byte, out any) error {
	payload, err := invokeGuestHostService(service, method, resourceRef, "", request)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	response, err := protocol.UnmarshalHostServiceJSONResponse(payload)
	if err != nil {
		return err
	}
	if response == nil || len(response.Value) == 0 {
		return gerror.New("capability response is empty")
	}
	if err = json.Unmarshal(response.Value, out); err != nil {
		return gerror.Wrap(err, "decode capability response failed")
	}
	return nil
}

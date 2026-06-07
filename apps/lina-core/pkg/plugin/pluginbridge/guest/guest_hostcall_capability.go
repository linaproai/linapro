// This file provides shared JSON response handling for guest-side framework
// capability host calls. Domain-specific Org and Tenant clients live in their
// own files; this helper keeps the common pluginbridge envelope decoding in one
// place without changing the wasip1 transport split.

package guest

import (
	"encoding/json"

	"lina-core/pkg/plugin/pluginbridge/protocol"

	"github.com/gogf/gf/v2/errors/gerror"
)

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
	response, err := protocol.UnmarshalHostServiceCapabilityJSONResponse(payload)
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

// This file implements the backend summary business logic for the dynamic
// sample plugin.

package dynamicservice

import "lina-core/pkg/pluginbridge"

const backendSummaryMessage = "This backend example is executed through the plugin-demo-dynamic Wasm bridge runtime."

// BuildBackendSummaryPayload builds the backend summary response payload.
func (s *serviceImpl) BuildBackendSummaryPayload(request *pluginbridge.BridgeRequestEnvelopeV1) *backendSummaryPayload {
	payload := &backendSummaryPayload{
		Message:       backendSummaryMessage,
		PluginID:      request.PluginID,
		PublicPath:    request.Route.PublicPath,
		Access:        request.Route.Access,
		Permission:    request.Route.Permission,
		Authenticated: request.Identity != nil && request.Identity.UserID > 0,
	}
	if request.Identity != nil {
		payload.Username = stringPointer(request.Identity.Username)
		payload.IsSuperAdmin = boolPointer(request.Identity.IsSuperAdmin)
	}
	return payload
}

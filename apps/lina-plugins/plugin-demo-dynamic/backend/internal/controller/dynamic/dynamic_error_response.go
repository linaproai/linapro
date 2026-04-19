// This file translates dynamic demo business errors into bridge responses.

package dynamic

import (
	"lina-core/pkg/pluginbridge"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// buildDynamicErrorResponse maps sample business errors to normalized bridge
// error responses.
func buildDynamicErrorResponse(err error) *pluginbridge.BridgeResponseEnvelopeV1 {
	if err == nil {
		return pluginbridge.NewInternalErrorResponse("Dynamic plugin execution failed")
	}
	if dynamicservice.IsDemoRecordInvalidInput(err) {
		return pluginbridge.NewBadRequestResponse(err.Error())
	}
	if dynamicservice.IsDemoRecordNotFound(err) {
		return pluginbridge.NewNotFoundResponse(err.Error())
	}
	return pluginbridge.NewInternalErrorResponse(err.Error())
}

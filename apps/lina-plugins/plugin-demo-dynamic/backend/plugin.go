package backend

import (
	"lina-core/pkg/pluginbridge"
	"lina-plugin-demo-dynamic/backend/internal/controller/dynamic"
)

// guestRouteDispatcher is the reflection-based bridge dispatcher for the demo plugin backend.
var guestRouteDispatcher = pluginbridge.MustNewGuestControllerRouteDispatcher(dynamic.New())

// HandleRequest dispatches bridge requests to the matching dynamic controller
// method using the build-time RequestType contract.
func HandleRequest(
	request *pluginbridge.BridgeRequestEnvelopeV1,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return guestRouteDispatcher.HandleRequest(request)
}

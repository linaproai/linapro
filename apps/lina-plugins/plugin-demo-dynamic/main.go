package main

import (
	"lina-core/pkg/pluginbridge"
	dynamicbackend "lina-plugin-demo-dynamic/backend"
)

var guestRuntime = pluginbridge.NewGuestRuntime(dynamicbackend.HandleRequest)

//go:wasmexport lina_dynamic_route_alloc
func linaDynamicRouteAlloc(size uint32) uint32 {
	return guestRuntime.Alloc(size)
}

//go:wasmexport lina_dynamic_route_execute
func linaDynamicRouteExecute(size uint32) uint64 {
	responsePointer, responseLength, err := guestRuntime.Execute(size)
	if err != nil {
		fallback, _ := pluginbridge.EncodeResponseEnvelope(pluginbridge.NewInternalErrorResponse(err.Error()))
		responsePointer, responseLength, _ = guestRuntime.ExposeResponseBuffer(fallback)
	}
	return uint64(responsePointer)<<32 | uint64(responseLength)
}

//go:wasmexport lina_host_call_alloc
func linaHostCallAlloc(size uint32) uint32 {
	return guestRuntime.HostCallAlloc(size)
}

func main() {}

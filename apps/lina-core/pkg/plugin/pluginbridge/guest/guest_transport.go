// This file binds guest capability clients to the raw pluginbridge host-service
// transport without re-exporting bridge protocol DTOs.

package guest

// invokeGuestHostService dispatches one structured host-service request through
// the raw pluginbridge guest transport.
func invokeGuestHostService(service string, method string, resourceRef string, table string, payload []byte) ([]byte, error) {
	return InvokeHostService(service, method, resourceRef, table, payload)
}

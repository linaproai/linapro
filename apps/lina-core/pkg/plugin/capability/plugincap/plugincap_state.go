// This file defines source-plugin enablement lookups without publishing
// host-internal plugin service packages to plugin implementations.
package plugincap

// stateServiceAdapter bridges host plugin state into the published plugin contract.
type stateServiceAdapter struct {
	service StateService
}

// NewState creates and returns a plugin state service backed by the given service.
func NewState(service StateService) StateService {
	return &stateServiceAdapter{service: service}
}

// This file adapts text AI provider declarations to the shared capability
// registry.

package aitext

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	internalregistry "lina-core/pkg/plugin/capability/internal/capabilityregistry"
)

// ProviderStatuses returns all text AI provider states.
func (m *Manager) ProviderStatuses(ctx context.Context, runtime ProviderRuntime) []capmodel.ProviderStatus {
	if m == nil || m.registry == nil {
		return nil
	}
	statuses := m.registry.Statuses(ctx, runtime)
	result := make([]capmodel.ProviderStatus, 0, len(statuses))
	for _, status := range statuses {
		result = append(result, convertProviderStatus(status))
	}
	return result
}

// convertCapabilityStatus copies internal capability state into public DTOs.
func convertCapabilityStatus(status internalregistry.CapabilityStatus) capmodel.CapabilityStatus {
	providers := make([]capmodel.ProviderStatus, 0, len(status.Providers))
	for _, provider := range status.Providers {
		providers = append(providers, convertProviderStatus(provider))
	}
	return capmodel.CapabilityStatus{
		CapabilityID:   status.CapabilityID,
		Available:      status.Available,
		ActiveProvider: status.ActiveProvider,
		Reason:         status.Reason,
		Providers:      providers,
	}
}

// convertProviderStatus copies one internal provider state into a public DTO.
func convertProviderStatus(status internalregistry.ProviderStatus) capmodel.ProviderStatus {
	return capmodel.ProviderStatus{
		CapabilityID: status.CapabilityID,
		PluginID:     status.PluginID,
		Active:       status.Active,
		Conflict:     status.Conflict,
		Reason:       status.Reason,
	}
}

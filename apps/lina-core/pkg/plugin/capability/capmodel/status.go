// This file defines shared plugin capability declaration DTOs used by
// independent framework capability components.

package capmodel

// ProviderStatus describes one capability provider declaration state.
type ProviderStatus struct {
	// CapabilityID is the framework capability identifier.
	CapabilityID string
	// PluginID is the provider plugin identifier.
	PluginID string
	// Active reports whether this provider currently serves capability calls.
	Active bool
	// Conflict reports whether this provider is blocked by a singleton conflict.
	Conflict bool
	// Reason contains a stable diagnostic reason for inactive or conflicted state.
	Reason string
}

// CapabilityStatus describes the declared and currently usable provider state
// for one capability.
type CapabilityStatus struct {
	// CapabilityID is the framework capability identifier.
	CapabilityID string
	// Available reports whether the capability currently has an active provider.
	Available bool
	// ActiveProvider is the active provider plugin identifier, when available.
	ActiveProvider string
	// Reason contains a stable diagnostic reason when the capability is unavailable.
	Reason string
	// Providers contains all provider plugin states known for this capability.
	Providers []ProviderStatus
}

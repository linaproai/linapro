// This file defines shared plugin-domain capability primitives used by typed
// host capability packages. These types are intentionally storage-agnostic so
// plugins cannot infer host table names, auto-increment keys, or cache internals.

package capmodel

// DomainID is the storage-independent string encoding used by dynamic plugin
// protocols. Each domain should expose its own named ID type and convert through
// this representation at bridge boundaries.
type DomainID string

// BatchResult is the standard domain batch-read result. MissingIDs must not
// distinguish absent records from records hidden by tenant or data permissions.
type BatchResult[T any, ID comparable] struct {
	// Items contains visible records keyed by their requested domain ID.
	Items map[ID]T
	// MissingIDs contains requested IDs that were absent, invisible, or denied.
	MissingIDs []ID
}

// PageRequest constrains high-volume capability queries.
type PageRequest struct {
	// PageNum is one-based and defaults to 1 when omitted by an adapter.
	PageNum int
	// PageSize is the bounded page size requested by the caller.
	PageSize int
	// Limit is an optional hard item limit for non-page candidates.
	Limit int
	// Sort is an optional stable field sort key owned by the domain.
	Sort string
}

// PageResult is the standard bounded page response for capability methods.
type PageResult[T any] struct {
	// Items is the current page of visible records.
	Items []T
	// Total is the visible total within the caller scope.
	Total int
}

// LocalizedLabel carries the stable label key plus an optional localized label.
type LocalizedLabel struct {
	// Value is the stable domain value.
	Value string
	// LabelKey is the stable runtime i18n key for the value.
	LabelKey string
	// Label is the optional label resolved in the current request locale.
	Label string
}

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

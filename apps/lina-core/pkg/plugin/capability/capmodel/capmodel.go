// This file defines shared plugin-domain capability primitives used by typed
// host capability packages. These types are intentionally storage-agnostic so
// plugins cannot infer host table names, auto-increment keys, or cache internals.

package capmodel

import "time"

// DomainID is the storage-independent string encoding used by dynamic plugin
// protocols. Each domain should expose its own named ID type and convert through
// this representation at bridge boundaries.
type DomainID string

// ActorType identifies the actor category that initiated one capability call.
type ActorType string

const (
	// ActorTypeUser identifies an authenticated user actor.
	ActorTypeUser ActorType = "user"
	// ActorTypeSystem identifies a host-created system actor.
	ActorTypeSystem ActorType = "system"
)

// CapabilityActor carries the plugin-visible actor projection for auditing and
// authorization. UserID is storage-owned and should only be populated by host
// adapters that already know the current user identity.
type CapabilityActor struct {
	// Type is the actor category.
	Type ActorType
	// UserID is the authenticated host user identifier when Type is user.
	UserID int64
	// Name is the stable actor display or diagnostic name.
	Name string
	// SystemReason is required when Type is system and explains the host-owned
	// lifecycle, cron, hook, or provider action being performed.
	SystemReason string
}

// CapabilitySource identifies the plugin execution source.
type CapabilitySource string

const (
	// CapabilitySourceHTTP identifies a plugin HTTP request path.
	CapabilitySourceHTTP CapabilitySource = "http"
	// CapabilitySourceLifecycle identifies a plugin lifecycle callback.
	CapabilitySourceLifecycle CapabilitySource = "lifecycle"
	// CapabilitySourceHook identifies a plugin hook callback.
	CapabilitySourceHook CapabilitySource = "hook"
	// CapabilitySourceJobs identifies a plugin scheduled job.
	CapabilitySourceJobs CapabilitySource = "jobs"
	// CapabilitySourceProvider identifies a plugin provider callback.
	CapabilitySourceProvider CapabilitySource = "provider"
	// CapabilitySourceHost identifies host-originated framework orchestration.
	CapabilitySourceHost CapabilitySource = "host"
)

// CapabilityAuthorizationSnapshot is the plugin-visible slice of the
// install/enable authorization decision used by dynamic host service calls.
type CapabilityAuthorizationSnapshot struct {
	// Services maps service name to the authorized method set.
	Services map[string][]string
	// Resources maps service.method to authorized resource or projection keys.
	Resources map[string][]string
	// Permissions contains the caller's permission keys from the current identity snapshot.
	Permissions []string
	// Revision is the authorization snapshot revision used for diagnostics.
	Revision string
}

// CapabilityContext carries the required domain-call metadata. It is separate
// from context.Context because context.Context only transports request lifetime,
// cancellation, and logging scope; this value is part of the audited domain
// contract.
type CapabilityContext struct {
	// PluginID is the stable caller plugin identifier.
	PluginID string
	// Actor is the host-created actor projection for this call.
	Actor CapabilityActor
	// TenantID is the active tenant identifier encoded as a domain ID.
	TenantID DomainID
	// Source identifies the plugin execution source.
	Source CapabilitySource
	// SystemCall reports whether this is a host-created system actor call.
	SystemCall bool
	// Authorization is the dynamic-plugin authorization snapshot, when relevant.
	Authorization CapabilityAuthorizationSnapshot
	// Resource is the target resource or projection key for auditing.
	Resource string
	// AuditReason records a stable reason for sensitive management calls.
	AuditReason string
	// TraceID is the request or task trace identifier.
	TraceID string
	// RequestedAt is the host timestamp when the capability call was constructed.
	RequestedAt time.Time
}

// BatchResult is the standard domain batch-read result. MissingIDs must not
// distinguish absent records from records hidden by tenant or data permissions.
type BatchResult[T any, ID comparable] struct {
	// Items contains visible projections keyed by their requested domain ID.
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
	// Items is the current page of visible projections.
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

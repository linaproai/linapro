// Package plugincap defines plugin-governance capability contracts for plugins
// without exposing sys_plugin, plugin state tables, or runtime snapshots.
package plugincap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

// PluginID identifies one plugin resource.
type PluginID string

// Projection describes one plugin state visible through the domain contract.
type Projection struct {
	// ID is the stable plugin identifier.
	ID PluginID
	// Version is the effective version.
	Version string
	// Installed reports whether the plugin is installed.
	Installed bool
	// Enabled reports whether the plugin is enabled in the current scope.
	Enabled bool
	// Status is the domain-owned lifecycle status.
	Status string
}

// TenantProjection describes one tenant-controllable plugin projection.
type TenantProjection struct {
	// ID is the stable plugin identifier.
	ID PluginID
	// Name is the plugin display name.
	Name string
	// Version is the effective version.
	Version string
	// Type is the plugin top-level type.
	Type string
	// Description is the plugin description or remark.
	Description string
	// Installed reports whether the plugin is installed.
	Installed bool
	// Enabled reports whether the plugin is globally enabled.
	Enabled bool
	// ScopeNature is the plugin scope nature.
	ScopeNature string
	// InstallMode is the plugin install mode.
	InstallMode string
	// TenantEnabled reports the plugin enablement in the requested tenant.
	TenantEnabled bool
}

// SearchInput describes bounded plugin-governance projection search.
type SearchInput struct {
	// Keyword matches plugin ID, name or description.
	Keyword string
	// PluginID filters by plugin ID fragment.
	PluginID string
	// Name filters by plugin display name fragment.
	Name string
	// Type filters by top-level plugin type.
	Type string
	// Enabled optionally filters by global enablement status.
	Enabled *bool
	// Page constrains page number, page size and optional limit.
	Page capmodel.PageRequest
}

// TenantListInput describes bounded tenant plugin projection search.
type TenantListInput struct {
	// Keyword matches plugin ID, name or description.
	Keyword string
	// Type filters by top-level plugin type.
	Type string
	// TenantEnabled optionally filters by tenant enablement.
	TenantEnabled *bool
	// Page constrains page number, page size and optional limit.
	Page capmodel.PageRequest
}

// CapabilityKey identifies a framework capability status projection.
type CapabilityKey string

const (
	// CapabilityKeyOrg identifies the organization framework capability.
	CapabilityKeyOrg CapabilityKey = "org"
	// CapabilityKeyTenant identifies the tenant framework capability.
	CapabilityKeyTenant CapabilityKey = "tenant"
	// CapabilityKeyAIText identifies the text AI framework capability.
	CapabilityKeyAIText CapabilityKey = "ai.text"
)

const (
	// MaxPluginSearchPageSize is the maximum plugin search page size.
	MaxPluginSearchPageSize = 200
	// MaxCapabilityStatusBatchSize is the maximum framework capability status identifiers per request.
	MaxCapabilityStatusBatchSize = 50
)

// RegistryService defines read-oriented plugin governance capability methods.
type RegistryService interface {
	// Current returns the projection for the current caller plugin.
	Current(ctx context.Context, capCtx capmodel.CapabilityContext) (*Projection, error)
	// BatchGet returns visible plugin projections and opaque missing IDs.
	BatchGet(ctx context.Context, capCtx capmodel.CapabilityContext, ids []PluginID) (*capmodel.BatchResult[*Projection, PluginID], error)
	// Search returns bounded plugin governance projections.
	Search(ctx context.Context, capCtx capmodel.CapabilityContext, input SearchInput) (*capmodel.PageResult[*Projection], error)
	// ListTenantPlugins returns tenant-controllable plugins with tenant enablement.
	ListTenantPlugins(ctx context.Context, capCtx capmodel.CapabilityContext, input TenantListInput) (*capmodel.PageResult[*TenantProjection], error)
	// BatchGetCapabilityStatus returns framework capability status projections by stable key.
	BatchGetCapabilityStatus(ctx context.Context, capCtx capmodel.CapabilityContext, keys []CapabilityKey) (*capmodel.BatchResult[*capmodel.CapabilityStatus, CapabilityKey], error)
}

// Service defines the plugin-domain namespace exposed from capability.Services.
type Service interface {
	RegistryService
	// Config returns the current plugin's static configuration reader.
	Config() ConfigService
	// State returns plugin state and provider enablement lookups.
	State() StateService
	// Lifecycle returns plugin lifecycle orchestration operations.
	Lifecycle() LifecycleService
	// Registry returns the plugin governance projection service.
	Registry() RegistryService
}

// AdminService defines governed plugin lifecycle management commands.
type AdminService interface {
	// SetEnabled changes plugin enablement after tenant and lifecycle checks.
	SetEnabled(ctx context.Context, capCtx capmodel.CapabilityContext, id PluginID, enabled bool) error
	// ProvisionTenantDefaults creates missing default tenant plugin states.
	ProvisionTenantDefaults(ctx context.Context, capCtx capmodel.CapabilityContext, tenantID capmodel.DomainID) error
}

// ScopeService defines host-internal plugin governance helpers.
type ScopeService interface {
	// EnsureVisible rejects when any plugin is outside caller scope.
	EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []PluginID) error
}

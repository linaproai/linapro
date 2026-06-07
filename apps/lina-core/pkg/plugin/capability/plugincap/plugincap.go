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

// RegistryService defines read-oriented plugin governance capability methods.
type RegistryService interface {
	// BatchGetPlugins returns visible plugin projections and opaque missing IDs.
	BatchGetPlugins(ctx context.Context, capCtx capmodel.CapabilityContext, ids []PluginID) (*capmodel.BatchResult[*Projection, PluginID], error)
	// ListTenantPlugins returns tenant-controllable plugins with tenant enablement.
	ListTenantPlugins(ctx context.Context, capCtx capmodel.CapabilityContext) (*capmodel.PageResult[*TenantProjection], error)
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
	// SetPluginEnabled changes plugin enablement after tenant and lifecycle checks.
	SetPluginEnabled(ctx context.Context, capCtx capmodel.CapabilityContext, id PluginID, enabled bool) error
	// ProvisionTenantDefaults creates missing default tenant plugin states.
	ProvisionTenantDefaults(ctx context.Context, capCtx capmodel.CapabilityContext, tenantID capmodel.DomainID) error
}

// ScopeService defines host-internal plugin governance helpers.
type ScopeService interface {
	// EnsurePluginsVisible rejects when any plugin is outside caller scope.
	EnsurePluginsVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []PluginID) error
}

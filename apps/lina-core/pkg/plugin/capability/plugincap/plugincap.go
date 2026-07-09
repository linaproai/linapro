// Package plugincap defines plugin-governance capability contracts for plugins
// without exposing sys_plugin, plugin state tables, or runtime snapshots.
package plugincap

import (
	"context"
	"time"

	pluginv1 "lina-core/api/plugin/v1"
	"lina-core/pkg/plugin/capability/capmodel"

	"github.com/gogf/gf/v2/container/gvar"
)

// Service defines the plugin-domain namespace exposed from capability.Services.
// The plugin domain owns registry info records, plugin enablement lookups,
// plugin-owned configuration values, and governed lifecycle seams that are
// registered for dynamic plugins only when explicitly published by the host
// service registry.
type Service interface {
	// Config returns the current plugin's static configuration reader.
	Config() ConfigService
	// Registry returns the plugin governance info service.
	Registry() RegistryService
	// State returns plugin enablement state lookups.
	State() StateService
	// Lifecycle returns governed plugin lifecycle orchestration operations.
	Lifecycle() LifecycleService
}

// ConfigService defines the configuration operations published to source plugins.
type ConfigService interface {
	// Get returns the raw configuration value for the given key. When the key
	// is absent, Get returns defaultValue wrapped as *gvar.Var. Passing nil
	// preserves absent-key nil return semantics.
	Get(ctx context.Context, key string, defaultValue any) (*gvar.Var, error)
	// Exists reports whether the given configuration key exists.
	Exists(ctx context.Context, key string) (bool, error)
	// Scan scans the configuration section into target.
	Scan(ctx context.Context, key string, target any) error
	// String reads a string value or returns defaultValue when the key is absent or blank.
	String(ctx context.Context, key string, defaultValue string) (string, error)
	// Bool reads a bool value or returns defaultValue when the key is absent.
	Bool(ctx context.Context, key string, defaultValue bool) (bool, error)
	// Int reads an int value or returns defaultValue when the key is absent.
	Int(ctx context.Context, key string, defaultValue int) (int, error)
	// Duration reads a time.Duration value or returns defaultValue when the key is absent or blank.
	Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error)
}

// RegistryService defines read-oriented plugin registry capability methods.
type RegistryService interface {
	// Current returns the info for the current caller plugin.
	Current(ctx context.Context) (*PluginInfo, error)
	// Get returns one visible plugin info record.
	Get(ctx context.Context, id PluginID) (*PluginInfo, error)
	// BatchGet returns visible plugin info records and opaque missing IDs.
	BatchGet(ctx context.Context, ids []PluginID) (*capmodel.BatchResult[*PluginInfo, PluginID], error)
	// List returns bounded plugin governance info records.
	List(ctx context.Context, input ListInput) (*capmodel.PageResult[*PluginInfo], error)
	// ListTenantPlugins returns tenant-controllable plugin info records with tenant enablement.
	ListTenantPlugins(ctx context.Context, input TenantListInput) (*capmodel.PageResult[*TenantPluginInfo], error)
}

// StateService defines plugin enablement state lookup methods.
type StateService interface {
	// IsEnabled reports whether one plugin is enabled in the current scope.
	IsEnabled(ctx context.Context, pluginID PluginID) (bool, error)
	// IsProviderEnabled reports whether one plugin may serve provider calls.
	IsProviderEnabled(ctx context.Context, pluginID PluginID) (bool, error)
	// IsEnabledAuthoritative reports persisted plugin enablement bypassing local snapshots.
	IsEnabledAuthoritative(ctx context.Context, pluginID PluginID) (bool, error)
}

// LifecycleService exposes host-owned lifecycle orchestration to plugins that
// own tenant or plugin-governance modules. Dynamic plugins can call only the
// methods registered in the host-service registry and authorized through
// hostServices.
type LifecycleService interface {
	// EnsureTenantPluginDisableAllowed runs plugin lifecycle preconditions before
	// one tenant loses access to a plugin.
	EnsureTenantPluginDisableAllowed(ctx context.Context, pluginID string, tenantID int) error
	// NotifyTenantPluginDisabled runs best-effort lifecycle notifications after
	// one tenant loses access to a plugin.
	NotifyTenantPluginDisabled(ctx context.Context, pluginID string, tenantID int)
	// EnsureTenantDeleteAllowed runs plugin lifecycle preconditions before a
	// tenant is deleted.
	EnsureTenantDeleteAllowed(ctx context.Context, tenantID int) error
	// NotifyTenantDeleted runs best-effort lifecycle notifications after a
	// tenant has been deleted.
	NotifyTenantDeleted(ctx context.Context, tenantID int)
}

// PluginID identifies one plugin resource.
type PluginID string

// PluginInfo describes one plugin state visible through the domain contract.
type PluginInfo struct {
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

// TenantPluginInfo describes one tenant-controllable plugin info record.
type TenantPluginInfo struct {
	// ID is the stable plugin identifier.
	ID PluginID
	// Name is the plugin display name.
	Name string
	// Version is the effective version.
	Version string
	// Type is the plugin top-level type.
	Type pluginv1.PluginType
	// Description is the plugin description or remark.
	Description string
	// Installed reports whether the plugin is installed.
	Installed bool
	// Enabled reports whether the plugin is globally enabled.
	Enabled bool
	// ScopeNature is the plugin scope nature.
	ScopeNature pluginv1.ScopeNature
	// InstallMode is the plugin install mode.
	InstallMode pluginv1.InstallMode
	// TenantEnabled reports the plugin enablement in the requested tenant.
	TenantEnabled bool
}

// ListInput describes bounded plugin-governance info listing.
type ListInput struct {
	// Keyword matches plugin ID, name or description.
	Keyword string
	// PluginID filters by plugin ID fragment.
	PluginID string
	// Name filters by plugin display name fragment.
	Name string
	// Type filters by top-level plugin type.
	Type pluginv1.PluginType
	// Enabled optionally filters by global enablement status.
	Enabled *bool
	// Page constrains page number, page size and optional limit.
	Page capmodel.PageRequest
}

// TenantListInput describes bounded tenant plugin info search.
type TenantListInput struct {
	// Keyword matches plugin ID, name or description.
	Keyword string
	// Type filters by top-level plugin type.
	Type pluginv1.PluginType
	// TenantEnabled optionally filters by tenant enablement.
	TenantEnabled *bool
	// Page constrains page number, page size and optional limit.
	Page capmodel.PageRequest
}

const (
	// MaxPluginSearchPageSize is the maximum plugin search page size.
	MaxPluginSearchPageSize = 200
)

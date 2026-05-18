// This file defines source-plugin visible plugin lifecycle orchestration
// contracts for tenant-scoped plugin governance.

package contract

import "context"

// PluginLifecycleService exposes host-owned lifecycle orchestration to source
// plugins that own tenant or plugin-governance modules.
type PluginLifecycleService interface {
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

// PluginLifecycleRunner defines the host capability used by plugin lifecycle
// orchestration adapters.
type PluginLifecycleRunner interface {
	// EnsureTenantPluginDisableAllowed runs plugin lifecycle preconditions before
	// tenant-scoped plugin disable.
	EnsureTenantPluginDisableAllowed(ctx context.Context, pluginID string, tenantID int) error
	// NotifyTenantPluginDisabled runs best-effort notifications after
	// tenant-scoped plugin disable.
	NotifyTenantPluginDisabled(ctx context.Context, pluginID string, tenantID int)
	// EnsureTenantDeleteAllowed runs plugin lifecycle preconditions before
	// tenant deletion.
	EnsureTenantDeleteAllowed(ctx context.Context, tenantID int) error
	// NotifyTenantDeleted runs best-effort notifications after tenant deletion.
	NotifyTenantDeleted(ctx context.Context, tenantID int)
}

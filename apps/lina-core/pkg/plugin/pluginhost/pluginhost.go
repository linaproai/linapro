// Package pluginhost defines the public backend extension contracts that source
// plugins use to register routes, hooks, scheduled jobs, and access-control
// callbacks through grouped facade interfaces.
package pluginhost

import (
	"io/fs"

	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

const (
	// PluginAPINamespaceSegment is the first URL path segment reserved for plugin APIs.
	PluginAPINamespaceSegment = "x"
	// PluginAPINamespacePrefix is the public URL prefix reserved for plugin APIs.
	PluginAPINamespacePrefix = "/" + PluginAPINamespaceSegment
	// HostedAssetPathSegment is the first URL path segment for host-served plugin public assets.
	HostedAssetPathSegment = "x-assets"
	// HostedAssetURLPrefix is the public URL prefix for host-served plugin public assets.
	HostedAssetURLPrefix = "/" + HostedAssetPathSegment + "/"
	// DynamicPageComponentPath is the workbench component used by dynamic plugin pages.
	DynamicPageComponentPath = "system/plugin/dynamic-page"
	// DynamicEmbeddedSourceQueryKey is the menu query key carrying an embedded asset URL.
	DynamicEmbeddedSourceQueryKey = "embeddedSrc"
	// DynamicAccessModeQueryKey is the menu query key controlling dynamic plugin page access mode.
	DynamicAccessModeQueryKey = "pluginAccessMode"
	// DynamicAccessModeEmbeddedMount is the access mode for ESM-mounted dynamic plugin pages.
	DynamicAccessModeEmbeddedMount = "embedded-mount"
)

// Declarations defines the grouped declaration-time contract published to
// source plugins during compile-time registration.
type Declarations interface {
	// ID returns the stable plugin identifier that must match `plugin.yaml`.
	ID() string
	// Assets returns the plugin asset registration facade.
	Assets() AssetDeclarations
	// Lifecycle returns the plugin lifecycle callback registration facade.
	Lifecycle() LifecycleDeclarations
	// Hooks returns the event-hook registration facade.
	Hooks() HookDeclarations
	// HTTP returns the HTTP registration facade.
	HTTP() HTTPDeclarations
	// Jobs returns the scheduled-job registration facade.
	Jobs() JobDeclarations
	// Providers returns the framework capability provider declaration facade.
	Providers() ProviderDeclarations
	// Access returns the menu and permission access-control registration facade.
	Access() AccessDeclarations
}

// AssetDeclarations exposes plugin-owned asset declarations grouped under one
// dedicated facade.
type AssetDeclarations interface {
	// UseEmbeddedFiles binds one plugin-owned embedded filesystem.
	UseEmbeddedFiles(fileSystem fs.FS)
}

// LifecycleDeclarations exposes lifecycle callback registrations grouped under
// one dedicated facade.
type LifecycleDeclarations interface {
	// RegisterBeforeInstallHandler registers a pre-install lifecycle callback
	// for the source plugin. The host invokes this callback before it applies
	// install SQL, synchronizes plugin governance resources, or marks the plugin
	// as installed. Return ok=false to veto installation with a stable reason
	// key, or return an error when the precondition check itself failed. Use this
	// hook when installation depends on external configuration, tenant readiness,
	// license state, host capability checks, or other conditions that must be
	// satisfied before any install side effects are written.
	RegisterBeforeInstallHandler(handler SourcePluginBeforeLifecycleHandler) error
	// RegisterAfterInstallHandler registers a post-install lifecycle callback
	// for the source plugin. The host invokes this callback after install SQL,
	// governance synchronization, registry state update, release synchronization,
	// metadata synchronization, and cache refresh signals have completed. Use it
	// for follow-up work that observes a successful install, such as warming
	// plugin-local caches, emitting telemetry, or scheduling asynchronous
	// reconciliation. A failure is logged by the host and does not roll back the
	// already-effective installation.
	RegisterAfterInstallHandler(handler SourcePluginAfterLifecycleHandler) error
	// RegisterBeforeUpgradeHandler registers a pre-upgrade lifecycle callback
	// for the source plugin. The host invokes this callback after it has built
	// the upgrade plan and before it runs the plugin's custom upgrade handler,
	// upgrade SQL, governance synchronization, release switch, or cache
	// invalidation. Return ok=false to stop the upgrade with a stable reason key.
	// Use this hook to validate compatibility between the effective manifest and
	// the discovered target manifest, block unsupported version jumps, verify
	// required host services, or enforce plugin-specific migration prerequisites.
	RegisterBeforeUpgradeHandler(handler SourcePluginBeforeUpgradeHandler) error
	// RegisterUpgradeHandler registers the plugin-owned upgrade callback that
	// runs during a source-plugin runtime upgrade. The host invokes this callback
	// after all pre-upgrade callbacks allow the operation and before it executes
	// upgrade SQL and promotes the target release. Use this hook for custom,
	// version-aware migration work that cannot be represented by manifest SQL
	// alone, such as transforming plugin-owned data, preparing external
	// resources, or bridging data between old and new plugin contracts. The
	// callback should be idempotent or safely retryable because a failed upgrade
	// can be retried by an operator.
	RegisterUpgradeHandler(handler SourcePluginUpgradeHandler) error
	// RegisterAfterUpgradeHandler registers a post-upgrade lifecycle callback
	// for the source plugin. The host invokes this callback after upgrade SQL,
	// governance synchronization, release promotion, node-state synchronization,
	// and cache refresh signals have completed successfully. Use this hook for
	// best-effort follow-up work that observes the new effective version, such
	// as warming plugin caches, emitting plugin-local telemetry, refreshing
	// external integrations, or scheduling asynchronous reconciliation. A failure
	// is logged by the host and does not roll back the already-effective upgrade.
	RegisterAfterUpgradeHandler(handler SourcePluginUpgradeHandler) error
	// RegisterBeforeDisableHandler registers a pre-disable lifecycle callback
	// for the source plugin. The host invokes this callback before changing the
	// plugin from enabled to disabled and before business entry points are hidden
	// or stopped. Return ok=false to veto the disable operation with a stable
	// reason key. Use this hook when the plugin must prevent disable while jobs,
	// workflows, external subscriptions, tenant obligations, or other
	// plugin-owned runtime work is still active.
	RegisterBeforeDisableHandler(handler SourcePluginBeforeLifecycleHandler) error
	// RegisterAfterDisableHandler registers a post-disable lifecycle callback
	// for the source plugin. The host invokes this callback after the plugin has
	// been disabled, business entry points have been hidden or stopped, cache
	// refresh signals have completed, and lifecycle observers have been notified.
	// Use it for best-effort follow-up work such as closing external sessions,
	// emitting telemetry, or scheduling reconciliation. A failure is logged by
	// the host and does not roll back the disable operation.
	RegisterAfterDisableHandler(handler SourcePluginAfterLifecycleHandler) error
	// RegisterBeforeUninstallHandler registers a pre-uninstall lifecycle callback
	// for the source plugin. The host invokes this callback before it runs
	// plugin cleanup, uninstall SQL, governance resource deletion, registry state
	// changes, and uninstall hook events. Return ok=false to veto normal
	// uninstall with a stable reason key; force uninstall may bypass the veto
	// only when the host configuration explicitly permits it. Use this hook to
	// protect plugin-owned data, block uninstall while dependent resources still
	// exist, require operator confirmation outside the host, or verify that
	// external cleanup prerequisites are satisfied.
	RegisterBeforeUninstallHandler(handler SourcePluginBeforeLifecycleHandler) error
	// RegisterAfterUninstallHandler registers a post-uninstall lifecycle callback
	// for the source plugin. The host invokes this callback after plugin cleanup,
	// uninstall SQL when requested, governance deletion, registry state update,
	// release synchronization, metadata synchronization, cache refresh signals,
	// and lifecycle observers have completed. Use it for best-effort telemetry or
	// external reconciliation that should observe the final uninstalled state. A
	// failure is logged by the host and does not roll back uninstall.
	RegisterAfterUninstallHandler(handler SourcePluginAfterLifecycleHandler) error
	// RegisterBeforeTenantDisableHandler registers a tenant-scoped pre-disable
	// lifecycle callback for the source plugin. The host invokes this callback
	// before disabling the plugin for one tenant while leaving global plugin
	// installation state intact. Return ok=false to veto the tenant-scoped
	// disable with a stable reason key. Use this hook when tenant-specific
	// plugin activity, subscriptions, pending work, or data retention policy
	// must be checked before removing that tenant's access to the plugin.
	RegisterBeforeTenantDisableHandler(handler SourcePluginBeforeTenantLifecycleHandler) error
	// RegisterAfterTenantDisableHandler registers a tenant-scoped post-disable
	// lifecycle callback for the source plugin. The host invokes this callback
	// after one tenant has successfully lost access to the plugin. Use it for
	// tenant-local cache warming, telemetry, or external reconciliation. A
	// failure is logged by the host and does not roll back tenant disable.
	RegisterAfterTenantDisableHandler(handler SourcePluginAfterTenantLifecycleHandler) error
	// RegisterBeforeTenantDeleteHandler registers a tenant-delete precondition
	// callback for the source plugin. The host invokes this callback before a
	// tenant is deleted so installed plugins can protect tenant-owned plugin
	// data and external resources. Return ok=false to block tenant deletion with
	// a stable reason key. Use this hook when the plugin stores tenant-scoped
	// records, owns external tenant mappings, runs tenant-specific jobs, or must
	// require explicit cleanup before the tenant can be removed.
	RegisterBeforeTenantDeleteHandler(handler SourcePluginBeforeTenantLifecycleHandler) error
	// RegisterAfterTenantDeleteHandler registers a tenant-delete post-notification
	// callback for the source plugin. The host invokes this callback after the
	// tenant has been deleted and plugin-owned preconditions have passed. Use it
	// for best-effort cleanup of external tenant mappings or telemetry. A failure
	// is logged by the host and does not roll back tenant deletion.
	RegisterAfterTenantDeleteHandler(handler SourcePluginAfterTenantLifecycleHandler) error
	// RegisterBeforeInstallModeChangeHandler registers a precondition callback
	// for source-plugin install-mode transitions. The host invokes this callback
	// before switching the plugin between supported install modes, such as
	// global and tenant-scoped modes. Return ok=false to veto the transition with
	// a stable reason key. Use this hook when a mode change would alter tenant
	// visibility, data ownership, governance resources, or runtime assumptions
	// and the plugin must verify that existing data and active tenants can be
	// migrated or safely preserved.
	RegisterBeforeInstallModeChangeHandler(handler SourcePluginBeforeInstallModeChangeHandler) error
	// RegisterAfterInstallModeChangeHandler registers a post-notification
	// callback for source-plugin install-mode transitions. The host invokes this
	// callback after an install-mode transition succeeds. Use it for follow-up
	// reconciliation, telemetry, or cache warming that observes the new mode. A
	// failure is logged by the host and does not roll back the mode change.
	RegisterAfterInstallModeChangeHandler(handler SourcePluginAfterInstallModeChangeHandler) error
	// RegisterUninstallHandler registers the plugin-owned cleanup callback that
	// runs during uninstall when the operator requested storage/data purging. The
	// host invokes this callback after uninstall preconditions have passed and
	// before uninstall SQL removes plugin-owned tables. Use this hook to delete
	// or detach external resources, remove plugin-managed files, revoke external
	// subscriptions, or perform cleanup that cannot be expressed as uninstall
	// SQL. The callback should be idempotent because uninstall can be retried
	// after cleanup or SQL failures.
	RegisterUninstallHandler(handler SourcePluginUninstallHandler) error
}

// HookDeclarations exposes callback-style host hook registrations grouped
// under one dedicated facade.
type HookDeclarations interface {
	// RegisterHook registers one callback-style host hook handler.
	RegisterHook(point ExtensionPoint, mode CallbackExecutionMode, handler HookHandler) error
}

// HTTPDeclarations exposes HTTP-adjacent registrations grouped under one
// dedicated facade.
type HTTPDeclarations interface {
	// RegisterRoutes registers one callback that contributes plugin-owned HTTP routes.
	RegisterRoutes(point ExtensionPoint, mode CallbackExecutionMode, handler RouteRegisterHandler) error
}

// JobDeclarations exposes scheduled-job registrations grouped under one dedicated
// facade.
type JobDeclarations interface {
	// RegisterJobs registers one callback that contributes plugin-owned scheduled jobs.
	RegisterJobs(point ExtensionPoint, mode CallbackExecutionMode, handler JobRegisterHandler) error
}

// AccessDeclarations exposes menu and permission access-control callback
// registrations grouped under one dedicated facade.
type AccessDeclarations interface {
	// RegisterMenuFilter registers one callback that filters host menus.
	RegisterMenuFilter(point ExtensionPoint, mode CallbackExecutionMode, handler MenuFilterHandler) error
	// RegisterPermissionFilter registers one callback that filters host permissions.
	RegisterPermissionFilter(point ExtensionPoint, mode CallbackExecutionMode, handler PermissionFilterHandler) error
}

// ProviderDeclarations exposes framework capability provider factory declarations.
type ProviderDeclarations interface {
	// ProvideTenant declares the tenant capability provider factory implemented by this source plugin.
	ProvideTenant(factory tenantspi.ProviderFactory) error
	// ProvideOrg declares the organization capability provider factory implemented by this source plugin.
	ProvideOrg(factory orgspi.ProviderFactory) error
	// ProvideCapability declares one plugin-owned capability descriptor published by this source plugin.
	ProvideCapability(descriptor capregistry.Descriptor) error
}

// SourcePluginDefinition exposes the host-side read model restored from one
// grouped source-plugin registration.
type SourcePluginDefinition interface {
	Declarations
	// GetEmbeddedFiles returns the plugin-owned embedded filesystem when declared.
	GetEmbeddedFiles() fs.FS
	// GetHookHandlers returns the registered callback-style hook handlers.
	GetHookHandlers() []*HookHandlerRegistration
	// GetRouteRegistrars returns the registered route contribution callbacks.
	GetRouteRegistrars() []*RouteHandlerRegistration
	// GetJobRegistrars returns the registered scheduled-job contribution callbacks.
	GetJobRegistrars() []*JobHandlerRegistration
	// GetMenuFilters returns the registered menu filter callbacks.
	GetMenuFilters() []*MenuFilterHandlerRegistration
	// GetPermissionFilters returns the registered permission filter callbacks.
	GetPermissionFilters() []*PermissionFilterHandlerRegistration
	// GetTenantProviderFactory returns the declared tenant provider factory.
	GetTenantProviderFactory() tenantspi.ProviderFactory
	// GetOrgProviderFactory returns the declared organization provider factory.
	GetOrgProviderFactory() orgspi.ProviderFactory
	// GetCapabilityDescriptors returns plugin-owned capability descriptors declared by this source plugin.
	GetCapabilityDescriptors() []capregistry.Descriptor
	// GetBeforeInstallHandler returns the registered pre-install veto callback.
	GetBeforeInstallHandler() SourcePluginBeforeLifecycleHandler
	// GetAfterInstallHandler returns the registered post-install callback.
	GetAfterInstallHandler() SourcePluginAfterLifecycleHandler
	// GetBeforeUpgradeHandler returns the registered pre-upgrade veto callback.
	GetBeforeUpgradeHandler() SourcePluginBeforeUpgradeHandler
	// GetUpgradeHandler returns the registered source-plugin custom upgrade callback.
	GetUpgradeHandler() SourcePluginUpgradeHandler
	// GetAfterUpgradeHandler returns the registered post-upgrade event callback.
	GetAfterUpgradeHandler() SourcePluginUpgradeHandler
	// GetBeforeDisableHandler returns the registered pre-disable veto callback.
	GetBeforeDisableHandler() SourcePluginBeforeLifecycleHandler
	// GetAfterDisableHandler returns the registered post-disable callback.
	GetAfterDisableHandler() SourcePluginAfterLifecycleHandler
	// GetBeforeUninstallHandler returns the registered pre-uninstall veto callback.
	GetBeforeUninstallHandler() SourcePluginBeforeLifecycleHandler
	// GetAfterUninstallHandler returns the registered post-uninstall callback.
	GetAfterUninstallHandler() SourcePluginAfterLifecycleHandler
	// GetBeforeTenantDisableHandler returns the registered tenant-disable veto callback.
	GetBeforeTenantDisableHandler() SourcePluginBeforeTenantLifecycleHandler
	// GetAfterTenantDisableHandler returns the registered tenant-disable post callback.
	GetAfterTenantDisableHandler() SourcePluginAfterTenantLifecycleHandler
	// GetBeforeTenantDeleteHandler returns the registered tenant-delete veto callback.
	GetBeforeTenantDeleteHandler() SourcePluginBeforeTenantLifecycleHandler
	// GetAfterTenantDeleteHandler returns the registered tenant-delete post callback.
	GetAfterTenantDeleteHandler() SourcePluginAfterTenantLifecycleHandler
	// GetBeforeInstallModeChangeHandler returns the registered install-mode change veto callback.
	GetBeforeInstallModeChangeHandler() SourcePluginBeforeInstallModeChangeHandler
	// GetAfterInstallModeChangeHandler returns the registered install-mode change post callback.
	GetAfterInstallModeChangeHandler() SourcePluginAfterInstallModeChangeHandler
	// GetUninstallHandler returns the registered source-plugin uninstall cleanup callback.
	GetUninstallHandler() SourcePluginUninstallHandler
}

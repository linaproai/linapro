// This file defines the published source-plugin interfaces and grouped
// registration facades exposed to source-plugin authors.

package pluginhost

import (
	"io/fs"

	"github.com/gogf/gf/v2/errors/gerror"
)

// SourcePlugin defines the grouped plugin-facing contract published to source
// plugins during compile-time registration.
type SourcePlugin interface {
	// ID returns the stable plugin identifier that must match `plugin.yaml`.
	ID() string
	// Assets returns the plugin asset registration facade.
	Assets() SourcePluginAssets
	// Lifecycle returns the plugin lifecycle callback registration facade.
	Lifecycle() SourcePluginLifecycle
	// Hooks returns the event-hook registration facade.
	Hooks() SourcePluginHooks
	// HTTP returns the HTTP registration facade.
	HTTP() SourcePluginHTTP
	// Cron returns the cron registration facade.
	Cron() SourcePluginCron
	// Governance returns the menu and permission governance registration facade.
	Governance() SourcePluginGovernance
}

// SourcePluginAssets exposes plugin-owned asset declarations grouped under one
// dedicated facade.
type SourcePluginAssets interface {
	// UseEmbeddedFiles binds one plugin-owned embedded filesystem.
	UseEmbeddedFiles(fileSystem fs.FS)
}

// SourcePluginLifecycle exposes lifecycle callback registrations grouped under
// one dedicated facade.
type SourcePluginLifecycle interface {
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

// SourcePluginHooks exposes callback-style host hook registrations grouped
// under one dedicated facade.
type SourcePluginHooks interface {
	// RegisterHook registers one callback-style host hook handler.
	RegisterHook(point ExtensionPoint, mode CallbackExecutionMode, handler HookHandler) error
}

// SourcePluginHTTP exposes HTTP-adjacent registrations grouped under one
// dedicated facade.
type SourcePluginHTTP interface {
	// RegisterRoutes registers one callback that contributes plugin-owned HTTP routes.
	RegisterRoutes(point ExtensionPoint, mode CallbackExecutionMode, handler RouteRegisterHandler) error
}

// SourcePluginCron exposes cron registrations grouped under one dedicated
// facade.
type SourcePluginCron interface {
	// RegisterCron registers one callback that contributes plugin-owned cron jobs.
	RegisterCron(point ExtensionPoint, mode CallbackExecutionMode, handler CronRegisterHandler) error
}

// SourcePluginGovernance exposes governance callback registrations grouped
// under one dedicated facade.
type SourcePluginGovernance interface {
	// RegisterMenuFilter registers one callback that filters host menus.
	RegisterMenuFilter(point ExtensionPoint, mode CallbackExecutionMode, handler MenuFilterHandler) error
	// RegisterPermissionFilter registers one callback that filters host permissions.
	RegisterPermissionFilter(point ExtensionPoint, mode CallbackExecutionMode, handler PermissionFilterHandler) error
}

// SourcePluginDefinition exposes the host-side read model restored from one
// grouped source-plugin registration.
type SourcePluginDefinition interface {
	SourcePlugin
	// GetEmbeddedFiles returns the plugin-owned embedded filesystem when declared.
	GetEmbeddedFiles() fs.FS
	// GetHookHandlers returns the registered callback-style hook handlers.
	GetHookHandlers() []*HookHandlerRegistration
	// GetRouteRegistrars returns the registered route contribution callbacks.
	GetRouteRegistrars() []*RouteHandlerRegistration
	// GetCronRegistrars returns the registered cron contribution callbacks.
	GetCronRegistrars() []*CronHandlerRegistration
	// GetMenuFilters returns the registered menu filter callbacks.
	GetMenuFilters() []*MenuFilterHandlerRegistration
	// GetPermissionFilters returns the registered permission filter callbacks.
	GetPermissionFilters() []*PermissionFilterHandlerRegistration
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

// sourcePluginAssets is the asset-registration facade bound to one source
// plugin definition.
type sourcePluginAssets struct {
	plugin *sourcePlugin
}

// sourcePluginLifecycle is the lifecycle-registration facade bound to one
// source plugin definition.
type sourcePluginLifecycle struct {
	plugin *sourcePlugin
}

// sourcePluginHooks is the hook-registration facade bound to one source plugin
// definition.
type sourcePluginHooks struct {
	plugin *sourcePlugin
}

// sourcePluginHTTP is the HTTP-registration facade bound to one source plugin
// definition.
type sourcePluginHTTP struct {
	plugin *sourcePlugin
}

// sourcePluginCron is the cron-registration facade bound to one source plugin
// definition.
type sourcePluginCron struct {
	plugin *sourcePlugin
}

// sourcePluginGovernance is the governance-registration facade bound to one
// source plugin definition.
type sourcePluginGovernance struct {
	plugin *sourcePlugin
}

// ID returns the stable plugin identifier declared by the source plugin.
func (p *sourcePlugin) ID() string {
	if p == nil {
		return ""
	}
	return p.id
}

// Assets returns the plugin asset registration facade.
func (p *sourcePlugin) Assets() SourcePluginAssets {
	if p == nil {
		return nil
	}
	return p.assets
}

// Lifecycle returns the plugin lifecycle callback registration facade.
func (p *sourcePlugin) Lifecycle() SourcePluginLifecycle {
	if p == nil {
		return nil
	}
	return p.lifecycle
}

// Hooks returns the event-hook registration facade.
func (p *sourcePlugin) Hooks() SourcePluginHooks {
	if p == nil {
		return nil
	}
	return p.hooks
}

// HTTP returns the HTTP registration facade.
func (p *sourcePlugin) HTTP() SourcePluginHTTP {
	if p == nil {
		return nil
	}
	return p.http
}

// Cron returns the cron registration facade.
func (p *sourcePlugin) Cron() SourcePluginCron {
	if p == nil {
		return nil
	}
	return p.cron
}

// Governance returns the menu and permission governance registration facade.
func (p *sourcePlugin) Governance() SourcePluginGovernance {
	if p == nil {
		return nil
	}
	return p.governance
}

// UseEmbeddedFiles binds one plugin-owned embedded filesystem.
func (r *sourcePluginAssets) UseEmbeddedFiles(fileSystem fs.FS) {
	if r == nil || r.plugin == nil {
		return
	}
	r.plugin.useEmbeddedFiles(fileSystem)
}

// RegisterUninstallHandler registers one uninstall cleanup callback.
func (r *sourcePluginLifecycle) RegisterUninstallHandler(handler SourcePluginUninstallHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerUninstallHandler(handler)
}

// RegisterBeforeInstallHandler registers one pre-install veto callback.
func (r *sourcePluginLifecycle) RegisterBeforeInstallHandler(handler SourcePluginBeforeLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerBeforeInstallHandler(handler)
}

// RegisterAfterInstallHandler registers one post-install callback.
func (r *sourcePluginLifecycle) RegisterAfterInstallHandler(handler SourcePluginAfterLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerAfterInstallHandler(handler)
}

// RegisterBeforeUpgradeHandler registers one pre-upgrade veto callback.
func (r *sourcePluginLifecycle) RegisterBeforeUpgradeHandler(handler SourcePluginBeforeUpgradeHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerBeforeUpgradeHandler(handler)
}

// RegisterUpgradeHandler registers one source-plugin custom upgrade callback.
func (r *sourcePluginLifecycle) RegisterUpgradeHandler(handler SourcePluginUpgradeHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerUpgradeHandler(handler)
}

// RegisterAfterUpgradeHandler registers one post-upgrade event callback.
func (r *sourcePluginLifecycle) RegisterAfterUpgradeHandler(handler SourcePluginUpgradeHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerAfterUpgradeHandler(handler)
}

// RegisterBeforeDisableHandler registers one pre-disable veto callback.
func (r *sourcePluginLifecycle) RegisterBeforeDisableHandler(handler SourcePluginBeforeLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerBeforeDisableHandler(handler)
}

// RegisterAfterDisableHandler registers one post-disable callback.
func (r *sourcePluginLifecycle) RegisterAfterDisableHandler(handler SourcePluginAfterLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerAfterDisableHandler(handler)
}

// RegisterBeforeUninstallHandler registers one pre-uninstall veto callback.
func (r *sourcePluginLifecycle) RegisterBeforeUninstallHandler(handler SourcePluginBeforeLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerBeforeUninstallHandler(handler)
}

// RegisterAfterUninstallHandler registers one post-uninstall callback.
func (r *sourcePluginLifecycle) RegisterAfterUninstallHandler(handler SourcePluginAfterLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerAfterUninstallHandler(handler)
}

// RegisterBeforeTenantDisableHandler registers one tenant-disable precondition callback.
func (r *sourcePluginLifecycle) RegisterBeforeTenantDisableHandler(handler SourcePluginBeforeTenantLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerBeforeTenantDisableHandler(handler)
}

// RegisterAfterTenantDisableHandler registers one tenant-disable post callback.
func (r *sourcePluginLifecycle) RegisterAfterTenantDisableHandler(handler SourcePluginAfterTenantLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerAfterTenantDisableHandler(handler)
}

// RegisterBeforeTenantDeleteHandler registers one tenant-delete precondition callback.
func (r *sourcePluginLifecycle) RegisterBeforeTenantDeleteHandler(handler SourcePluginBeforeTenantLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerBeforeTenantDeleteHandler(handler)
}

// RegisterAfterTenantDeleteHandler registers one tenant-delete post callback.
func (r *sourcePluginLifecycle) RegisterAfterTenantDeleteHandler(handler SourcePluginAfterTenantLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerAfterTenantDeleteHandler(handler)
}

// RegisterBeforeInstallModeChangeHandler registers one install-mode change precondition callback.
func (r *sourcePluginLifecycle) RegisterBeforeInstallModeChangeHandler(handler SourcePluginBeforeInstallModeChangeHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerBeforeInstallModeChangeHandler(handler)
}

// RegisterAfterInstallModeChangeHandler registers one install-mode change post callback.
func (r *sourcePluginLifecycle) RegisterAfterInstallModeChangeHandler(handler SourcePluginAfterInstallModeChangeHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerAfterInstallModeChangeHandler(handler)
}

// RegisterHook registers one callback-style host hook handler.
func (r *sourcePluginHooks) RegisterHook(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler HookHandler,
) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin hook facade is nil")
	}
	return r.plugin.registerHook(point, mode, handler)
}

// RegisterRoutes registers one callback that contributes plugin-owned HTTP routes.
func (r *sourcePluginHTTP) RegisterRoutes(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler RouteRegisterHandler,
) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin http facade is nil")
	}
	return r.plugin.registerRoutes(point, mode, handler)
}

// RegisterCron registers one callback that contributes plugin-owned cron jobs.
func (r *sourcePluginCron) RegisterCron(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler CronRegisterHandler,
) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin cron facade is nil")
	}
	return r.plugin.registerCron(point, mode, handler)
}

// RegisterMenuFilter registers one callback that filters host menus.
func (r *sourcePluginGovernance) RegisterMenuFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler MenuFilterHandler,
) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin governance facade is nil")
	}
	return r.plugin.registerMenuFilter(point, mode, handler)
}

// RegisterPermissionFilter registers one callback that filters host permissions.
func (r *sourcePluginGovernance) RegisterPermissionFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler PermissionFilterHandler,
) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin governance facade is nil")
	}
	return r.plugin.registerPermissionFilter(point, mode, handler)
}

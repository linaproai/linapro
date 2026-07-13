// This file defines the published source-plugin interfaces and grouped
// registration facades exposed to source-plugin authors.

package pluginhost

import (
	"io/fs"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/authcap/extlogin/extidspi"
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

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

// sourcePluginJobs is the scheduled-job registration facade bound to one source plugin
// definition.
type sourcePluginJobs struct {
	plugin *sourcePlugin
}

// sourcePluginProviders is the framework provider declaration facade bound to one source plugin definition.
type sourcePluginProviders struct {
	plugin *sourcePlugin
}

// sourcePluginAccess is the access-control registration facade bound to one
// source plugin definition.
type sourcePluginAccess struct {
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
func (p *sourcePlugin) Assets() AssetDeclarations {
	if p == nil {
		return nil
	}
	return p.assets
}

// Lifecycle returns the plugin lifecycle callback registration facade.
func (p *sourcePlugin) Lifecycle() LifecycleDeclarations {
	if p == nil {
		return nil
	}
	return p.lifecycle
}

// Hooks returns the event-hook registration facade.
func (p *sourcePlugin) Hooks() HookDeclarations {
	if p == nil {
		return nil
	}
	return p.hooks
}

// HTTP returns the HTTP registration facade.
func (p *sourcePlugin) HTTP() HTTPDeclarations {
	if p == nil {
		return nil
	}
	return p.http
}

// Jobs returns the scheduled-job registration facade.
func (p *sourcePlugin) Jobs() JobDeclarations {
	if p == nil {
		return nil
	}
	return p.jobs
}

// Providers returns the framework capability provider declaration facade.
func (p *sourcePlugin) Providers() ProviderDeclarations {
	if p == nil {
		return nil
	}
	return p.providers
}

// Access returns the menu and permission access-control registration facade.
func (p *sourcePlugin) Access() AccessDeclarations {
	if p == nil {
		return nil
	}
	return p.access
}

// ProvideTenant declares one source-plugin tenant provider factory.
func (r *sourcePluginProviders) ProvideTenant(factory tenantspi.ProviderFactory) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin provider facade is nil")
	}
	return r.plugin.registerTenantProvider(factory)
}

// ProvideOrg declares one source-plugin organization provider factory.
func (r *sourcePluginProviders) ProvideOrg(factory orgspi.ProviderFactory) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin provider facade is nil")
	}
	return r.plugin.registerOrgProvider(factory)
}

// ProvideCapability declares one plugin-owned capability descriptor.
func (r *sourcePluginProviders) ProvideCapability(descriptor capregistry.Descriptor) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin provider facade is nil")
	}
	return r.plugin.registerCapabilityDescriptor(descriptor)
}

// ProvideExternalIdentity declares one source-plugin external-identity provider ID.
func (r *sourcePluginProviders) ProvideExternalIdentity(providerID string) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin provider facade is nil")
	}
	return r.plugin.registerExternalIdentityProvider(providerID)
}

// ProvideExternalIdentityProvider declares this source plugin's external-identity
// provider engine factory (linapro-extlogin-core). It is distinct from
// ProvideExternalIdentity, which stamps provider-ID ownership for calling
// plugins. A plugin that supplies the engine need not own any provider ID.
func (r *sourcePluginProviders) ProvideExternalIdentityProvider(factory extidspi.ProviderFactory) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin provider facade is nil")
	}
	return r.plugin.registerExternalIdentityProviderFactory(factory)
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

// RegisterBeforeEnableHandler registers one pre-enable veto callback.
func (r *sourcePluginLifecycle) RegisterBeforeEnableHandler(handler SourcePluginBeforeLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerBeforeEnableHandler(handler)
}

// RegisterAfterEnableHandler registers one post-enable callback.
func (r *sourcePluginLifecycle) RegisterAfterEnableHandler(handler SourcePluginAfterLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerAfterEnableHandler(handler)
}

// RegisterGlobalBeforeInstallHandler registers one global pre-install veto callback.
func (r *sourcePluginLifecycle) RegisterGlobalBeforeInstallHandler(handler SourcePluginGlobalLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerGlobalBeforeInstallHandler(handler)
}

// RegisterGlobalBeforeEnableHandler registers one global pre-enable veto callback.
func (r *sourcePluginLifecycle) RegisterGlobalBeforeEnableHandler(handler SourcePluginGlobalLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerGlobalBeforeEnableHandler(handler)
}

// RegisterGlobalBeforeDisableHandler registers one global pre-disable veto callback.
func (r *sourcePluginLifecycle) RegisterGlobalBeforeDisableHandler(handler SourcePluginGlobalLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerGlobalBeforeDisableHandler(handler)
}

// RegisterGlobalBeforeUninstallHandler registers one global pre-uninstall veto callback.
func (r *sourcePluginLifecycle) RegisterGlobalBeforeUninstallHandler(handler SourcePluginGlobalLifecycleHandler) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return r.plugin.registerGlobalBeforeUninstallHandler(handler)
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

// RegisterJobs registers one callback that contributes plugin-owned scheduled jobs.
func (r *sourcePluginJobs) RegisterJobs(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler JobRegisterHandler,
) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin jobs facade is nil")
	}
	return r.plugin.registerJobs(point, mode, handler)
}

// RegisterMenuFilter registers one callback that filters host menus.
func (r *sourcePluginAccess) RegisterMenuFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler MenuFilterHandler,
) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin access facade is nil")
	}
	return r.plugin.registerMenuFilter(point, mode, handler)
}

// RegisterPermissionFilter registers one callback that filters host permissions.
func (r *sourcePluginAccess) RegisterPermissionFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler PermissionFilterHandler,
) error {
	if r == nil || r.plugin == nil {
		return gerror.New("pluginhost: source plugin access facade is nil")
	}
	return r.plugin.registerPermissionFilter(point, mode, handler)
}

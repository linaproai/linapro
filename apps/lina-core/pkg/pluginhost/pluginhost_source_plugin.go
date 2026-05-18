// This file defines the host-owned source-plugin registration storage used by
// grouped source-plugin facade implementations.

package pluginhost

import "io/fs"

// sourcePlugin stores one compile-time source plugin definition behind the
// published grouped SourcePlugin interfaces.
type sourcePlugin struct {
	// id is the stable plugin id and must match `plugin.yaml`.
	id string
	// assets exposes grouped asset registration helpers.
	assets SourcePluginAssets
	// lifecycle exposes grouped lifecycle registration helpers.
	lifecycle SourcePluginLifecycle
	// hooks exposes grouped hook registration helpers.
	hooks SourcePluginHooks
	// http exposes grouped HTTP registration helpers.
	http SourcePluginHTTP
	// cron exposes grouped cron registration helpers.
	cron SourcePluginCron
	// governance exposes grouped menu and permission governance helpers.
	governance SourcePluginGovernance

	embeddedFiles     fs.FS
	beforeInstall     SourcePluginBeforeLifecycleHandler
	afterInstall      SourcePluginAfterLifecycleHandler
	beforeUpgrade     SourcePluginBeforeUpgradeHandler
	upgradeHandler    SourcePluginUpgradeHandler
	afterUpgrade      SourcePluginUpgradeHandler
	beforeDisable     SourcePluginBeforeLifecycleHandler
	afterDisable      SourcePluginAfterLifecycleHandler
	beforeUninstall   SourcePluginBeforeLifecycleHandler
	afterUninstall    SourcePluginAfterLifecycleHandler
	beforeTenantDis   SourcePluginBeforeTenantLifecycleHandler
	afterTenantDis    SourcePluginAfterTenantLifecycleHandler
	beforeTenantDel   SourcePluginBeforeTenantLifecycleHandler
	afterTenantDel    SourcePluginAfterTenantLifecycleHandler
	beforeModeChange  SourcePluginBeforeInstallModeChangeHandler
	afterModeChange   SourcePluginAfterInstallModeChangeHandler
	uninstallHandler  SourcePluginUninstallHandler
	hookHandlers      []*HookHandlerRegistration
	routeRegistrars   []*RouteHandlerRegistration
	cronRegistrars    []*CronHandlerRegistration
	menuFilters       []*MenuFilterHandlerRegistration
	permissionFilters []*PermissionFilterHandlerRegistration
}

// NewSourcePlugin creates and returns a new grouped source plugin definition.
func NewSourcePlugin(id string) SourcePlugin {
	plugin := &sourcePlugin{
		id:                id,
		hookHandlers:      make([]*HookHandlerRegistration, 0),
		routeRegistrars:   make([]*RouteHandlerRegistration, 0),
		cronRegistrars:    make([]*CronHandlerRegistration, 0),
		menuFilters:       make([]*MenuFilterHandlerRegistration, 0),
		permissionFilters: make([]*PermissionFilterHandlerRegistration, 0),
	}
	plugin.assets = &sourcePluginAssets{plugin: plugin}
	plugin.lifecycle = &sourcePluginLifecycle{plugin: plugin}
	plugin.hooks = &sourcePluginHooks{plugin: plugin}
	plugin.http = &sourcePluginHTTP{plugin: plugin}
	plugin.cron = &sourcePluginCron{plugin: plugin}
	plugin.governance = &sourcePluginGovernance{plugin: plugin}
	return plugin
}

// useEmbeddedFiles binds one plugin-owned embedded filesystem to the source plugin.
func (p *sourcePlugin) useEmbeddedFiles(fileSystem fs.FS) {
	if p == nil {
		return
	}
	p.embeddedFiles = fileSystem
}

// GetEmbeddedFiles returns the plugin-owned embedded filesystem when declared.
func (p *sourcePlugin) GetEmbeddedFiles() fs.FS {
	if p == nil {
		return nil
	}
	return p.embeddedFiles
}

// GetHookHandlers returns the registered callback-style hook handlers.
func (p *sourcePlugin) GetHookHandlers() []*HookHandlerRegistration {
	if p == nil {
		return []*HookHandlerRegistration{}
	}
	items := make([]*HookHandlerRegistration, len(p.hookHandlers))
	copy(items, p.hookHandlers)
	return items
}

// GetRouteRegistrars returns the registered route contribution callbacks.
func (p *sourcePlugin) GetRouteRegistrars() []*RouteHandlerRegistration {
	if p == nil {
		return []*RouteHandlerRegistration{}
	}
	items := make([]*RouteHandlerRegistration, len(p.routeRegistrars))
	copy(items, p.routeRegistrars)
	return items
}

// GetCronRegistrars returns the registered cron contribution callbacks.
func (p *sourcePlugin) GetCronRegistrars() []*CronHandlerRegistration {
	if p == nil {
		return []*CronHandlerRegistration{}
	}
	items := make([]*CronHandlerRegistration, len(p.cronRegistrars))
	copy(items, p.cronRegistrars)
	return items
}

// GetMenuFilters returns the registered menu filter callbacks.
func (p *sourcePlugin) GetMenuFilters() []*MenuFilterHandlerRegistration {
	if p == nil {
		return []*MenuFilterHandlerRegistration{}
	}
	items := make([]*MenuFilterHandlerRegistration, len(p.menuFilters))
	copy(items, p.menuFilters)
	return items
}

// GetPermissionFilters returns the registered permission filter callbacks.
func (p *sourcePlugin) GetPermissionFilters() []*PermissionFilterHandlerRegistration {
	if p == nil {
		return []*PermissionFilterHandlerRegistration{}
	}
	items := make([]*PermissionFilterHandlerRegistration, len(p.permissionFilters))
	copy(items, p.permissionFilters)
	return items
}

// GetBeforeInstallHandler returns the registered source-plugin pre-install callback.
func (p *sourcePlugin) GetBeforeInstallHandler() SourcePluginBeforeLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.beforeInstall
}

// GetAfterInstallHandler returns the registered source-plugin post-install callback.
func (p *sourcePlugin) GetAfterInstallHandler() SourcePluginAfterLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.afterInstall
}

// GetBeforeUpgradeHandler returns the registered source-plugin pre-upgrade callback.
func (p *sourcePlugin) GetBeforeUpgradeHandler() SourcePluginBeforeUpgradeHandler {
	if p == nil {
		return nil
	}
	return p.beforeUpgrade
}

// GetUpgradeHandler returns the registered source-plugin custom upgrade callback.
func (p *sourcePlugin) GetUpgradeHandler() SourcePluginUpgradeHandler {
	if p == nil {
		return nil
	}
	return p.upgradeHandler
}

// GetAfterUpgradeHandler returns the registered source-plugin post-upgrade callback.
func (p *sourcePlugin) GetAfterUpgradeHandler() SourcePluginUpgradeHandler {
	if p == nil {
		return nil
	}
	return p.afterUpgrade
}

// GetBeforeDisableHandler returns the registered source-plugin pre-disable callback.
func (p *sourcePlugin) GetBeforeDisableHandler() SourcePluginBeforeLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.beforeDisable
}

// GetAfterDisableHandler returns the registered source-plugin post-disable callback.
func (p *sourcePlugin) GetAfterDisableHandler() SourcePluginAfterLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.afterDisable
}

// GetBeforeUninstallHandler returns the registered source-plugin pre-uninstall callback.
func (p *sourcePlugin) GetBeforeUninstallHandler() SourcePluginBeforeLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.beforeUninstall
}

// GetAfterUninstallHandler returns the registered source-plugin post-uninstall callback.
func (p *sourcePlugin) GetAfterUninstallHandler() SourcePluginAfterLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.afterUninstall
}

// GetBeforeTenantDisableHandler returns the registered source-plugin tenant-disable callback.
func (p *sourcePlugin) GetBeforeTenantDisableHandler() SourcePluginBeforeTenantLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.beforeTenantDis
}

// GetAfterTenantDisableHandler returns the registered source-plugin post-tenant-disable callback.
func (p *sourcePlugin) GetAfterTenantDisableHandler() SourcePluginAfterTenantLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.afterTenantDis
}

// GetBeforeTenantDeleteHandler returns the registered source-plugin tenant-delete callback.
func (p *sourcePlugin) GetBeforeTenantDeleteHandler() SourcePluginBeforeTenantLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.beforeTenantDel
}

// GetAfterTenantDeleteHandler returns the registered source-plugin post-tenant-delete callback.
func (p *sourcePlugin) GetAfterTenantDeleteHandler() SourcePluginAfterTenantLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.afterTenantDel
}

// GetBeforeInstallModeChangeHandler returns the registered source-plugin install-mode callback.
func (p *sourcePlugin) GetBeforeInstallModeChangeHandler() SourcePluginBeforeInstallModeChangeHandler {
	if p == nil {
		return nil
	}
	return p.beforeModeChange
}

// GetAfterInstallModeChangeHandler returns the registered source-plugin post-install-mode callback.
func (p *sourcePlugin) GetAfterInstallModeChangeHandler() SourcePluginAfterInstallModeChangeHandler {
	if p == nil {
		return nil
	}
	return p.afterModeChange
}

// GetUninstallHandler returns the registered source-plugin uninstall cleanup callback.
func (p *sourcePlugin) GetUninstallHandler() SourcePluginUninstallHandler {
	if p == nil {
		return nil
	}
	return p.uninstallHandler
}

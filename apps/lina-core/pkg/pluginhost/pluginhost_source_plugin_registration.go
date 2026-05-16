// This file implements source-plugin callback registration validation and
// storage behind the published grouped facades.

package pluginhost

import "github.com/gogf/gf/v2/errors/gerror"

// RegisterBeforeInstallHandler registers one source-plugin pre-install callback.
func (p *sourcePlugin) registerBeforeInstallHandler(handler SourcePluginBeforeLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: before install handler is nil")
	}
	p.beforeInstall = handler
	return nil
}

// RegisterAfterInstallHandler registers one source-plugin post-install callback.
func (p *sourcePlugin) registerAfterInstallHandler(handler SourcePluginAfterLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: after install handler is nil")
	}
	p.afterInstall = handler
	return nil
}

// RegisterBeforeUpgradeHandler registers one source-plugin pre-upgrade callback.
func (p *sourcePlugin) registerBeforeUpgradeHandler(handler SourcePluginBeforeUpgradeHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: before upgrade handler is nil")
	}
	p.beforeUpgrade = handler
	return nil
}

// RegisterUpgradeHandler registers one source-plugin custom upgrade callback.
func (p *sourcePlugin) registerUpgradeHandler(handler SourcePluginUpgradeHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: upgrade handler is nil")
	}
	p.upgradeHandler = handler
	return nil
}

// RegisterAfterUpgradeHandler registers one source-plugin post-upgrade callback.
func (p *sourcePlugin) registerAfterUpgradeHandler(handler SourcePluginUpgradeHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: after upgrade handler is nil")
	}
	p.afterUpgrade = handler
	return nil
}

// RegisterBeforeDisableHandler registers one source-plugin pre-disable callback.
func (p *sourcePlugin) registerBeforeDisableHandler(handler SourcePluginBeforeLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: before disable handler is nil")
	}
	p.beforeDisable = handler
	return nil
}

// RegisterAfterDisableHandler registers one source-plugin post-disable callback.
func (p *sourcePlugin) registerAfterDisableHandler(handler SourcePluginAfterLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: after disable handler is nil")
	}
	p.afterDisable = handler
	return nil
}

// RegisterBeforeUninstallHandler registers one source-plugin pre-uninstall callback.
func (p *sourcePlugin) registerBeforeUninstallHandler(handler SourcePluginBeforeLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: before uninstall handler is nil")
	}
	p.beforeUninstall = handler
	return nil
}

// RegisterAfterUninstallHandler registers one source-plugin post-uninstall callback.
func (p *sourcePlugin) registerAfterUninstallHandler(handler SourcePluginAfterLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: after uninstall handler is nil")
	}
	p.afterUninstall = handler
	return nil
}

// RegisterBeforeTenantDisableHandler registers one source-plugin tenant-disable precondition callback.
func (p *sourcePlugin) registerBeforeTenantDisableHandler(handler SourcePluginBeforeTenantLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: before tenant disable handler is nil")
	}
	p.beforeTenantDis = handler
	return nil
}

// RegisterAfterTenantDisableHandler registers one source-plugin tenant-disable post callback.
func (p *sourcePlugin) registerAfterTenantDisableHandler(handler SourcePluginAfterTenantLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: after tenant disable handler is nil")
	}
	p.afterTenantDis = handler
	return nil
}

// RegisterBeforeTenantDeleteHandler registers one source-plugin tenant-delete precondition callback.
func (p *sourcePlugin) registerBeforeTenantDeleteHandler(handler SourcePluginBeforeTenantLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: before tenant delete handler is nil")
	}
	p.beforeTenantDel = handler
	return nil
}

// RegisterAfterTenantDeleteHandler registers one source-plugin tenant-delete post callback.
func (p *sourcePlugin) registerAfterTenantDeleteHandler(handler SourcePluginAfterTenantLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: after tenant delete handler is nil")
	}
	p.afterTenantDel = handler
	return nil
}

// RegisterBeforeInstallModeChangeHandler registers one source-plugin install-mode precondition callback.
func (p *sourcePlugin) registerBeforeInstallModeChangeHandler(handler SourcePluginBeforeInstallModeChangeHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: before install mode change handler is nil")
	}
	p.beforeModeChange = handler
	return nil
}

// RegisterAfterInstallModeChangeHandler registers one source-plugin install-mode post callback.
func (p *sourcePlugin) registerAfterInstallModeChangeHandler(handler SourcePluginAfterInstallModeChangeHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: after install mode change handler is nil")
	}
	p.afterModeChange = handler
	return nil
}

// RegisterUninstallHandler registers one source-plugin uninstall cleanup callback.
func (p *sourcePlugin) registerUninstallHandler(handler SourcePluginUninstallHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: uninstall handler is nil")
	}
	p.uninstallHandler = handler
	return nil
}

// RegisterHook registers one callback-style host hook handler.
func (p *sourcePlugin) registerHook(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler HookHandler,
) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if !IsHookExtensionPoint(point) {
		return gerror.Newf("pluginhost: unpublished hook extension point: %s", point.String())
	}
	if handler == nil {
		return gerror.New("pluginhost: hook handler is nil")
	}
	normalizedMode, err := normalizeCallbackExecutionMode(point, mode)
	if err != nil {
		return err
	}
	// Store the normalized registration so the host can execute callbacks without
	// repeatedly re-validating plugin declarations at dispatch time.
	p.hookHandlers = append(p.hookHandlers, &HookHandlerRegistration{
		Mode:    normalizedMode,
		Point:   point,
		Handler: handler,
	})
	return nil
}

// RegisterRoutes registers one callback that contributes plugin-owned HTTP routes.
func (p *sourcePlugin) registerRoutes(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler RouteRegisterHandler,
) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: route registrar is nil")
	}
	normalizedMode, err := normalizeRegistrationPointMode(point, ExtensionPointHTTPRouteRegister, mode)
	if err != nil {
		return err
	}
	p.routeRegistrars = append(p.routeRegistrars, &RouteHandlerRegistration{
		Handler: handler,
		Mode:    normalizedMode,
		Point:   point,
	})
	return nil
}

// RegisterCron registers one callback that contributes plugin-owned cron jobs.
func (p *sourcePlugin) registerCron(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler CronRegisterHandler,
) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: cron registrar is nil")
	}
	normalizedMode, err := normalizeRegistrationPointMode(point, ExtensionPointCronRegister, mode)
	if err != nil {
		return err
	}
	p.cronRegistrars = append(p.cronRegistrars, &CronHandlerRegistration{
		Handler: handler,
		Mode:    normalizedMode,
		Point:   point,
	})
	return nil
}

// RegisterMenuFilter registers one callback that filters host menus.
func (p *sourcePlugin) registerMenuFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler MenuFilterHandler,
) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: menu filter handler is nil")
	}
	normalizedMode, err := normalizeRegistrationPointMode(point, ExtensionPointMenuFilter, mode)
	if err != nil {
		return err
	}
	p.menuFilters = append(p.menuFilters, &MenuFilterHandlerRegistration{
		Handler: handler,
		Mode:    normalizedMode,
		Point:   point,
	})
	return nil
}

// RegisterPermissionFilter registers one callback that filters host permissions.
func (p *sourcePlugin) registerPermissionFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler PermissionFilterHandler,
) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin is nil")
	}
	if handler == nil {
		return gerror.New("pluginhost: permission filter handler is nil")
	}
	normalizedMode, err := normalizeRegistrationPointMode(point, ExtensionPointPermissionFilter, mode)
	if err != nil {
		return err
	}
	p.permissionFilters = append(p.permissionFilters, &PermissionFilterHandlerRegistration{
		Handler: handler,
		Mode:    normalizedMode,
		Point:   point,
	})
	return nil
}

// normalizeCallbackExecutionMode validates one callback mode against the
// published pluginhost contract for the given extension point.
func normalizeCallbackExecutionMode(
	point ExtensionPoint,
	mode CallbackExecutionMode,
) (CallbackExecutionMode, error) {
	if mode == "" {
		mode = DefaultCallbackExecutionMode(point)
	}
	if !IsPublishedCallbackExecutionMode(mode) {
		return "", gerror.Newf("pluginhost: unsupported callback execution mode: %s", mode.String())
	}
	if !IsExtensionPointExecutionModeSupported(point, mode) {
		return "", gerror.Newf("pluginhost: callback execution mode is not supported by extension point: %s", point.String())
	}
	return mode, nil
}

// normalizeRegistrationPointMode validates a registration callback mode and
// ensures the handler is registered against the expected extension point.
func normalizeRegistrationPointMode(
	point ExtensionPoint,
	expected ExtensionPoint,
	mode CallbackExecutionMode,
) (CallbackExecutionMode, error) {
	if !IsRegistrationExtensionPoint(point) {
		return "", gerror.Newf("pluginhost: unpublished registration extension point: %s", point.String())
	}
	if point != expected {
		return "", gerror.Newf("pluginhost: unexpected registration extension point: %s", point.String())
	}
	return normalizeCallbackExecutionMode(point, mode)
}

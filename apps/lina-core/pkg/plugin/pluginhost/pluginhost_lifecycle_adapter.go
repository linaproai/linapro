// This file adapts source-plugin lifecycle facade callbacks into the shared
// lifecycle precondition callback runner.

package pluginhost

import "context"

// NewSourcePluginLifecycleCallbackAdapter returns hook callbacks registered
// through Declarations.Lifecycle().
func NewSourcePluginLifecycleCallbackAdapter(plugin SourcePluginDefinition) LifecycleCallbacks {
	if plugin == nil {
		return LifecycleCallbacks{}
	}

	var callbacks LifecycleCallbacks
	if handler := plugin.GetBeforeInstallHandler(); handler != nil {
		callbacks.set(LifecycleHookBeforeInstall, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return handler(ctx, req.PluginInput)
		})
	}
	if handler := plugin.GetAfterInstallHandler(); handler != nil {
		callbacks.set(LifecycleHookAfterInstall, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return true, "", handler(ctx, req.PluginInput)
		})
	}
	if handler := plugin.GetBeforeEnableHandler(); handler != nil {
		callbacks.set(LifecycleHookBeforeEnable, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return handler(ctx, req.PluginInput)
		})
	}
	if handler := plugin.GetAfterEnableHandler(); handler != nil {
		callbacks.set(LifecycleHookAfterEnable, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return true, "", handler(ctx, req.PluginInput)
		})
	}
	if handler := plugin.GetBeforeUpgradeHandler(); handler != nil {
		callbacks.set(LifecycleHookBeforeUpgrade, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return handler(ctx, req.UpgradeInput)
		})
	}
	if handler := plugin.GetUpgradeHandler(); handler != nil {
		callbacks.set(LifecycleHookUpgrade, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return true, "", handler(ctx, req.UpgradeInput)
		})
	}
	if handler := plugin.GetAfterUpgradeHandler(); handler != nil {
		callbacks.set(LifecycleHookAfterUpgrade, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return true, "", handler(ctx, req.UpgradeInput)
		})
	}
	if handler := plugin.GetBeforeDisableHandler(); handler != nil {
		callbacks.set(LifecycleHookBeforeDisable, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return handler(ctx, req.PluginInput)
		})
	}
	if handler := plugin.GetAfterDisableHandler(); handler != nil {
		callbacks.set(LifecycleHookAfterDisable, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return true, "", handler(ctx, req.PluginInput)
		})
	}
	if handler := plugin.GetBeforeUninstallHandler(); handler != nil {
		callbacks.set(LifecycleHookBeforeUninstall, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return handler(ctx, req.PluginInput)
		})
	}
	if handler := plugin.GetUninstallHandler(); handler != nil {
		callbacks.set(LifecycleHookUninstall, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return true, "", handler(ctx, req.UninstallInput)
		})
	}
	if handler := plugin.GetAfterUninstallHandler(); handler != nil {
		callbacks.set(LifecycleHookAfterUninstall, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return true, "", handler(ctx, req.PluginInput)
		})
	}
	if handler := plugin.GetBeforeTenantDisableHandler(); handler != nil {
		callbacks.set(LifecycleHookBeforeTenantDisable, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return handler(ctx, req.TenantInput)
		})
	}
	if handler := plugin.GetAfterTenantDisableHandler(); handler != nil {
		callbacks.set(LifecycleHookAfterTenantDisable, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return true, "", handler(ctx, req.TenantInput)
		})
	}
	if handler := plugin.GetBeforeTenantDeleteHandler(); handler != nil {
		callbacks.set(LifecycleHookBeforeTenantDelete, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return handler(ctx, req.TenantInput)
		})
	}
	if handler := plugin.GetAfterTenantDeleteHandler(); handler != nil {
		callbacks.set(LifecycleHookAfterTenantDelete, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
			return true, "", handler(ctx, req.TenantInput)
		})
	}
	if handler := plugin.GetBeforeInstallModeChangeHandler(); handler != nil {
		callbacks.set(LifecycleHookBeforeInstallModeChange, func(
			ctx context.Context,
			req LifecycleRequest,
		) (bool, string, error) {
			return handler(ctx, req.InstallModeInput)
		})
	}
	if handler := plugin.GetAfterInstallModeChangeHandler(); handler != nil {
		callbacks.set(LifecycleHookAfterInstallModeChange, func(
			ctx context.Context,
			req LifecycleRequest,
		) (bool, string, error) {
			return true, "", handler(ctx, req.InstallModeInput)
		})
	}
	return callbacks
}

// NewSourcePluginGlobalLifecycleCallbackAdapter returns callbacks for one
// explicitly requested global Before* hook. Plugins without that registration
// yield an empty set.
func NewSourcePluginGlobalLifecycleCallbackAdapter(
	plugin SourcePluginDefinition,
	hook LifecycleHook,
) LifecycleCallbacks {
	if plugin == nil || !IsGlobalLifecycleHook(hook) {
		return LifecycleCallbacks{}
	}
	var handler SourcePluginGlobalLifecycleHandler
	switch hook {
	case LifecycleHookGlobalBeforeInstall:
		handler = plugin.GetGlobalBeforeInstallHandler()
	case LifecycleHookGlobalBeforeEnable:
		handler = plugin.GetGlobalBeforeEnableHandler()
	case LifecycleHookGlobalBeforeDisable:
		handler = plugin.GetGlobalBeforeDisableHandler()
	case LifecycleHookGlobalBeforeUninstall:
		handler = plugin.GetGlobalBeforeUninstallHandler()
	default:
		return LifecycleCallbacks{}
	}
	if handler == nil {
		return LifecycleCallbacks{}
	}
	var callbacks LifecycleCallbacks
	callbacks.set(hook, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
		return handler(ctx, req.GlobalInput)
	})
	return callbacks
}

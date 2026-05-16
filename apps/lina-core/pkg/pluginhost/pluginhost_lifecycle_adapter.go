// This file adapts source-plugin lifecycle facade callbacks into the shared
// lifecycle precondition callback runner.

package pluginhost

import "context"

// sourcePluginLifecycleCallbackAdapter exposes lifecycle facade callbacks
// through the shared callback participant interface.
type sourcePluginLifecycleCallbackAdapter struct {
	plugin SourcePluginDefinition
}

// NewSourcePluginLifecycleCallbackAdapter returns an adapter for callbacks
// registered through SourcePlugin.Lifecycle().
func NewSourcePluginLifecycleCallbackAdapter(plugin SourcePluginDefinition) any {
	if plugin == nil {
		return nil
	}
	return &sourcePluginLifecycleCallbackAdapter{plugin: plugin}
}

// BeforeInstall invokes the source-plugin pre-install callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) BeforeInstall(ctx context.Context, input SourcePluginLifecycleInput) (bool, string, error) {
	if a == nil || a.plugin == nil || a.plugin.GetBeforeInstallHandler() == nil {
		return true, "", nil
	}
	return a.plugin.GetBeforeInstallHandler()(ctx, input)
}

// AfterInstall invokes the source-plugin post-install callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) AfterInstall(ctx context.Context, input SourcePluginLifecycleInput) error {
	if a == nil || a.plugin == nil || a.plugin.GetAfterInstallHandler() == nil {
		return nil
	}
	return a.plugin.GetAfterInstallHandler()(ctx, input)
}

// BeforeUpgrade invokes the source-plugin pre-upgrade callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) BeforeUpgrade(ctx context.Context, input SourcePluginUpgradeInput) (bool, string, error) {
	if a == nil || a.plugin == nil || a.plugin.GetBeforeUpgradeHandler() == nil {
		return true, "", nil
	}
	return a.plugin.GetBeforeUpgradeHandler()(ctx, input)
}

// AfterUpgrade invokes the source-plugin post-upgrade callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) AfterUpgrade(ctx context.Context, input SourcePluginUpgradeInput) error {
	if a == nil || a.plugin == nil || a.plugin.GetAfterUpgradeHandler() == nil {
		return nil
	}
	return a.plugin.GetAfterUpgradeHandler()(ctx, input)
}

// Upgrade invokes the source-plugin custom upgrade callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) Upgrade(ctx context.Context, input SourcePluginUpgradeInput) error {
	if a == nil || a.plugin == nil || a.plugin.GetUpgradeHandler() == nil {
		return nil
	}
	return a.plugin.GetUpgradeHandler()(ctx, input)
}

// BeforeDisable invokes the source-plugin pre-disable callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) BeforeDisable(ctx context.Context, input SourcePluginLifecycleInput) (bool, string, error) {
	if a == nil || a.plugin == nil || a.plugin.GetBeforeDisableHandler() == nil {
		return true, "", nil
	}
	return a.plugin.GetBeforeDisableHandler()(ctx, input)
}

// AfterDisable invokes the source-plugin post-disable callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) AfterDisable(ctx context.Context, input SourcePluginLifecycleInput) error {
	if a == nil || a.plugin == nil || a.plugin.GetAfterDisableHandler() == nil {
		return nil
	}
	return a.plugin.GetAfterDisableHandler()(ctx, input)
}

// BeforeUninstall invokes the source-plugin pre-uninstall callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) BeforeUninstall(ctx context.Context, input SourcePluginLifecycleInput) (bool, string, error) {
	if a == nil || a.plugin == nil || a.plugin.GetBeforeUninstallHandler() == nil {
		return true, "", nil
	}
	return a.plugin.GetBeforeUninstallHandler()(ctx, input)
}

// AfterUninstall invokes the source-plugin post-uninstall callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) AfterUninstall(ctx context.Context, input SourcePluginLifecycleInput) error {
	if a == nil || a.plugin == nil || a.plugin.GetAfterUninstallHandler() == nil {
		return nil
	}
	return a.plugin.GetAfterUninstallHandler()(ctx, input)
}

// Uninstall invokes the source-plugin uninstall cleanup callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) Uninstall(ctx context.Context, input SourcePluginUninstallInput) error {
	if a == nil || a.plugin == nil || a.plugin.GetUninstallHandler() == nil {
		return nil
	}
	return a.plugin.GetUninstallHandler()(ctx, input)
}

// BeforeTenantDisable invokes the source-plugin tenant-disable callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) BeforeTenantDisable(
	ctx context.Context,
	input SourcePluginTenantLifecycleInput,
) (bool, string, error) {
	if a == nil || a.plugin == nil || a.plugin.GetBeforeTenantDisableHandler() == nil {
		return true, "", nil
	}
	return a.plugin.GetBeforeTenantDisableHandler()(ctx, input)
}

// AfterTenantDisable invokes the source-plugin tenant-disable post callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) AfterTenantDisable(
	ctx context.Context,
	input SourcePluginTenantLifecycleInput,
) error {
	if a == nil || a.plugin == nil || a.plugin.GetAfterTenantDisableHandler() == nil {
		return nil
	}
	return a.plugin.GetAfterTenantDisableHandler()(ctx, input)
}

// BeforeTenantDelete invokes the source-plugin tenant-delete callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) BeforeTenantDelete(
	ctx context.Context,
	input SourcePluginTenantLifecycleInput,
) (bool, string, error) {
	if a == nil || a.plugin == nil || a.plugin.GetBeforeTenantDeleteHandler() == nil {
		return true, "", nil
	}
	return a.plugin.GetBeforeTenantDeleteHandler()(ctx, input)
}

// AfterTenantDelete invokes the source-plugin tenant-delete post callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) AfterTenantDelete(
	ctx context.Context,
	input SourcePluginTenantLifecycleInput,
) error {
	if a == nil || a.plugin == nil || a.plugin.GetAfterTenantDeleteHandler() == nil {
		return nil
	}
	return a.plugin.GetAfterTenantDeleteHandler()(ctx, input)
}

// BeforeInstallModeChange invokes the source-plugin install-mode callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) BeforeInstallModeChange(
	ctx context.Context,
	input SourcePluginInstallModeChangeInput,
) (bool, string, error) {
	if a == nil || a.plugin == nil || a.plugin.GetBeforeInstallModeChangeHandler() == nil {
		return true, "", nil
	}
	return a.plugin.GetBeforeInstallModeChangeHandler()(ctx, input)
}

// AfterInstallModeChange invokes the source-plugin install-mode post callback when present.
func (a *sourcePluginLifecycleCallbackAdapter) AfterInstallModeChange(
	ctx context.Context,
	input SourcePluginInstallModeChangeInput,
) error {
	if a == nil || a.plugin == nil || a.plugin.GetAfterInstallModeChangeHandler() == nil {
		return nil
	}
	return a.plugin.GetAfterInstallModeChangeHandler()(ctx, input)
}

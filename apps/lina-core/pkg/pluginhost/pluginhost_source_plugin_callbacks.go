// This file defines source-plugin callback handler types and registration
// snapshots stored by the grouped source-plugin facades.

package pluginhost

import "context"

// HookHandler defines one callback-style hook handler.
type HookHandler func(ctx context.Context, payload HookPayload) error

// HookHandlerRegistration defines one hook subscription registered by a source plugin.
type HookHandlerRegistration struct {
	// Handler is the callback invoked by the host.
	Handler HookHandler
	// Mode is the declared callback execution mode.
	Mode CallbackExecutionMode
	// Point is the published backend extension point.
	Point ExtensionPoint
}

// RouteHandlerRegistration defines one route-registration callback subscribed by a source plugin.
type RouteHandlerRegistration struct {
	// Handler is the callback invoked by the host startup registrar.
	Handler RouteRegisterHandler
	// Mode is the declared callback execution mode.
	Mode CallbackExecutionMode
	// Point is the published backend extension point.
	Point ExtensionPoint
}

// CronHandlerRegistration defines one cron-registration callback subscribed by a source plugin.
type CronHandlerRegistration struct {
	// Handler is the callback invoked by the host cron registrar.
	Handler CronRegisterHandler
	// Mode is the declared callback execution mode.
	Mode CallbackExecutionMode
	// Point is the published backend extension point.
	Point ExtensionPoint
}

// MenuFilterHandlerRegistration defines one menu-filter callback subscribed by a source plugin.
type MenuFilterHandlerRegistration struct {
	// Handler is the callback invoked by the host.
	Handler MenuFilterHandler
	// Mode is the declared callback execution mode.
	Mode CallbackExecutionMode
	// Point is the published backend extension point.
	Point ExtensionPoint
}

// PermissionFilterHandlerRegistration defines one permission-filter callback subscribed by a source plugin.
type PermissionFilterHandlerRegistration struct {
	// Handler is the callback invoked by the host.
	Handler PermissionFilterHandler
	// Mode is the declared callback execution mode.
	Mode CallbackExecutionMode
	// Point is the published backend extension point.
	Point ExtensionPoint
}

// SourcePluginBeforeLifecycleHandler defines one callback that may veto a
// generic source-plugin lifecycle operation.
type SourcePluginBeforeLifecycleHandler func(ctx context.Context, input SourcePluginLifecycleInput) (ok bool, reason string, err error)

// SourcePluginAfterLifecycleHandler defines one non-blocking notification
// callback invoked after a generic source-plugin lifecycle operation succeeds.
type SourcePluginAfterLifecycleHandler func(ctx context.Context, input SourcePluginLifecycleInput) error

// SourcePluginBeforeUpgradeHandler defines one callback that may veto a
// source-plugin runtime upgrade before side effects run.
type SourcePluginBeforeUpgradeHandler func(ctx context.Context, input SourcePluginUpgradeInput) (ok bool, reason string, err error)

// SourcePluginBeforeTenantLifecycleHandler defines one callback that may veto a
// tenant lifecycle operation.
type SourcePluginBeforeTenantLifecycleHandler func(
	ctx context.Context,
	input SourcePluginTenantLifecycleInput,
) (ok bool, reason string, err error)

// SourcePluginAfterTenantLifecycleHandler defines one non-blocking notification
// callback invoked after a tenant lifecycle operation succeeds.
type SourcePluginAfterTenantLifecycleHandler func(
	ctx context.Context,
	input SourcePluginTenantLifecycleInput,
) error

// SourcePluginBeforeInstallModeChangeHandler defines one callback that may veto
// a source-plugin install-mode transition.
type SourcePluginBeforeInstallModeChangeHandler func(
	ctx context.Context,
	input SourcePluginInstallModeChangeInput,
) (ok bool, reason string, err error)

// SourcePluginAfterInstallModeChangeHandler defines one non-blocking
// notification callback invoked after an install-mode transition succeeds.
type SourcePluginAfterInstallModeChangeHandler func(
	ctx context.Context,
	input SourcePluginInstallModeChangeInput,
) error

// SourcePluginUpgradeHandler defines one callback invoked during or after a
// source-plugin runtime upgrade.
type SourcePluginUpgradeHandler func(ctx context.Context, input SourcePluginUpgradeInput) error

// SourcePluginUninstallHandler defines one callback invoked before the host executes source-plugin uninstall SQL.
type SourcePluginUninstallHandler func(ctx context.Context, input SourcePluginUninstallInput) error

// RouteRegisterHandler defines one callback that registers plugin-owned HTTP routes
// and global middleware through the published HTTP registrar.
type RouteRegisterHandler func(ctx context.Context, registrar HTTPRegistrar) error

// CronRegisterHandler defines one callback that registers plugin-owned cron jobs.
type CronRegisterHandler func(ctx context.Context, registrar CronRegistrar) error

// MenuFilterHandler defines one callback that decides whether a menu should stay visible.
type MenuFilterHandler func(ctx context.Context, menu MenuDescriptor) (bool, error)

// PermissionFilterHandler defines one callback that decides whether a permission should stay effective.
type PermissionFilterHandler func(ctx context.Context, permission PermissionDescriptor) (bool, error)

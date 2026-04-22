// This file defines the published source-plugin interfaces and grouped
// registration facades exposed to source-plugin authors.

package pluginhost

import "io/fs"

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
	// RegisterUninstallHandler registers one uninstall cleanup callback.
	RegisterUninstallHandler(handler SourcePluginUninstallHandler)
}

// SourcePluginHooks exposes callback-style host hook registrations grouped
// under one dedicated facade.
type SourcePluginHooks interface {
	// RegisterHook registers one callback-style host hook handler.
	RegisterHook(point ExtensionPoint, mode CallbackExecutionMode, handler HookHandler)
}

// SourcePluginHTTP exposes HTTP-adjacent registrations grouped under one
// dedicated facade.
type SourcePluginHTTP interface {
	// RegisterRoutes registers one callback that contributes plugin-owned HTTP routes.
	RegisterRoutes(point ExtensionPoint, mode CallbackExecutionMode, handler RouteRegisterHandler)
}

// SourcePluginCron exposes cron registrations grouped under one dedicated
// facade.
type SourcePluginCron interface {
	// RegisterCron registers one callback that contributes plugin-owned cron jobs.
	RegisterCron(point ExtensionPoint, mode CallbackExecutionMode, handler CronRegisterHandler)
}

// SourcePluginGovernance exposes governance callback registrations grouped
// under one dedicated facade.
type SourcePluginGovernance interface {
	// RegisterMenuFilter registers one callback that filters host menus.
	RegisterMenuFilter(point ExtensionPoint, mode CallbackExecutionMode, handler MenuFilterHandler)
	// RegisterPermissionFilter registers one callback that filters host permissions.
	RegisterPermissionFilter(point ExtensionPoint, mode CallbackExecutionMode, handler PermissionFilterHandler)
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
func (r *sourcePluginLifecycle) RegisterUninstallHandler(handler SourcePluginUninstallHandler) {
	if r == nil || r.plugin == nil {
		panic("pluginhost: source plugin lifecycle facade is nil")
	}
	r.plugin.registerUninstallHandler(handler)
}

// RegisterHook registers one callback-style host hook handler.
func (r *sourcePluginHooks) RegisterHook(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler HookHandler,
) {
	if r == nil || r.plugin == nil {
		panic("pluginhost: source plugin hook facade is nil")
	}
	r.plugin.registerHook(point, mode, handler)
}

// RegisterRoutes registers one callback that contributes plugin-owned HTTP routes.
func (r *sourcePluginHTTP) RegisterRoutes(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler RouteRegisterHandler,
) {
	if r == nil || r.plugin == nil {
		panic("pluginhost: source plugin http facade is nil")
	}
	r.plugin.registerRoutes(point, mode, handler)
}

// RegisterCron registers one callback that contributes plugin-owned cron jobs.
func (r *sourcePluginCron) RegisterCron(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler CronRegisterHandler,
) {
	if r == nil || r.plugin == nil {
		panic("pluginhost: source plugin cron facade is nil")
	}
	r.plugin.registerCron(point, mode, handler)
}

// RegisterMenuFilter registers one callback that filters host menus.
func (r *sourcePluginGovernance) RegisterMenuFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler MenuFilterHandler,
) {
	if r == nil || r.plugin == nil {
		panic("pluginhost: source plugin governance facade is nil")
	}
	r.plugin.registerMenuFilter(point, mode, handler)
}

// RegisterPermissionFilter registers one callback that filters host permissions.
func (r *sourcePluginGovernance) RegisterPermissionFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler PermissionFilterHandler,
) {
	if r == nil || r.plugin == nil {
		panic("pluginhost: source plugin governance facade is nil")
	}
	r.plugin.registerPermissionFilter(point, mode, handler)
}

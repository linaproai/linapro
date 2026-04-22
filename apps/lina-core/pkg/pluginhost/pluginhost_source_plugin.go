// This file defines the host-owned source-plugin registration storage plus the
// callback input wrappers that isolate plugins from host internals.

package pluginhost

import (
	"context"
	"io/fs"
)

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
	uninstallHandler  SourcePluginUninstallHandler
	hookHandlers      []*HookHandlerRegistration
	routeRegistrars   []*RouteHandlerRegistration
	cronRegistrars    []*CronHandlerRegistration
	menuFilters       []*MenuFilterHandlerRegistration
	permissionFilters []*PermissionFilterHandlerRegistration
}

// HookPayload exposes one published host hook payload.
type HookPayload interface {
	// ExtensionPoint returns the published extension point of the current callback.
	ExtensionPoint() ExtensionPoint
	// Value returns one published payload field by key.
	Value(key string) interface{}
	// Values returns a copy of all published payload fields.
	Values() map[string]interface{}
}

// hookPayload is the host-owned implementation of the published HookPayload view.
type hookPayload struct {
	point  ExtensionPoint
	values map[string]interface{}
}

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

// SourcePluginUninstallInput exposes one host-confirmed uninstall policy snapshot to a source plugin.
type SourcePluginUninstallInput interface {
	// PluginID returns the source-plugin identifier being uninstalled.
	PluginID() string
	// PurgeStorageData reports whether the host expects the plugin to clear its
	// own business data and stored files during uninstall.
	PurgeStorageData() bool
}

// sourcePluginUninstallInput is the host-owned implementation of the published
// uninstall policy snapshot passed to source plugins.
type sourcePluginUninstallInput struct {
	pluginID         string
	purgeStorageData bool
}

// SourcePluginUninstallHandler defines one callback invoked before the host executes source-plugin uninstall SQL.
type SourcePluginUninstallHandler func(ctx context.Context, input SourcePluginUninstallInput) error

// RouteRegisterHandler defines one callback that registers plugin-owned HTTP routes
// and global middleware through the published HTTP registrar.
type RouteRegisterHandler func(ctx context.Context, registrar HTTPRegistrar) error

// CronRegisterHandler defines one callback that registers plugin-owned cron jobs.
type CronRegisterHandler func(ctx context.Context, registrar CronRegistrar) error

// MenuDescriptor exposes one published menu descriptor for plugin menu filtering.
type MenuDescriptor interface {
	// ID returns the menu id.
	ID() int
	// ParentID returns the parent menu id.
	ParentID() int
	// Name returns the menu display name.
	Name() string
	// Path returns the menu path.
	Path() string
	// Component returns the routed component name.
	Component() string
	// Permissions returns the permission key bound to the menu.
	Permissions() string
	// MenuKey returns the stable menu business key.
	MenuKey() string
	// Type returns the menu type.
	Type() string
	// Visible returns the visible status.
	Visible() int
	// Status returns the enabled status.
	Status() int
}

// menuDescriptor is the host-owned implementation of MenuDescriptor.
type menuDescriptor struct {
	id         int
	parentID   int
	name       string
	path       string
	component  string
	permission string
	menuKey    string
	menuType   string
	visible    int
	status     int
}

// MenuFilterHandler defines one callback that decides whether a menu should stay visible.
type MenuFilterHandler func(ctx context.Context, menu MenuDescriptor) (bool, error)

// PermissionDescriptor exposes one published permission descriptor for plugin permission filtering.
type PermissionDescriptor interface {
	// MenuKey returns the stable menu business key.
	MenuKey() string
	// MenuName returns the display name of the menu that owns the permission.
	MenuName() string
	// Permission returns the permission string.
	Permission() string
}

// permissionDescriptor is the host-owned implementation of PermissionDescriptor.
type permissionDescriptor struct {
	menuKey    string
	menuName   string
	permission string
}

// PermissionFilterHandler defines one callback that decides whether a permission should stay effective.
type PermissionFilterHandler func(ctx context.Context, permission PermissionDescriptor) (bool, error)

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

// NewHookPayload creates one published hook payload wrapper for plugins.
func NewHookPayload(point ExtensionPoint, values map[string]interface{}) HookPayload {
	return &hookPayload{
		point:  point,
		values: cloneValueMap(values),
	}
}

// NewMenuDescriptor creates one published menu descriptor wrapper for plugins.
func NewMenuDescriptor(
	id int,
	parentID int,
	name string,
	path string,
	component string,
	permission string,
	menuKey string,
	menuType string,
	visible int,
	status int,
) MenuDescriptor {
	return &menuDescriptor{
		id:         id,
		parentID:   parentID,
		name:       name,
		path:       path,
		component:  component,
		permission: permission,
		menuKey:    menuKey,
		menuType:   menuType,
		visible:    visible,
		status:     status,
	}
}

// NewPermissionDescriptor creates one published permission descriptor wrapper for plugins.
func NewPermissionDescriptor(menuKey string, menuName string, permission string) PermissionDescriptor {
	return &permissionDescriptor{
		menuKey:    menuKey,
		menuName:   menuName,
		permission: permission,
	}
}

// NewSourcePluginUninstallInput creates one published source-plugin uninstall input wrapper.
func NewSourcePluginUninstallInput(
	pluginID string,
	purgeStorageData bool,
) SourcePluginUninstallInput {
	return &sourcePluginUninstallInput{
		pluginID:         pluginID,
		purgeStorageData: purgeStorageData,
	}
}

// RegisterUninstallHandler registers one source-plugin uninstall cleanup callback.
func (p *sourcePlugin) registerUninstallHandler(handler SourcePluginUninstallHandler) {
	if p == nil {
		panic("pluginhost: source plugin is nil")
	}
	if handler == nil {
		panic("pluginhost: uninstall handler is nil")
	}
	p.uninstallHandler = handler
}

// RegisterHook registers one callback-style host hook handler.
func (p *sourcePlugin) registerHook(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler HookHandler,
) {
	if p == nil {
		panic("pluginhost: source plugin is nil")
	}
	if !IsHookExtensionPoint(point) {
		panic("pluginhost: unpublished hook extension point: " + point.String())
	}
	if handler == nil {
		panic("pluginhost: hook handler is nil")
	}
	mode = normalizeCallbackExecutionMode(point, mode)
	// Store the normalized registration so the host can execute callbacks without
	// repeatedly re-validating plugin declarations at dispatch time.
	p.hookHandlers = append(p.hookHandlers, &HookHandlerRegistration{
		Mode:    mode,
		Point:   point,
		Handler: handler,
	})
}

// RegisterRoutes registers one callback that contributes plugin-owned HTTP routes.
func (p *sourcePlugin) registerRoutes(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler RouteRegisterHandler,
) {
	if p == nil {
		panic("pluginhost: source plugin is nil")
	}
	if handler == nil {
		panic("pluginhost: route registrar is nil")
	}
	mode = normalizeRegistrationPointMode(point, ExtensionPointHTTPRouteRegister, mode)
	p.routeRegistrars = append(p.routeRegistrars, &RouteHandlerRegistration{
		Handler: handler,
		Mode:    mode,
		Point:   point,
	})
}

// RegisterCron registers one callback that contributes plugin-owned cron jobs.
func (p *sourcePlugin) registerCron(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler CronRegisterHandler,
) {
	if p == nil {
		panic("pluginhost: source plugin is nil")
	}
	if handler == nil {
		panic("pluginhost: cron registrar is nil")
	}
	mode = normalizeRegistrationPointMode(point, ExtensionPointCronRegister, mode)
	p.cronRegistrars = append(p.cronRegistrars, &CronHandlerRegistration{
		Handler: handler,
		Mode:    mode,
		Point:   point,
	})
}

// RegisterMenuFilter registers one callback that filters host menus.
func (p *sourcePlugin) registerMenuFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler MenuFilterHandler,
) {
	if p == nil {
		panic("pluginhost: source plugin is nil")
	}
	if handler == nil {
		panic("pluginhost: menu filter handler is nil")
	}
	mode = normalizeRegistrationPointMode(point, ExtensionPointMenuFilter, mode)
	p.menuFilters = append(p.menuFilters, &MenuFilterHandlerRegistration{
		Handler: handler,
		Mode:    mode,
		Point:   point,
	})
}

// RegisterPermissionFilter registers one callback that filters host permissions.
func (p *sourcePlugin) registerPermissionFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler PermissionFilterHandler,
) {
	if p == nil {
		panic("pluginhost: source plugin is nil")
	}
	if handler == nil {
		panic("pluginhost: permission filter handler is nil")
	}
	mode = normalizeRegistrationPointMode(point, ExtensionPointPermissionFilter, mode)
	p.permissionFilters = append(p.permissionFilters, &PermissionFilterHandlerRegistration{
		Handler: handler,
		Mode:    mode,
		Point:   point,
	})
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

// GetUninstallHandler returns the registered source-plugin uninstall cleanup callback.
func (p *sourcePlugin) GetUninstallHandler() SourcePluginUninstallHandler {
	if p == nil {
		return nil
	}
	return p.uninstallHandler
}

// ExtensionPoint returns the published extension point of the current hook payload.
func (p *hookPayload) ExtensionPoint() ExtensionPoint {
	if p == nil {
		return ""
	}
	return p.point
}

// Value returns one published payload field by key.
func (p *hookPayload) Value(key string) interface{} {
	if p == nil {
		return nil
	}
	return p.values[key]
}

// Values returns a shallow copy of all published payload fields.
func (p *hookPayload) Values() map[string]interface{} {
	if p == nil {
		return map[string]interface{}{}
	}
	return cloneValueMap(p.values)
}

// PluginID returns the source-plugin identifier being uninstalled.
func (i *sourcePluginUninstallInput) PluginID() string {
	if i == nil {
		return ""
	}
	return i.pluginID
}

// PurgeStorageData reports whether the host expects business data cleanup.
func (i *sourcePluginUninstallInput) PurgeStorageData() bool {
	if i == nil {
		return false
	}
	return i.purgeStorageData
}

// ID returns the menu identifier.
func (d *menuDescriptor) ID() int {
	if d == nil {
		return 0
	}
	return d.id
}

// ParentID returns the parent menu identifier.
func (d *menuDescriptor) ParentID() int {
	if d == nil {
		return 0
	}
	return d.parentID
}

// Name returns the menu display name.
func (d *menuDescriptor) Name() string {
	if d == nil {
		return ""
	}
	return d.name
}

// Path returns the menu route path.
func (d *menuDescriptor) Path() string {
	if d == nil {
		return ""
	}
	return d.path
}

// Component returns the menu component binding.
func (d *menuDescriptor) Component() string {
	if d == nil {
		return ""
	}
	return d.component
}

// Permissions returns the menu permission string.
func (d *menuDescriptor) Permissions() string {
	if d == nil {
		return ""
	}
	return d.permission
}

// MenuKey returns the stable menu business key.
func (d *menuDescriptor) MenuKey() string {
	if d == nil {
		return ""
	}
	return d.menuKey
}

// Type returns the menu type code.
func (d *menuDescriptor) Type() string {
	if d == nil {
		return ""
	}
	return d.menuType
}

// Visible returns the menu visibility status.
func (d *menuDescriptor) Visible() int {
	if d == nil {
		return 0
	}
	return d.visible
}

// Status returns the menu enabled status.
func (d *menuDescriptor) Status() int {
	if d == nil {
		return 0
	}
	return d.status
}

// MenuKey returns the stable business key of the menu owning the permission.
func (d *permissionDescriptor) MenuKey() string {
	if d == nil {
		return ""
	}
	return d.menuKey
}

// MenuName returns the display name of the menu owning the permission.
func (d *permissionDescriptor) MenuName() string {
	if d == nil {
		return ""
	}
	return d.menuName
}

// Permission returns the concrete permission string.
func (d *permissionDescriptor) Permission() string {
	if d == nil {
		return ""
	}
	return d.permission
}

// normalizeCallbackExecutionMode validates one callback mode against the
// published pluginhost contract for the given extension point.
func normalizeCallbackExecutionMode(
	point ExtensionPoint,
	mode CallbackExecutionMode,
) CallbackExecutionMode {
	if mode == "" {
		mode = DefaultCallbackExecutionMode(point)
	}
	if !IsPublishedCallbackExecutionMode(mode) {
		panic("pluginhost: unsupported callback execution mode: " + mode.String())
	}
	if !IsExtensionPointExecutionModeSupported(point, mode) {
		panic("pluginhost: callback execution mode is not supported by extension point: " + point.String())
	}
	return mode
}

// normalizeRegistrationPointMode validates a registration callback mode and
// ensures the handler is registered against the expected extension point.
func normalizeRegistrationPointMode(
	point ExtensionPoint,
	expected ExtensionPoint,
	mode CallbackExecutionMode,
) CallbackExecutionMode {
	if !IsRegistrationExtensionPoint(point) {
		panic("pluginhost: unpublished registration extension point: " + point.String())
	}
	if point != expected {
		panic("pluginhost: unexpected registration extension point: " + point.String())
	}
	return normalizeCallbackExecutionMode(point, mode)
}

// cloneValueMap returns a shallow copy of the given payload value map.
func cloneValueMap(values map[string]interface{}) map[string]interface{} {
	if len(values) == 0 {
		return map[string]interface{}{}
	}
	cloned := make(map[string]interface{}, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

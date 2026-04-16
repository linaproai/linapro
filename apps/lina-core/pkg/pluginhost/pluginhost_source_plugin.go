// This file defines the public source-plugin contract, callback registration
// APIs, and wrapper interfaces that isolate plugins from host internals.

package pluginhost

import (
	"context"
	"io/fs"

	"github.com/gogf/gf/v2/net/ghttp"
)

// SourcePlugin defines one compile-time source plugin contribution.
type SourcePlugin struct {
	// ID is the stable plugin id and must match `plugin.yaml`.
	ID string

	embeddedFiles     fs.FS
	hookHandlers      []*HookHandlerRegistration
	routeRegistrars   []*RouteHandlerRegistration
	afterAuthHandlers []*AfterAuthHandlerRegistration
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

// AfterAuthHandlerRegistration defines one after-auth callback subscribed by a source plugin.
type AfterAuthHandlerRegistration struct {
	// Handler is the callback invoked by the host.
	Handler AfterAuthHandler
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

// AfterAuthInput exposes the published request context after host authentication succeeds.
type AfterAuthInput interface {
	// Request returns the current authenticated HTTP request.
	Request() *ghttp.Request
	// SetResponseHeader writes one response header to the current request.
	SetResponseHeader(key string, value string)
	// TokenID returns the current access token identifier.
	TokenID() string
	// UserID returns the authenticated user id.
	UserID() int
	// Username returns the authenticated username.
	Username() string
	// Status returns the authenticated user status.
	Status() int
}

type afterAuthInput struct {
	request  *ghttp.Request
	tokenID  string
	userID   int
	username string
	status   int
}

// AfterAuthHandler defines one callback invoked after host authentication succeeds.
type AfterAuthHandler func(ctx context.Context, input AfterAuthInput) error

// RouteRegisterHandler defines one callback that registers plugin-owned HTTP routes.
type RouteRegisterHandler func(ctx context.Context, registrar RouteRegistrar) error

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

type permissionDescriptor struct {
	menuKey    string
	menuName   string
	permission string
}

// PermissionFilterHandler defines one callback that decides whether a permission should stay effective.
type PermissionFilterHandler func(ctx context.Context, permission PermissionDescriptor) (bool, error)

// NewSourcePlugin creates and returns a new source plugin definition.
func NewSourcePlugin(id string) *SourcePlugin {
	return &SourcePlugin{
		ID:                id,
		hookHandlers:      make([]*HookHandlerRegistration, 0),
		routeRegistrars:   make([]*RouteHandlerRegistration, 0),
		afterAuthHandlers: make([]*AfterAuthHandlerRegistration, 0),
		cronRegistrars:    make([]*CronHandlerRegistration, 0),
		menuFilters:       make([]*MenuFilterHandlerRegistration, 0),
		permissionFilters: make([]*PermissionFilterHandlerRegistration, 0),
	}
}

// UseEmbeddedFiles binds one plugin-owned embedded filesystem to the source plugin.
func (p *SourcePlugin) UseEmbeddedFiles(fileSystem fs.FS) {
	if p == nil {
		return
	}
	p.embeddedFiles = fileSystem
}

// GetEmbeddedFiles returns the plugin-owned embedded filesystem when declared.
func (p *SourcePlugin) GetEmbeddedFiles() fs.FS {
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

// NewAfterAuthInput creates one published after-auth input wrapper for plugins.
func NewAfterAuthInput(
	request *ghttp.Request,
	tokenID string,
	userID int,
	username string,
	status int,
) AfterAuthInput {
	return &afterAuthInput{
		request:  request,
		tokenID:  tokenID,
		userID:   userID,
		username: username,
		status:   status,
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

// RegisterHook registers one callback-style host hook handler.
func (p *SourcePlugin) RegisterHook(
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
func (p *SourcePlugin) RegisterRoutes(
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

// RegisterAfterAuthHandler registers one callback invoked after host authentication succeeds.
func (p *SourcePlugin) RegisterAfterAuthHandler(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler AfterAuthHandler,
) {
	if p == nil {
		panic("pluginhost: source plugin is nil")
	}
	if handler == nil {
		panic("pluginhost: after-auth handler is nil")
	}
	mode = normalizeRegistrationPointMode(point, ExtensionPointHTTPRequestAfterAuth, mode)
	p.afterAuthHandlers = append(p.afterAuthHandlers, &AfterAuthHandlerRegistration{
		Handler: handler,
		Mode:    mode,
		Point:   point,
	})
}

// RegisterCron registers one callback that contributes plugin-owned cron jobs.
func (p *SourcePlugin) RegisterCron(
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
func (p *SourcePlugin) RegisterMenuFilter(
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
func (p *SourcePlugin) RegisterPermissionFilter(
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
func (p *SourcePlugin) GetHookHandlers() []*HookHandlerRegistration {
	if p == nil {
		return []*HookHandlerRegistration{}
	}
	items := make([]*HookHandlerRegistration, len(p.hookHandlers))
	copy(items, p.hookHandlers)
	return items
}

// GetRouteRegistrars returns the registered route contribution callbacks.
func (p *SourcePlugin) GetRouteRegistrars() []*RouteHandlerRegistration {
	if p == nil {
		return []*RouteHandlerRegistration{}
	}
	items := make([]*RouteHandlerRegistration, len(p.routeRegistrars))
	copy(items, p.routeRegistrars)
	return items
}

// GetAfterAuthHandlers returns the registered after-auth callbacks.
func (p *SourcePlugin) GetAfterAuthHandlers() []*AfterAuthHandlerRegistration {
	if p == nil {
		return []*AfterAuthHandlerRegistration{}
	}
	items := make([]*AfterAuthHandlerRegistration, len(p.afterAuthHandlers))
	copy(items, p.afterAuthHandlers)
	return items
}

// GetCronRegistrars returns the registered cron contribution callbacks.
func (p *SourcePlugin) GetCronRegistrars() []*CronHandlerRegistration {
	if p == nil {
		return []*CronHandlerRegistration{}
	}
	items := make([]*CronHandlerRegistration, len(p.cronRegistrars))
	copy(items, p.cronRegistrars)
	return items
}

// GetMenuFilters returns the registered menu filter callbacks.
func (p *SourcePlugin) GetMenuFilters() []*MenuFilterHandlerRegistration {
	if p == nil {
		return []*MenuFilterHandlerRegistration{}
	}
	items := make([]*MenuFilterHandlerRegistration, len(p.menuFilters))
	copy(items, p.menuFilters)
	return items
}

// GetPermissionFilters returns the registered permission filter callbacks.
func (p *SourcePlugin) GetPermissionFilters() []*PermissionFilterHandlerRegistration {
	if p == nil {
		return []*PermissionFilterHandlerRegistration{}
	}
	items := make([]*PermissionFilterHandlerRegistration, len(p.permissionFilters))
	copy(items, p.permissionFilters)
	return items
}

func (p *hookPayload) ExtensionPoint() ExtensionPoint {
	if p == nil {
		return ""
	}
	return p.point
}

func (p *hookPayload) Value(key string) interface{} {
	if p == nil {
		return nil
	}
	return p.values[key]
}

func (p *hookPayload) Values() map[string]interface{} {
	if p == nil {
		return map[string]interface{}{}
	}
	return cloneValueMap(p.values)
}

func (i *afterAuthInput) Request() *ghttp.Request {
	if i == nil {
		return nil
	}
	return i.request
}

func (i *afterAuthInput) SetResponseHeader(key string, value string) {
	if i == nil || i.request == nil {
		return
	}
	i.request.Response.Header().Set(key, value)
}

func (i *afterAuthInput) TokenID() string {
	if i == nil {
		return ""
	}
	return i.tokenID
}

func (i *afterAuthInput) UserID() int {
	if i == nil {
		return 0
	}
	return i.userID
}

func (i *afterAuthInput) Username() string {
	if i == nil {
		return ""
	}
	return i.username
}

func (i *afterAuthInput) Status() int {
	if i == nil {
		return 0
	}
	return i.status
}

func (d *menuDescriptor) ID() int {
	if d == nil {
		return 0
	}
	return d.id
}

func (d *menuDescriptor) ParentID() int {
	if d == nil {
		return 0
	}
	return d.parentID
}

func (d *menuDescriptor) Name() string {
	if d == nil {
		return ""
	}
	return d.name
}

func (d *menuDescriptor) Path() string {
	if d == nil {
		return ""
	}
	return d.path
}

func (d *menuDescriptor) Component() string {
	if d == nil {
		return ""
	}
	return d.component
}

func (d *menuDescriptor) Permissions() string {
	if d == nil {
		return ""
	}
	return d.permission
}

func (d *menuDescriptor) MenuKey() string {
	if d == nil {
		return ""
	}
	return d.menuKey
}

func (d *menuDescriptor) Type() string {
	if d == nil {
		return ""
	}
	return d.menuType
}

func (d *menuDescriptor) Visible() int {
	if d == nil {
		return 0
	}
	return d.visible
}

func (d *menuDescriptor) Status() int {
	if d == nil {
		return 0
	}
	return d.status
}

func (d *permissionDescriptor) MenuKey() string {
	if d == nil {
		return ""
	}
	return d.menuKey
}

func (d *permissionDescriptor) MenuName() string {
	if d == nil {
		return ""
	}
	return d.menuName
}

func (d *permissionDescriptor) Permission() string {
	if d == nil {
		return ""
	}
	return d.permission
}

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

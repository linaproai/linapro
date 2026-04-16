// This file bridges pluginhost callback registrations into host route, cron,
// after-auth, menu-filter, and permission-filter integration flows.

package integration

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

// RegisterHTTPRoutes registers callback-contributed HTTP routes for source plugins.
func (s *serviceImpl) RegisterHTTPRoutes(
	ctx context.Context,
	pluginGroup *ghttp.RouterGroup,
	middlewares pluginhost.RouteMiddlewares,
) error {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return err
	}

	checker := s.buildBackgroundEnabledChecker()
	for _, manifest := range manifests {
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		registrar := pluginhost.NewRouteRegistrar(
			pluginGroup,
			manifest.ID,
			checker,
			middlewares,
		)
		for _, handler := range sourcePlugin.GetRouteRegistrars() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			if err = handler.Handler(ctx, registrar); err != nil {
				return err
			}
		}
	}
	return nil
}

// RegisterCrons registers callback-contributed cron jobs for source plugins.
func (s *serviceImpl) RegisterCrons(ctx context.Context) error {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return err
	}

	checker := s.buildBackgroundEnabledChecker()
	for _, manifest := range manifests {
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		registrar := pluginhost.NewCronRegistrar(
			manifest.ID,
			checker,
			s.buildPrimaryNodeChecker(),
		)
		for _, handler := range sourcePlugin.GetCronRegistrars() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			if err = handler.Handler(ctx, registrar); err != nil {
				return err
			}
		}
	}
	return nil
}

// DispatchAfterAuth dispatches callback-style after-auth request handlers.
// It implements runtime.AfterAuthDispatcher.
func (s *serviceImpl) DispatchAfterAuth(
	ctx context.Context,
	input pluginhost.AfterAuthInput,
) {
	if input == nil {
		return
	}

	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		logger.Warningf(ctx, "scan plugin manifests for after-auth dispatch failed: %v", err)
		return
	}
	runtime, runtimeErr := s.buildFilterRuntimeFromManifests(ctx, manifests)
	if runtimeErr != nil {
		logger.Warningf(ctx, "load plugin enablement runtime for after-auth dispatch failed: %v", runtimeErr)
	}

	for _, manifest := range manifests {
		if runtime != nil {
			if !runtime.isEnabled(manifest.ID) {
				continue
			}
		} else if !s.IsEnabled(ctx, manifest.ID) {
			continue
		}
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		for _, handler := range sourcePlugin.GetAfterAuthHandlers() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			if err = handler.Handler(ctx, input); err != nil {
				logger.Warningf(ctx, "plugin after-auth handler failed plugin=%s err=%v", manifest.ID, err)
			}
		}
	}
}

// DispatchPluginHookEvent dispatches one named hook event to all enabled plugins.
// It implements catalog.HookDispatcher and runtime.HookDispatcher.
func (s *serviceImpl) DispatchPluginHookEvent(
	ctx context.Context,
	eventName pluginhost.ExtensionPoint,
	payload map[string]interface{},
) error {
	discoveredManifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return err
	}

	manifests := make([]*catalog.Manifest, 0, len(discoveredManifests))
	for _, manifest := range discoveredManifests {
		if manifest == nil {
			continue
		}
		if catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic {
			manifests = append(manifests, manifest)
			continue
		}
		activeManifest, activeErr := s.catalogSvc.GetActiveManifest(ctx, manifest.ID)
		if activeErr != nil {
			logger.Warningf(ctx, "load active dynamic plugin manifest failed plugin=%s err=%v", manifest.ID, activeErr)
			continue
		}
		manifests = append(manifests, activeManifest)
	}

	runtime, runtimeErr := s.buildFilterRuntimeFromManifests(ctx, manifests)
	if runtimeErr != nil {
		logger.Warningf(ctx, "load plugin enablement runtime for hook dispatch failed: %v", runtimeErr)
	}

	targetPluginID := pluginhost.HookPayloadStringValue(payload, pluginhost.HookPayloadKeyPluginID)
	for _, manifest := range manifests {
		if !s.shouldDispatchHookToPlugin(ctx, runtime, manifest.ID, eventName, targetPluginID) {
			continue
		}
		for _, hook := range manifest.Hooks {
			if hook == nil || hook.Event != eventName {
				continue
			}
			s.executePluginDeclaredHook(ctx, manifest.ID, hook, payload)
		}
		s.executeSourcePluginHookHandlers(ctx, manifest.ID, eventName, payload)
	}
	return nil
}

// FilterMenus filters disabled plugin menus by menu_key prefix "plugin:<plugin-id>".
func (s *serviceImpl) FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	if len(menus) == 0 {
		return menus
	}

	runtime, err := s.buildFilterRuntime(ctx)
	if err != nil {
		return s.filterMenusSlow(ctx, menus)
	}
	return s.filterMenusWithRuntime(ctx, menus, runtime)
}

// FilterPermissionMenus filters permission menus based on plugin enablement and plugin-defined permission visibility.
// It implements runtime.PermissionMenuFilter.
func (s *serviceImpl) FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	if len(menus) == 0 {
		return menus
	}

	runtime, err := s.buildFilterRuntime(ctx)
	if err != nil {
		return s.filterPermissionMenusSlow(ctx, menus)
	}

	filteredMenus := s.filterMenusWithRuntime(ctx, menus, runtime)
	filtered := make([]*entity.SysMenu, 0, len(filteredMenus))
	for _, menu := range filteredMenus {
		if menu == nil {
			continue
		}
		if s.shouldKeepPermissionWithRuntime(ctx, menu, runtime) {
			filtered = append(filtered, menu)
		}
	}
	return filtered
}

// ShouldKeepPermission reports whether a permission should stay effective after plugin filtering.
func (s *serviceImpl) ShouldKeepPermission(ctx context.Context, menu *entity.SysMenu) bool {
	runtime, err := s.buildFilterRuntime(ctx)
	if err != nil {
		return s.shouldKeepPermissionSlow(ctx, menu)
	}
	return s.shouldKeepPermissionWithRuntime(ctx, menu, runtime)
}

func (s *serviceImpl) filterMenusWithRuntime(
	ctx context.Context,
	menus []*entity.SysMenu,
	runtime *filterRuntime,
) []*entity.SysMenu {
	filtered := make([]*entity.SysMenu, 0, len(menus))
	for _, menu := range menus {
		if menu == nil {
			continue
		}
		pluginID := catalog.ParsePluginIDFromMenu(menu)
		if pluginID != "" && !runtime.isEnabled(pluginID) {
			continue
		}
		if s.shouldKeepMenuWithRuntime(ctx, menu, runtime) {
			filtered = append(filtered, menu)
		}
	}
	return filtered
}

func (s *serviceImpl) filterMenusSlow(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	filtered := make([]*entity.SysMenu, 0, len(menus))
	for _, menu := range menus {
		if menu == nil {
			continue
		}
		pluginID := catalog.ParsePluginIDFromMenu(menu)
		if pluginID != "" && !s.IsEnabled(ctx, pluginID) {
			continue
		}
		if s.shouldKeepMenuSlow(ctx, menu) {
			filtered = append(filtered, menu)
		}
	}
	return filtered
}

func (s *serviceImpl) filterPermissionMenusSlow(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	filteredMenus := s.filterMenusSlow(ctx, menus)
	filtered := make([]*entity.SysMenu, 0, len(filteredMenus))
	for _, menu := range filteredMenus {
		if menu == nil {
			continue
		}
		if s.shouldKeepPermissionSlow(ctx, menu) {
			filtered = append(filtered, menu)
		}
	}
	return filtered
}

func (s *serviceImpl) shouldKeepMenuWithRuntime(
	ctx context.Context,
	menu *entity.SysMenu,
	runtime *filterRuntime,
) bool {
	if menu == nil {
		return false
	}

	descriptor := pluginhost.NewMenuDescriptor(
		menu.Id,
		menu.ParentId,
		menu.Name,
		menu.Path,
		menu.Component,
		menu.Perms,
		menu.MenuKey,
		menu.Type,
		menu.Visible,
		menu.Status,
	)

	for _, manifest := range runtime.manifests {
		if !runtime.isEnabled(manifest.ID) {
			continue
		}
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		for _, handler := range sourcePlugin.GetMenuFilters() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			visible, filterErr := handler.Handler(ctx, descriptor)
			if filterErr != nil {
				logger.Warningf(ctx, "plugin menu filter failed plugin=%s err=%v", manifest.ID, filterErr)
				continue
			}
			if !visible {
				return false
			}
		}
	}
	return true
}

func (s *serviceImpl) shouldKeepMenuSlow(ctx context.Context, menu *entity.SysMenu) bool {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		logger.Warningf(ctx, "scan plugin manifests for menu filter failed: %v", err)
		return true
	}

	runtime, runtimeErr := s.buildFilterRuntimeFromManifests(ctx, manifests)
	if runtimeErr != nil {
		logger.Warningf(ctx, "load plugin enablement runtime for menu filter failed: %v", runtimeErr)
	}
	if runtime != nil {
		return s.shouldKeepMenuWithRuntime(ctx, menu, runtime)
	}

	if menu == nil {
		return false
	}

	descriptor := pluginhost.NewMenuDescriptor(
		menu.Id,
		menu.ParentId,
		menu.Name,
		menu.Path,
		menu.Component,
		menu.Perms,
		menu.MenuKey,
		menu.Type,
		menu.Visible,
		menu.Status,
	)
	for _, manifest := range manifests {
		if !s.IsEnabled(ctx, manifest.ID) {
			continue
		}
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		for _, handler := range sourcePlugin.GetMenuFilters() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			visible, filterErr := handler.Handler(ctx, descriptor)
			if filterErr != nil {
				logger.Warningf(ctx, "plugin menu filter failed plugin=%s err=%v", manifest.ID, filterErr)
				continue
			}
			if !visible {
				return false
			}
		}
	}
	return true
}

func (s *serviceImpl) shouldKeepPermissionWithRuntime(
	ctx context.Context,
	menu *entity.SysMenu,
	runtime *filterRuntime,
) bool {
	if menu == nil {
		return false
	}

	descriptor := pluginhost.NewPermissionDescriptor(
		menu.MenuKey,
		menu.Name,
		menu.Perms,
	)

	for _, manifest := range runtime.manifests {
		if !runtime.isEnabled(manifest.ID) {
			continue
		}
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		for _, handler := range sourcePlugin.GetPermissionFilters() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			allowed, filterErr := handler.Handler(ctx, descriptor)
			if filterErr != nil {
				logger.Warningf(ctx, "plugin permission filter failed plugin=%s err=%v", manifest.ID, filterErr)
				continue
			}
			if !allowed {
				return false
			}
		}
	}
	return true
}

func (s *serviceImpl) shouldKeepPermissionSlow(ctx context.Context, menu *entity.SysMenu) bool {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		logger.Warningf(ctx, "scan plugin manifests for permission filter failed: %v", err)
		return true
	}

	runtime, runtimeErr := s.buildFilterRuntimeFromManifests(ctx, manifests)
	if runtimeErr != nil {
		logger.Warningf(ctx, "load plugin enablement runtime for permission filter failed: %v", runtimeErr)
	}
	if runtime != nil {
		return s.shouldKeepPermissionWithRuntime(ctx, menu, runtime)
	}

	if menu == nil {
		return false
	}

	descriptor := pluginhost.NewPermissionDescriptor(
		menu.MenuKey,
		menu.Name,
		menu.Perms,
	)
	for _, manifest := range manifests {
		if !s.IsEnabled(ctx, manifest.ID) {
			continue
		}
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		for _, handler := range sourcePlugin.GetPermissionFilters() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			allowed, filterErr := handler.Handler(ctx, descriptor)
			if filterErr != nil {
				logger.Warningf(ctx, "plugin permission filter failed plugin=%s err=%v", manifest.ID, filterErr)
				continue
			}
			if !allowed {
				return false
			}
		}
	}
	return true
}

// shouldDispatchHookToPlugin determines whether the hook event should be delivered to a plugin.
func (s *serviceImpl) shouldDispatchHookToPlugin(
	ctx context.Context,
	runtime *filterRuntime,
	pluginID string,
	eventName pluginhost.ExtensionPoint,
	targetPluginID string,
) bool {
	switch eventName {
	case pluginhost.ExtensionPointPluginInstalled,
		pluginhost.ExtensionPointPluginEnabled,
		pluginhost.ExtensionPointPluginDisabled,
		pluginhost.ExtensionPointPluginUninstalled:
		return pluginID == targetPluginID
	default:
		if runtime != nil {
			return runtime.isEnabled(pluginID)
		}
		return s.IsEnabled(ctx, pluginID)
	}
}

// executePluginDeclaredHook executes a single declared hook for a plugin.
func (s *serviceImpl) executePluginDeclaredHook(
	ctx context.Context,
	pluginID string,
	hook *catalog.HookSpec,
	payload map[string]interface{},
) {
	if hook == nil {
		return
	}

	execute := func(executeCtx context.Context, hookPayload map[string]interface{}, async bool) {
		var (
			timeoutCtx context.Context
			cancel     context.CancelFunc
		)
		timeoutCtx, cancel = buildPluginHookTimeoutContext(executeCtx, hook)
		defer cancel()

		startedAt := gtime.Now()
		err := s.runPluginDeclaredHook(timeoutCtx, pluginID, hook, hookPayload)
		if err != nil {
			if async {
				logger.Warningf(timeoutCtx, "plugin async declared hook failed plugin=%s event=%s action=%s cost=%s err=%v", pluginID, hook.Event, hook.Action, gtime.Now().Sub(startedAt), err)
				return
			}
			logger.Warningf(timeoutCtx, "plugin declared hook failed plugin=%s event=%s action=%s cost=%s err=%v", pluginID, hook.Event, hook.Action, gtime.Now().Sub(startedAt), err)
			return
		}
		if async {
			logger.Infof(timeoutCtx, "plugin async declared hook succeeded plugin=%s event=%s action=%s cost=%s", pluginID, hook.Event, hook.Action, gtime.Now().Sub(startedAt))
			return
		}
		logger.Infof(timeoutCtx, "plugin declared hook succeeded plugin=%s event=%s action=%s cost=%s", pluginID, hook.Event, hook.Action, gtime.Now().Sub(startedAt))
	}

	values := pluginhost.CloneHookPayloadValues(payload)
	mode := normalizePluginHookMode(hook)
	if mode == pluginhost.CallbackExecutionModeAsync {
		go execute(context.WithoutCancel(ctx), values, true)
		return
	}
	execute(ctx, values, false)
}

// BuildPluginHookTimeoutContext is the exported form of buildPluginHookTimeoutContext for cross-package access.
func BuildPluginHookTimeoutContext(ctx context.Context, hook *catalog.HookSpec) (context.Context, context.CancelFunc) {
	return buildPluginHookTimeoutContext(ctx, hook)
}

// RunPluginDeclaredHook is the exported form of runPluginDeclaredHook for cross-package access.
func (s *serviceImpl) RunPluginDeclaredHook(
	ctx context.Context,
	pluginID string,
	hook *catalog.HookSpec,
	payload map[string]interface{},
) error {
	return s.runPluginDeclaredHook(ctx, pluginID, hook, payload)
}

// buildPluginHookTimeoutContext creates a context with the hook's configured timeout.
func buildPluginHookTimeoutContext(ctx context.Context, hook *catalog.HookSpec) (context.Context, context.CancelFunc) {
	timeout := 3 * time.Second
	if hook != nil && hook.TimeoutMs > 0 {
		timeout = time.Duration(hook.TimeoutMs) * time.Millisecond
	}
	return context.WithTimeout(ctx, timeout)
}

// normalizePluginHookMode resolves the effective execution mode for a hook spec.
func normalizePluginHookMode(hook *catalog.HookSpec) pluginhost.CallbackExecutionMode {
	if hook == nil {
		return pluginhost.CallbackExecutionModeBlocking
	}
	mode := hook.Mode
	if mode == "" {
		mode = pluginhost.DefaultCallbackExecutionMode(hook.Event)
	}
	if !pluginhost.IsExtensionPointExecutionModeSupported(hook.Event, mode) {
		return pluginhost.DefaultCallbackExecutionMode(hook.Event)
	}
	return mode
}

// runPluginDeclaredHook dispatches one declared hook action.
func (s *serviceImpl) runPluginDeclaredHook(
	ctx context.Context,
	pluginID string,
	hook *catalog.HookSpec,
	payload map[string]interface{},
) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = gerror.Newf("plugin declared hook panicked: %v", recovered)
		}
	}()

	if hook == nil {
		return nil
	}

	switch hook.Action {
	case pluginhost.HookActionInsert:
		err = s.executePluginInsertHook(ctx, pluginID, hook, payload)
	case pluginhost.HookActionSleep:
		err = executePluginSleepHook(ctx, hook)
	case pluginhost.HookActionError:
		err = executePluginErrorHook(hook)
	default:
		err = gerror.Newf("插件 Hook 动作不支持: %s", hook.Action)
	}
	if err != nil {
		if ctx.Err() != nil {
			return gerror.Wrapf(ctx.Err(), "plugin hook execution exceeded timeout for %s", hook.Event)
		}
		return err
	}
	if ctx.Err() != nil {
		return gerror.Wrapf(ctx.Err(), "plugin hook execution exceeded timeout for %s", hook.Event)
	}
	return nil
}

// executeSourcePluginHookHandlers executes registered callback-style hook handlers for one source plugin.
func (s *serviceImpl) executeSourcePluginHookHandlers(
	ctx context.Context,
	pluginID string,
	eventName pluginhost.ExtensionPoint,
	payload map[string]interface{},
) {
	sourcePlugin, ok := pluginhost.GetSourcePlugin(pluginID)
	if !ok {
		return
	}
	for _, item := range sourcePlugin.GetHookHandlers() {
		if item == nil || item.Point != eventName || item.Handler == nil {
			continue
		}
		s.executeSourcePluginHookHandler(ctx, pluginID, item, payload)
	}
}

// executeSourcePluginHookHandler executes one registered callback-style hook handler.
func (s *serviceImpl) executeSourcePluginHookHandler(
	ctx context.Context,
	pluginID string,
	item *pluginhost.HookHandlerRegistration,
	payload map[string]interface{},
) {
	if item == nil || item.Handler == nil {
		return
	}

	execute := func(executeCtx context.Context, values map[string]interface{}, async bool) {
		startedAt := gtime.Now()
		if err := item.Handler(executeCtx, pluginhost.NewHookPayload(item.Point, values)); err != nil {
			if async {
				logger.Warningf(executeCtx, "plugin async callback hook failed plugin=%s event=%s cost=%s err=%v", pluginID, item.Point, gtime.Now().Sub(startedAt), err)
				return
			}
			logger.Warningf(executeCtx, "plugin callback hook failed plugin=%s event=%s cost=%s err=%v", pluginID, item.Point, gtime.Now().Sub(startedAt), err)
			return
		}
		if async {
			logger.Infof(executeCtx, "plugin async callback hook succeeded plugin=%s event=%s cost=%s", pluginID, item.Point, gtime.Now().Sub(startedAt))
			return
		}
		logger.Infof(executeCtx, "plugin callback hook succeeded plugin=%s event=%s cost=%s", pluginID, item.Point, gtime.Now().Sub(startedAt))
	}

	values := pluginhost.CloneHookPayloadValues(payload)
	if item.Mode == pluginhost.CallbackExecutionModeAsync {
		go execute(context.WithoutCancel(ctx), values, true)
		return
	}
	execute(ctx, values, false)
}

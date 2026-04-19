// This file synchronizes source-plugin scheduled-job handlers with the host
// plugin lifecycle and startup enablement state.

package jobhandler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/jobmeta"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/pluginhost"
)

// PluginStatusChecker exposes plugin enablement state needed during startup sync.
type PluginStatusChecker interface {
	// IsEnabled reports whether the specified plugin is currently enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
}

// pluginLifecycleObserver maps plugin lifecycle callbacks to registry mutations.
type pluginLifecycleObserver struct {
	registry Registry
}

// Ensure pluginLifecycleObserver implements the plugin lifecycle observer contract.
var _ pluginsvc.LifecycleObserver = (*pluginLifecycleObserver)(nil)

// AttachPluginLifecycle subscribes the registry to synchronous plugin
// lifecycle callbacks and eagerly registers handlers for already-enabled source
// plugins.
func AttachPluginLifecycle(
	ctx context.Context,
	registry Registry,
	checker PluginStatusChecker,
) (func(), error) {
	if registry == nil {
		return nil, gerror.New("任务处理器注册表不能为空")
	}
	if checker == nil {
		return nil, gerror.New("插件状态检查器不能为空")
	}

	observer := &pluginLifecycleObserver{registry: registry}
	unsubscribe := pluginsvc.RegisterLifecycleObserver(observer)
	if err := observer.syncEnabledSourcePlugins(ctx, checker); err != nil {
		unsubscribe()
		return nil, err
	}
	return unsubscribe, nil
}

// OnPluginEnabled registers all scheduled-job handlers declared by one enabled plugin.
func (o *pluginLifecycleObserver) OnPluginEnabled(ctx context.Context, pluginID string) error {
	return o.registerPluginHandlers(strings.TrimSpace(pluginID))
}

// OnPluginDisabled unregisters all scheduled-job handlers declared by one disabled plugin.
func (o *pluginLifecycleObserver) OnPluginDisabled(ctx context.Context, pluginID string) error {
	o.unregisterPluginHandlers(strings.TrimSpace(pluginID))
	return nil
}

// OnPluginUninstalled unregisters all scheduled-job handlers declared by one uninstalled plugin.
func (o *pluginLifecycleObserver) OnPluginUninstalled(ctx context.Context, pluginID string) error {
	o.unregisterPluginHandlers(strings.TrimSpace(pluginID))
	return nil
}

// syncEnabledSourcePlugins registers handlers for all build-linked source
// plugins that are already enabled when the host starts.
func (o *pluginLifecycleObserver) syncEnabledSourcePlugins(
	ctx context.Context,
	checker PluginStatusChecker,
) error {
	for _, sourcePlugin := range pluginhost.ListSourcePlugins() {
		if sourcePlugin == nil || strings.TrimSpace(sourcePlugin.ID) == "" {
			continue
		}
		if !checker.IsEnabled(ctx, sourcePlugin.ID) {
			continue
		}
		if err := o.registerPluginHandlers(sourcePlugin.ID); err != nil {
			return err
		}
	}
	return nil
}

// registerPluginHandlers publishes all scheduled-job handlers declared by one source plugin.
func (o *pluginLifecycleObserver) registerPluginHandlers(pluginID string) error {
	if o == nil || o.registry == nil || pluginID == "" {
		return nil
	}

	sourcePlugin, ok := pluginhost.GetSourcePlugin(pluginID)
	if !ok || sourcePlugin == nil {
		return nil
	}

	// Remove any stale definitions first so repeated enable flows stay idempotent.
	o.unregisterPluginHandlers(pluginID)

	registeredRefs := make([]string, 0, len(sourcePlugin.GetJobHandlers()))
	for _, item := range sourcePlugin.GetJobHandlers() {
		if item == nil {
			continue
		}
		ref, err := buildPluginHandlerRef(pluginID, item.Name)
		if err != nil {
			for _, registeredRef := range registeredRefs {
				o.registry.Unregister(registeredRef)
			}
			return err
		}
		if err = o.registry.Register(HandlerDef{
			Ref:          ref,
			DisplayName:  strings.TrimSpace(item.DisplayName),
			Description:  strings.TrimSpace(item.Description),
			ParamsSchema: strings.TrimSpace(item.ParamsSchema),
			Source:       jobmeta.HandlerSourcePlugin,
			PluginID:     pluginID,
			Invoke:       InvokeFunc(item.Handler),
		}); err != nil {
			for _, registeredRef := range registeredRefs {
				o.registry.Unregister(registeredRef)
			}
			return err
		}
		registeredRefs = append(registeredRefs, ref)
	}
	return nil
}

// unregisterPluginHandlers removes all registry entries owned by one plugin.
func (o *pluginLifecycleObserver) unregisterPluginHandlers(pluginID string) {
	if o == nil || o.registry == nil || pluginID == "" {
		return
	}

	for _, item := range o.registry.List() {
		if item.Source != jobmeta.HandlerSourcePlugin || item.PluginID != pluginID {
			continue
		}
		o.registry.Unregister(item.Ref)
	}
}

// buildPluginHandlerRef converts one plugin-local handler name into the public
// `plugin:<plugin-id>/<name>` registry ref.
func buildPluginHandlerRef(pluginID string, name string) (string, error) {
	trimmedPluginID := strings.TrimSpace(pluginID)
	trimmedName := strings.TrimSpace(name)
	if trimmedPluginID == "" {
		return "", gerror.New("插件ID不能为空")
	}
	if trimmedName == "" {
		return "", gerror.New("插件处理器名称不能为空")
	}
	return fmt.Sprintf("plugin:%s/%s", trimmedPluginID, trimmedName), nil
}

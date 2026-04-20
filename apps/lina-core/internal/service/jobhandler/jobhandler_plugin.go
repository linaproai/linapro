// This file synchronizes source-plugin scheduled-job handlers with the host
// plugin lifecycle and startup enablement state.

package jobhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/jobmeta"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/pluginhost"
)

// PluginLifecycleBridge exposes plugin enablement state and plugin-owned cron
// definitions needed during lifecycle synchronization.
type PluginLifecycleBridge interface {
	// IsEnabled reports whether the specified plugin is currently enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
	// ListManagedCronJobsByPlugin returns plugin-owned cron definitions for one plugin.
	ListManagedCronJobsByPlugin(ctx context.Context, pluginID string) ([]pluginsvc.ManagedCronJob, error)
}

// pluginLifecycleObserver maps plugin lifecycle callbacks to registry mutations.
type pluginLifecycleObserver struct {
	registry Registry             // registry stores published handler definitions.
	bridge   PluginLifecycleBridge // bridge resolves plugin enablement and managed cron definitions.
}

// Ensure pluginLifecycleObserver implements the plugin lifecycle observer contract.
var _ pluginsvc.LifecycleObserver = (*pluginLifecycleObserver)(nil)

// AttachPluginLifecycle subscribes the registry to synchronous plugin
// lifecycle callbacks and eagerly registers handlers for already-enabled source
// plugins.
func AttachPluginLifecycle(
	ctx context.Context,
	registry Registry,
	bridge PluginLifecycleBridge,
) (func(), error) {
	if registry == nil {
		return nil, gerror.New("任务处理器注册表不能为空")
	}
	if bridge == nil {
		return nil, gerror.New("插件生命周期桥接器不能为空")
	}

	observer := &pluginLifecycleObserver{registry: registry, bridge: bridge}
	unsubscribe := pluginsvc.RegisterLifecycleObserver(observer)
	if err := observer.syncEnabledSourcePlugins(ctx, bridge); err != nil {
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
	bridge PluginLifecycleBridge,
) error {
	for _, sourcePlugin := range pluginhost.ListSourcePlugins() {
		if sourcePlugin == nil || strings.TrimSpace(sourcePlugin.ID) == "" {
			continue
		}
		if !bridge.IsEnabled(ctx, sourcePlugin.ID) {
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

	if o.bridge == nil {
		return nil
	}

	managedJobs, err := o.bridge.ListManagedCronJobsByPlugin(context.Background(), pluginID)
	if err != nil {
		for _, registeredRef := range registeredRefs {
			o.registry.Unregister(registeredRef)
		}
		return err
	}
	for _, item := range managedJobs {
		ref, refErr := buildPluginCronHandlerRef(pluginID, item.Name)
		if refErr != nil {
			for _, registeredRef := range registeredRefs {
				o.registry.Unregister(registeredRef)
			}
			return refErr
		}
		handler := item.Handler
		if handler == nil {
			continue
		}
		if err = o.registry.Register(HandlerDef{
			Ref:          ref,
			DisplayName:  buildManagedCronDisplayName(item),
			Description:  buildManagedCronDescription(item),
			ParamsSchema: `{"type":"object","properties":{}}`,
			Source:       jobmeta.HandlerSourcePlugin,
			PluginID:     pluginID,
			Invoke: func(ctx context.Context, _ json.RawMessage) (result any, err error) {
				if runErr := handler(ctx); runErr != nil {
					return nil, runErr
				}
				return map[string]any{"executed": true}, nil
			},
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

// buildPluginCronHandlerRef converts one plugin cron callback name into the
// public `plugin:<plugin-id>/cron:<name>` synthetic handler reference.
func buildPluginCronHandlerRef(pluginID string, name string) (string, error) {
	trimmedPluginID := strings.TrimSpace(pluginID)
	trimmedName := strings.TrimSpace(name)
	if trimmedPluginID == "" {
		return "", gerror.New("插件ID不能为空")
	}
	if trimmedName == "" {
		return "", gerror.New("插件内置定时任务名称不能为空")
	}
	return fmt.Sprintf("plugin:%s/cron:%s", trimmedPluginID, trimmedName), nil
}

// buildManagedCronDisplayName derives the UI display name for one plugin cron definition.
func buildManagedCronDisplayName(item pluginsvc.ManagedCronJob) string {
	if trimmed := strings.TrimSpace(item.DisplayName); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(item.Name)
}

// buildManagedCronDescription derives the UI description for one plugin cron definition.
func buildManagedCronDescription(item pluginsvc.ManagedCronJob) string {
	if trimmed := strings.TrimSpace(item.Description); trimmed != "" {
		return trimmed
	}
	return "插件注册的内置定时任务。"
}

// This file synchronizes source-plugin scheduled-job handlers with the host
// plugin lifecycle and startup enablement state.

package jobhandler

import (
	"context"
	"encoding/json"
	"strings"

	"lina-core/internal/service/jobmeta"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginbridge"
)

// PluginLifecycleBridge exposes plugin enablement state and plugin-owned cron
// definitions needed during lifecycle synchronization.
type PluginLifecycleBridge interface {
	// IsEnabled reports whether the specified plugin is currently enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
	// ListEnabledPluginIDs returns all currently enabled plugin identifiers so
	// startup sync can restore synthetic handlers after host restart.
	ListEnabledPluginIDs(ctx context.Context) ([]string, error)
	// ListManagedCronJobsByPlugin returns plugin-owned cron definitions for one plugin.
	ListManagedCronJobsByPlugin(ctx context.Context, pluginID string) ([]pluginsvc.ManagedCronJob, error)
}

// pluginLifecycleObserver maps plugin lifecycle callbacks to registry mutations.
type pluginLifecycleObserver struct {
	registry Registry              // registry stores published handler definitions.
	bridge   PluginLifecycleBridge // bridge resolves plugin enablement and managed cron definitions.
}

// Ensure pluginLifecycleObserver implements the plugin lifecycle observer contract.
var _ pluginsvc.LifecycleObserver = (*pluginLifecycleObserver)(nil)

// AttachPluginLifecycle subscribes the registry to synchronous plugin
// lifecycle callbacks and eagerly registers handlers for already-enabled
// plugins.
func AttachPluginLifecycle(
	ctx context.Context,
	registry Registry,
	bridge PluginLifecycleBridge,
) (func(), error) {
	if registry == nil {
		return nil, bizerr.NewCode(CodeJobHandlerRegistryRequired)
	}
	if bridge == nil {
		return nil, bizerr.NewCode(CodeJobHandlerLifecycleBridgeRequired)
	}

	observer := &pluginLifecycleObserver{registry: registry, bridge: bridge}
	unsubscribe := pluginsvc.RegisterLifecycleObserver(observer)
	if err := observer.syncEnabledPlugins(ctx, bridge); err != nil {
		unsubscribe()
		return nil, err
	}
	return unsubscribe, nil
}

// OnPluginInstalled is a no-op for the handler registry because executable
// handlers are only published once the plugin becomes enabled.
func (o *pluginLifecycleObserver) OnPluginInstalled(ctx context.Context, pluginID string) error {
	return nil
}

// OnPluginEnabled registers all scheduled-job handlers declared by one enabled plugin.
func (o *pluginLifecycleObserver) OnPluginEnabled(ctx context.Context, pluginID string) error {
	return o.registerPluginHandlers(ctx, strings.TrimSpace(pluginID))
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

// syncEnabledPlugins registers handlers for all plugins that are already
// enabled when the host starts.
func (o *pluginLifecycleObserver) syncEnabledPlugins(
	ctx context.Context,
	bridge PluginLifecycleBridge,
) error {
	if bridge == nil {
		return nil
	}
	pluginIDs, err := bridge.ListEnabledPluginIDs(ctx)
	if err != nil {
		return err
	}
	for _, pluginID := range pluginIDs {
		if err := o.registerPluginHandlers(ctx, strings.TrimSpace(pluginID)); err != nil {
			return err
		}
	}
	return nil
}

// registerPluginHandlers publishes all projected builtin cron handlers declared
// by one enabled plugin.
func (o *pluginLifecycleObserver) registerPluginHandlers(ctx context.Context, pluginID string) error {
	if o == nil || o.registry == nil || pluginID == "" {
		return nil
	}

	// Remove any stale definitions first so repeated enable flows stay idempotent.
	o.unregisterPluginHandlers(pluginID)

	registeredRefs := make([]string, 0)

	if o.bridge == nil {
		return nil
	}

	managedJobs, err := o.bridge.ListManagedCronJobsByPlugin(ctx, pluginID)
	if err != nil {
		for _, registeredRef := range registeredRefs {
			o.registry.Unregister(registeredRef)
		}
		return err
	}
	for _, item := range managedJobs {
		ref, refErr := pluginbridge.BuildPluginCronHandlerRef(pluginID, item.Name)
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
	return "Plugin registered built-in scheduled job."
}

// This file keeps scheduled-job projections synchronized with plugin lifecycle
// transitions so source and dynamic built-in jobs stay visible in management.

package cron

import (
	"context"

	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
)

// pluginLifecycleSyncObserver refreshes built-in scheduled-job projections
// after plugin lifecycle transitions mutate handler availability.
type pluginLifecycleSyncObserver struct {
	svc *serviceImpl
}

// Ensure pluginLifecycleSyncObserver satisfies the plugin lifecycle observer contract.
var _ pluginsvc.LifecycleObserver = (*pluginLifecycleSyncObserver)(nil)

// attachPluginLifecycleObserver subscribes the cron service to plugin
// lifecycle callbacks exactly once.
func (s *serviceImpl) attachPluginLifecycleObserver() {
	if s == nil {
		return
	}

	s.pluginObserverOnce.Do(func() {
		pluginsvc.RegisterLifecycleObserver(&pluginLifecycleSyncObserver{svc: s})
	})
}

// OnPluginInstalled refreshes built-in job projections so newly installed
// plugins surface their managed jobs immediately, even before enablement.
func (o *pluginLifecycleSyncObserver) OnPluginInstalled(ctx context.Context, pluginID string) error {
	return o.syncBuiltinJobs(ctx, pluginID)
}

// OnPluginEnabled refreshes built-in job projections after plugin handlers are registered.
func (o *pluginLifecycleSyncObserver) OnPluginEnabled(ctx context.Context, pluginID string) error {
	return o.syncBuiltinJobs(ctx, pluginID)
}

// OnPluginDisabled refreshes built-in job projections after plugin handlers are removed.
func (o *pluginLifecycleSyncObserver) OnPluginDisabled(ctx context.Context, pluginID string) error {
	return o.syncBuiltinJobs(ctx, pluginID)
}

// OnPluginUninstalled refreshes built-in job projections after plugin handlers are removed.
func (o *pluginLifecycleSyncObserver) OnPluginUninstalled(ctx context.Context, pluginID string) error {
	return o.syncBuiltinJobs(ctx, pluginID)
}

// syncBuiltinJobs reruns the built-in projection pass and logs the triggering
// plugin ID when synchronization fails.
func (o *pluginLifecycleSyncObserver) syncBuiltinJobs(ctx context.Context, pluginID string) error {
	if o == nil || o.svc == nil {
		return nil
	}
	if err := o.svc.syncBuiltinScheduledJobs(ctx); err != nil {
		logger.Warningf(ctx, "sync builtin scheduled jobs after plugin lifecycle failed: plugin=%s err=%v", pluginID, err)
		return err
	}
	return nil
}

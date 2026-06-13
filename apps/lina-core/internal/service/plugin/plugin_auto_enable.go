// This file coordinates startup-time plugin bootstrap so plugin.autoEnable can
// install and enable required plugins before later host wiring runs.

package plugin

import (
	"context"

	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/lifecycle"
)

// BootstrapAutoEnable synchronizes manifests and ensures every plugin listed
// in plugin.autoEnable is installed and enabled before later host wiring runs.
// Per-entry mock-data opt-in flags from config flow into the InstallOptions
// passed down to Install.
func (s *serviceImpl) BootstrapAutoEnable(ctx context.Context) error {
	if _, err := s.syncAndList(ctx); err != nil {
		return err
	}
	entries := s.configSvc.GetPluginAutoEnableEntries(ctx)
	return s.lifecycleSvc.BootstrapAutoEnable(ctx, lifecycle.BootstrapAutoEnableOptions{
		Entries:          lifecycleAutoEnableEntries(entries),
		FrameworkVersion: s.frameworkVersion(ctx),
	})
}

// ReconcileAutoEnabledTenantPlugins applies plugin.autoEnable to tenant-scoped
// plugin governance after source plugins have registered tenant-capability
// providers. Startup auto-enable first installs and enables plugins at the host
// registry level; this later pass converts tenant-scoped entries into the
// platform's new-tenant default policy and asks the linapro-tenant-core provider to
// provision existing tenants.
func (s *serviceImpl) ReconcileAutoEnabledTenantPlugins(ctx context.Context) error {
	if s == nil || s.configSvc == nil {
		return nil
	}
	entries := s.configSvc.GetPluginAutoEnableEntries(ctx)
	return s.lifecycleSvc.ReconcileAutoEnabledTenantPlugins(ctx, lifecycleAutoEnableEntries(entries))
}

// lifecycleAutoEnableEntries maps host config entries to lifecycle-owned pure
// value inputs without making lifecycle depend on the config service package.
func lifecycleAutoEnableEntries(entries []configsvc.PluginAutoEnableEntry) []lifecycle.AutoEnableEntry {
	if len(entries) == 0 {
		return nil
	}
	out := make([]lifecycle.AutoEnableEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, lifecycle.AutoEnableEntry{
			ID:           entry.ID,
			WithMockData: entry.WithMockData,
		})
	}
	return out
}

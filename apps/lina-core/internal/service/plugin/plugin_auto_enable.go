// This file coordinates startup-time plugin bootstrap so plugin.autoEnable can
// install and enable required plugins before later host wiring runs.

package plugin

import (
	"context"
	pluginv1 "lina-core/api/plugin/v1"

	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/logger"
)

// BootstrapBuiltinPlugins synchronizes manifests and converges built-in source
// plugins before plugin.autoEnable runs. It is a startup-only path and reuses
// the internal lifecycle and runtime-upgrade services so ordinary management
// guards do not block host-owned builtin reconciliation.
func (s *serviceImpl) BootstrapBuiltinPlugins(ctx context.Context) error {
	out, readCtx, err := s.buildPluginProjection(ctx, pluginProjectionInput{
		mode: projectionModeList,
		sync: true,
	})
	if err != nil {
		return err
	}
	if err = s.integrationSvc.RefreshEnabledSnapshot(readCtx); err != nil {
		return err
	}
	manifests := out.manifests
	if err = s.lifecycleSvc.BootstrapBuiltinPlugins(readCtx, lifecycle.BootstrapBuiltinOptions{
		Manifests:        manifests,
		FrameworkVersion: s.frameworkVersion(ctx),
		Upgrade:          s.bootstrapBuiltinRuntimeUpgrade,
	}); err != nil {
		return err
	}
	s.warnAutoEnableBuiltinOverlap(ctx, manifests)
	return nil
}

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

// bootstrapBuiltinRuntimeUpgrade executes the unified runtime-upgrade envelope
// with startup confirmation already supplied by host-owned builtin governance.
func (s *serviceImpl) bootstrapBuiltinRuntimeUpgrade(ctx context.Context, pluginID string) error {
	_, err := s.lifecycleSvc.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true})
	return err
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

// warnAutoEnableBuiltinOverlap records a diagnostic warning when plugin.autoEnable
// names a builtin plugin; builtin reconciliation has already converged it, so
// the later auto-enable pass should become a no-op.
func (s *serviceImpl) warnAutoEnableBuiltinOverlap(ctx context.Context, manifests []*SourceManifest) {
	if s == nil || s.configSvc == nil || len(manifests) == 0 {
		return
	}
	builtinIDs := make(map[string]struct{}, len(manifests))
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		if plugintypes.NormalizeDistribution(manifest.Distribution) == pluginv1.PluginDistributionBuiltin {
			builtinIDs[manifest.ID] = struct{}{}
		}
	}
	if len(builtinIDs) == 0 {
		return
	}
	for _, entry := range s.configSvc.GetPluginAutoEnableEntries(ctx) {
		if _, ok := builtinIDs[entry.ID]; !ok {
			continue
		}
		logger.Warningf(
			ctx,
			"plugin.autoEnable contains builtin plugin %s; builtin startup governance already converged it",
			entry.ID,
		)
	}
}

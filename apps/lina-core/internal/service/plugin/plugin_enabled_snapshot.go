// This file keeps the integration-layer enablement snapshot aligned with plugin
// lifecycle transitions so guarded source-plugin routes, cron jobs, and global
// middleware can react immediately without per-request registry lookups.

package plugin

import (
	"context"

	"lina-core/internal/service/plugin/internal/catalog"
)

// syncEnabledSnapshotFromRegistry refreshes the in-memory enablement snapshot
// for one plugin using the latest registry row after a lifecycle transition.
func (s *serviceImpl) syncEnabledSnapshotFromRegistry(ctx context.Context, pluginID string) error {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != catalog.InstalledYes {
		s.integrationSvc.DeletePluginEnabledState(pluginID)
		return nil
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	runtimeState, err := s.catalogSvc.BuildRuntimeUpgradeState(ctx, registry, manifest)
	if err != nil {
		return err
	}
	enabled := registry.Status == catalog.StatusEnabled &&
		catalog.RuntimeStateAllowsBusinessEntry(runtimeState.State)
	s.integrationSvc.SetPluginEnabledState(pluginID, enabled)
	return nil
}

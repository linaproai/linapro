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
	s.integrationSvc.SetPluginEnabledState(pluginID, registry.Status == catalog.StatusEnabled)
	return nil
}

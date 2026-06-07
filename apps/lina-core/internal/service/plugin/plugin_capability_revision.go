// This file keeps plugin capability availability aligned with lifecycle
// revisions by refreshing enabled snapshots and publishing runtime changes.

package plugin

import (
	"context"

	"lina-core/internal/service/plugin/internal/catalog"
)

// syncEnabledSnapshotFromRegistry refreshes the in-memory enablement snapshot
// for one plugin using the latest registry row after a lifecycle transition.
func (s *serviceImpl) syncEnabledSnapshotFromRegistry(ctx context.Context, pluginID string) error {
	return s.syncEnabledSnapshotStateFromRegistry(ctx, pluginID)
}

// syncEnabledSnapshotStateFromRegistry updates only the in-memory enabled
// snapshot for the same registry state.
func (s *serviceImpl) syncEnabledSnapshotStateFromRegistry(
	ctx context.Context,
	pluginID string,
) error {
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

// syncEnabledSnapshotAndPublishRuntimeChange updates local enablement, publishes
// the runtime revision, and lets capability providers observe the refreshed
// platform enabled snapshot at use time.
func (s *serviceImpl) syncEnabledSnapshotAndPublishRuntimeChange(
	ctx context.Context,
	pluginID string,
	reason string,
) error {
	if err := s.syncEnabledSnapshotStateFromRegistry(ctx, pluginID); err != nil {
		return err
	}
	_, err := s.markRuntimeCacheChanged(ctx, reason)
	return err
}

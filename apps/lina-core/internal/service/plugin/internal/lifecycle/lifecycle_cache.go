// This file coordinates lifecycle-owned cache publication during the C-stage
// migration before all plugin-change publication paths are unified.

package lifecycle

import (
	"context"

	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
)

// syncEnabledSnapshotAndPublishRuntimeChange updates local enablement and
// publishes the runtime revision for one successful lifecycle transition.
func (s *serviceImpl) syncEnabledSnapshotAndPublishRuntimeChange(
	ctx context.Context,
	pluginID string,
	reason string,
) error {
	if err := s.syncEnabledSnapshotStateFromRegistry(ctx, pluginID); err != nil {
		return err
	}
	if s.cachePublisher == nil {
		return nil
	}
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	pluginType := ""
	if registry != nil {
		pluginType = registry.Type
	}
	return s.cachePublisher.PublishPluginChange(ctx, pluginID, pluginType, reason)
}

// syncEnabledSnapshotStateFromRegistry updates only the in-memory enabled
// snapshot for the same registry state.
func (s *serviceImpl) syncEnabledSnapshotStateFromRegistry(
	ctx context.Context,
	pluginID string,
) error {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != plugintypes.InstalledYes {
		s.integrationSvc.DeletePluginEnabledState(pluginID)
		return nil
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	runtimeState, err := s.storeSvc.BuildRuntimeUpgradeState(ctx, registry, manifest)
	if err != nil {
		return err
	}
	enabled := registry.Status == plugintypes.StatusEnabled &&
		store.RuntimeStateAllowsBusinessEntry(runtimeState.State)
	s.integrationSvc.SetPluginEnabledState(pluginID, enabled)
	return nil
}

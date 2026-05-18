// This file contains dynamic-plugin reconciliation rollback and failed-release
// projection helpers.

package runtime

import (
	"context"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/logger"
)

// rollbackInstallOrUpgrade attempts to restore the last stable plugin state
// after install, upgrade, or refresh work fails midway through reconciliation.
func (s *serviceImpl) rollbackInstallOrUpgrade(
	ctx context.Context,
	registry *entity.SysPlugin,
	restoreManifest *catalog.Manifest,
	failedManifest *catalog.Manifest,
	failedReleaseID int,
	reconcileErr error,
) error {
	if failedManifest != nil {
		if rollbackErr := s.lifecycleSvc.ExecuteManifestSQLFiles(ctx, failedManifest, catalog.MigrationDirectionRollback); rollbackErr != nil {
			logger.Warningf(ctx, "rollback dynamic plugin SQL failed plugin=%s err=%v", failedManifest.ID, rollbackErr)
		}
		if menuErr := s.deletePluginMenusByManifest(ctx, failedManifest); menuErr != nil {
			logger.Warningf(ctx, "delete dynamic plugin menus during rollback failed plugin=%s err=%v", failedManifest.ID, menuErr)
		}
	}
	if restoreManifest != nil {
		if restoreErr := s.syncPluginMenusAndPermissions(ctx, restoreManifest); restoreErr != nil {
			logger.Warningf(ctx, "restore previous plugin menus and permissions failed plugin=%s err=%v", restoreManifest.ID, restoreErr)
		}
	}
	return s.rollbackReleaseFailure(ctx, registry, failedReleaseID, reconcileErr)
}

// rollbackReleaseFailure marks the release and node projection as failed when a
// reconciliation error cannot be hidden behind a fully restored stable state.
func (s *serviceImpl) rollbackReleaseFailure(
	ctx context.Context,
	registry *entity.SysPlugin,
	releaseID int,
	reconcileErr error,
) error {
	restoredRegistry := registry
	if releaseID > 0 {
		if updateErr := s.catalogSvc.UpdateReleaseState(ctx, releaseID, catalog.ReleaseStatusFailed, ""); updateErr != nil {
			logger.Warningf(ctx, "mark dynamic plugin release failed failed plugin=%s releaseID=%d err=%v", registry.PluginId, releaseID, updateErr)
		}
	}
	if restored, err := s.restoreStableState(ctx, registry); err == nil && restored != nil {
		restoredRegistry = restored
	}
	if restoredRegistry != nil {
		if syncErr := s.syncNodeProjection(ctx, nodeProjectionInput{
			PluginID:     restoredRegistry.PluginId,
			ReleaseID:    restoredRegistry.ReleaseId,
			DesiredState: restoredRegistry.DesiredState,
			CurrentState: catalog.NodeStateFailed.String(),
			Generation:   restoredRegistry.Generation,
			Message:      reconcileErr.Error(),
		}); syncErr != nil {
			logger.Warningf(ctx, "sync failed-node projection failed plugin=%s err=%v", restoredRegistry.PluginId, syncErr)
		}
	}
	return reconcileErr
}

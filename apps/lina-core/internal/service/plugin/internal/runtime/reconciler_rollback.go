// This file contains dynamic-plugin reconciliation rollback and failed-release
// projection helpers.

package runtime

import (
	"context"
	"errors"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
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
	rollbackErrs := make([]error, 0, 3)
	if failedManifest != nil {
		if rollbackErr := s.lifecycleSvc.ExecuteManifestSQLFiles(ctx, failedManifest, catalog.MigrationDirectionRollback); rollbackErr != nil {
			logger.Warningf(ctx, "rollback dynamic plugin SQL failed plugin=%s err=%v", failedManifest.ID, rollbackErr)
			rollbackErrs = append(rollbackErrs, rollbackErr)
		}
		if menuErr := s.deletePluginMenusByManifest(ctx, failedManifest); menuErr != nil {
			logger.Warningf(ctx, "delete dynamic plugin menus during rollback failed plugin=%s err=%v", failedManifest.ID, menuErr)
			rollbackErrs = append(rollbackErrs, menuErr)
		}
	}
	if restoreManifest != nil {
		if restoreErr := s.syncPluginMenusAndPermissions(ctx, restoreManifest); restoreErr != nil {
			logger.Warningf(ctx, "restore previous plugin menus and permissions failed plugin=%s err=%v", restoreManifest.ID, restoreErr)
			rollbackErrs = append(rollbackErrs, restoreErr)
		}
	}
	return s.rollbackReleaseFailure(ctx, registry, failedReleaseID, appendRollbackDiagnostics(reconcileErr, rollbackErrs...))
}

// rollbackReleaseFailure marks the release and node projection as failed when a
// reconciliation error cannot be hidden behind a fully restored stable state.
func (s *serviceImpl) rollbackReleaseFailure(
	ctx context.Context,
	registry *entity.SysPlugin,
	releaseID int,
	reconcileErr error,
) error {
	if reconcileErr == nil {
		reconcileErr = errors.New("dynamic plugin reconciliation failed")
	}
	restoredRegistry := registry
	if releaseID > 0 {
		if updateErr := s.catalogSvc.UpdateReleaseState(ctx, releaseID, catalog.ReleaseStatusFailed, ""); updateErr != nil {
			logger.Warningf(ctx, "mark dynamic plugin release failed failed plugin=%s releaseID=%d err=%v", rollbackPluginID(registry), releaseID, updateErr)
			reconcileErr = appendRollbackDiagnostics(reconcileErr, updateErr)
		}
	}
	if restored, err := s.restoreStableState(ctx, registry); err == nil && restored != nil {
		restoredRegistry = restored
	} else if err != nil {
		logger.Warningf(ctx, "restore dynamic plugin stable state failed plugin=%s err=%v", rollbackPluginID(registry), err)
		reconcileErr = appendRollbackDiagnostics(reconcileErr, err)
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
			reconcileErr = appendRollbackDiagnostics(reconcileErr, syncErr)
		}
	}
	if registry != nil && releaseID > 0 {
		if markErr := s.markRegistryFailed(ctx, registry.PluginId); markErr != nil {
			logger.Warningf(ctx, "mark dynamic plugin registry failed failed plugin=%s err=%v", registry.PluginId, markErr)
			reconcileErr = appendRollbackDiagnostics(reconcileErr, markErr)
		}
	}
	return reconcileErr
}

// appendRollbackDiagnostics keeps the original reconciliation failure while
// surfacing every rollback failure in the returned error chain.
func appendRollbackDiagnostics(reconcileErr error, rollbackErrs ...error) error {
	if len(rollbackErrs) == 0 {
		return reconcileErr
	}
	errs := make([]error, 0, len(rollbackErrs)+1)
	if reconcileErr != nil {
		errs = append(errs, reconcileErr)
	}
	for _, rollbackErr := range rollbackErrs {
		if rollbackErr != nil {
			errs = append(errs, rollbackErr)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// markRegistryFailed records the authoritative failed state on sys_plugin so a
// rollback failure is readable beyond logs in later governance projections.
func (s *serviceImpl) markRegistryFailed(ctx context.Context, pluginID string) error {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return nil
	}
	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(do.SysPlugin{CurrentState: catalog.HostStateFailed.String()}).
		Update()
	if err != nil {
		return err
	}
	_, err = s.catalogSvc.RefreshStartupRegistry(ctx, pluginID)
	return err
}

// rollbackPluginID returns a safe plugin identifier for diagnostics.
func rollbackPluginID(registry *entity.SysPlugin) string {
	if registry == nil {
		return ""
	}
	return registry.PluginId
}

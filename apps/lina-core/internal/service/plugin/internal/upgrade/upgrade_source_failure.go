// This file contains best-effort runtime-upgrade failure diagnostics shared by
// source and dynamic upgrade strategies.

package upgrade

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/logger"
)

// markSourcePluginReleaseFailed best-effort records that one explicit
// source-plugin upgrade stopped after the target release had been prepared.
func (s *serviceImpl) markSourcePluginReleaseFailed(
	ctx context.Context,
	manifest *catalog.Manifest,
	release *store.ReleaseRecord,
) {
	if manifest == nil || release == nil {
		return
	}
	if err := s.storeSvc.UpdateReleaseState(
		ctx,
		release.Id,
		plugintypes.ReleaseStatusFailed,
		s.storeSvc.BuildPackagePath(manifest),
	); err != nil {
		return
	}
}

// markSourcePluginUpgradeFailed best-effort records the failed phase and marks
// the target release as failed without switching the effective release.
func (s *serviceImpl) markSourcePluginUpgradeFailed(
	ctx context.Context,
	manifest *catalog.Manifest,
	release *store.ReleaseRecord,
	phase string,
	upgradeErr error,
) {
	s.markSourcePluginReleaseFailed(ctx, manifest, release)
	if manifest == nil || release == nil || upgradeErr == nil {
		return
	}
	if err := recordRuntimeUpgradeFailureMigration(ctx, manifest.ID, release.Id, phase, upgradeErr); err != nil {
		logger.Warningf(
			ctx,
			"record source plugin upgrade failure failed plugin=%s phase=%s err=%v",
			manifest.ID, phase, err,
		)
	}
}

// recordRuntimeUpgradeFailureMigration stores a synthetic upgrade migration
// entry for failed upgrade phases so management UIs can diagnose and retry safely.
func recordRuntimeUpgradeFailureMigration(
	ctx context.Context,
	pluginID string,
	releaseID int,
	phase string,
	upgradeErr error,
) error {
	normalizedPhase := strings.TrimSpace(phase)
	if normalizedPhase == "" {
		normalizedPhase = "unknown"
	}
	errorMessage := strings.TrimSpace(upgradeErr.Error())
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(normalizedPhase+":"+errorMessage)))
	executedAt := time.Now()
	data := do.SysPluginMigration{
		PluginId:       strings.TrimSpace(pluginID),
		ReleaseId:      releaseID,
		Phase:          plugintypes.MigrationDirectionUpgrade.String(),
		MigrationKey:   "upgrade-phase-" + normalizedPhase,
		ExecutionOrder: 0,
		Checksum:       checksum,
		Status:         plugintypes.MigrationExecutionStatusFailed.String(),
		ErrorMessage:   errorMessage,
		ExecutedAt:     &executedAt,
	}

	existing := (*entity.SysPluginMigration)(nil)
	err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:     data.PluginId,
			ReleaseId:    data.ReleaseId,
			Phase:        data.Phase,
			MigrationKey: data.MigrationKey,
		}).
		Scan(&existing)
	if err != nil {
		return err
	}
	if existing == nil {
		_, err = dao.SysPluginMigration.Ctx(ctx).Data(data).Insert()
		return err
	}
	_, err = dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{Id: existing.Id}).
		Data(data).
		Update()
	return err
}

// This file contains best-effort source-plugin upgrade failure diagnostics.

package sourceupgrade

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/logger"
)

// markSourcePluginReleaseFailed best-effort records that one explicit
// source-plugin upgrade stopped after the target release had been prepared.
func (s *serviceImpl) markSourcePluginReleaseFailed(
	ctx context.Context,
	manifest *catalog.Manifest,
	release *entity.SysPluginRelease,
) {
	if manifest == nil || release == nil {
		return
	}
	if err := s.catalogSvc.UpdateReleaseState(
		ctx,
		release.Id,
		catalog.ReleaseStatusFailed,
		s.catalogSvc.BuildPackagePath(manifest),
	); err != nil {
		return
	}
}

// markSourcePluginUpgradeFailed best-effort records the failed phase and marks
// the target release as failed without switching the effective release.
func (s *serviceImpl) markSourcePluginUpgradeFailed(
	ctx context.Context,
	manifest *catalog.Manifest,
	release *entity.SysPluginRelease,
	phase string,
	upgradeErr error,
) {
	s.markSourcePluginReleaseFailed(ctx, manifest, release)
	if manifest == nil || release == nil || upgradeErr == nil {
		return
	}
	if err := recordSourceUpgradeFailureMigration(ctx, manifest.ID, release.Id, phase, upgradeErr); err != nil {
		logger.Warningf(
			ctx,
			"record source plugin upgrade failure failed plugin=%s phase=%s err=%v",
			manifest.ID, phase, err,
		)
	}
}

// recordSourceUpgradeFailureMigration stores a synthetic upgrade migration
// entry for non-SQL phases so management UIs can diagnose and retry safely.
func recordSourceUpgradeFailureMigration(
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
	data := do.SysPluginMigration{
		PluginId:       strings.TrimSpace(pluginID),
		ReleaseId:      releaseID,
		Phase:          catalog.MigrationDirectionUpgrade.String(),
		MigrationKey:   "upgrade-phase-" + normalizedPhase,
		ExecutionOrder: 0,
		Checksum:       checksum,
		Status:         catalog.MigrationExecutionStatusFailed.String(),
		ErrorMessage:   errorMessage,
		ExecutedAt:     gtime.Now(),
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

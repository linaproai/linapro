// This file executes plugin SQL migrations and records abstract migration
// history entries for later review and lifecycle reconciliation.

package lifecycle

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

// SQLAsset describes one install/uninstall SQL step after host extraction.
type SQLAsset struct {
	// Key is the canonical identifier for this SQL step.
	Key string
	// Content is the raw SQL text to execute.
	Content string
}

// ExecuteManifestSQLFiles executes plugin manifest SQL files and records every attempt
// in sys_plugin_migration.
func (s *serviceImpl) ExecuteManifestSQLFiles(
	ctx context.Context,
	manifest *catalog.Manifest,
	direction catalog.MigrationDirection,
) error {
	if manifest == nil {
		return gerror.New("plugin manifest cannot be nil")
	}

	sqlAssets, err := s.ResolveSQLAssets(manifest, direction)
	if err != nil {
		return err
	}

	for index, asset := range sqlAssets {
		if asset == nil {
			return gerror.New("插件 SQL 资源不能为空")
		}

		checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(asset.Content)))
		release, err := s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
		if err != nil {
			return err
		}
		if release == nil {
			return gerror.Newf("插件发布记录不存在: %s@%s", manifest.ID, manifest.Version)
		}
		migrationKey := buildMigrationKey(direction, index+1)
		executedAt := gtime.Now()
		_, execErr := g.DB().Exec(ctx, asset.Content)
		if recordErr := s.recordMigration(ctx, manifest.ID, release.Id, direction, migrationKey, index+1, checksum, executedAt, execErr); recordErr != nil {
			return recordErr
		}
		if execErr != nil {
			return gerror.Wrapf(execErr, "执行插件SQL失败: %s", asset.Key)
		}
	}
	return nil
}

// ResolveSQLAssets extracts lifecycle SQL either from embedded runtime artifact sections
// or from source-style directory conventions, while preserving execution order.
func (s *serviceImpl) ResolveSQLAssets(
	manifest *catalog.Manifest,
	direction catalog.MigrationDirection,
) ([]*SQLAsset, error) {
	if manifest == nil {
		return []*SQLAsset{}, nil
	}

	if manifest.RuntimeArtifact != nil {
		embeddedAssets := manifest.RuntimeArtifact.InstallSQLAssets
		if direction == catalog.MigrationDirectionUninstall {
			embeddedAssets = manifest.RuntimeArtifact.UninstallSQLAssets
		}
		if len(embeddedAssets) > 0 {
			items := make([]*SQLAsset, 0, len(embeddedAssets))
			for _, asset := range embeddedAssets {
				if asset == nil {
					continue
				}
				items = append(items, &SQLAsset{
					Key:     asset.Key,
					Content: asset.Content,
				})
			}
			return items, nil
		}
	}

	var relativePaths []string
	if direction == catalog.MigrationDirectionUninstall {
		relativePaths = s.catalogSvc.ListUninstallSQLPaths(manifest)
	} else {
		relativePaths = s.catalogSvc.ListInstallSQLPaths(manifest)
	}
	items := make([]*SQLAsset, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		sqlContent, err := s.catalogSvc.ReadSourcePluginAssetContent(manifest, relativePath)
		if err != nil {
			return nil, err
		}
		if sqlContent == "" {
			return nil, gerror.Newf("插件SQL文件为空: %s", relativePath)
		}
		items = append(items, &SQLAsset{
			Key:     filepath.Base(relativePath),
			Content: sqlContent,
		})
	}
	return items, nil
}

// ResolvePluginSQLAssets resolves SQL assets from the manifest and returns them as catalog.ArtifactSQLAsset
// slices for callers that expect the catalog asset type rather than lifecycle.SQLAsset.
func (s *serviceImpl) ResolvePluginSQLAssets(manifest *catalog.Manifest, direction catalog.MigrationDirection) ([]*catalog.ArtifactSQLAsset, error) {
	assets, err := s.ResolveSQLAssets(manifest, direction)
	if err != nil {
		return nil, err
	}
	result := make([]*catalog.ArtifactSQLAsset, 0, len(assets))
	for _, a := range assets {
		if a == nil {
			continue
		}
		result = append(result, &catalog.ArtifactSQLAsset{Key: a.Key, Content: a.Content})
	}
	return result, nil
}

// buildMigrationKey returns the canonical migration key for a given phase and sequence number.
func buildMigrationKey(direction catalog.MigrationDirection, sequenceNo int) string {
	normalizedDirection := strings.TrimSpace(strings.ToLower(direction.String()))
	if normalizedDirection == "" {
		normalizedDirection = catalog.MigrationDirectionInstall.String()
	}
	if sequenceNo <= 0 {
		sequenceNo = 1
	}
	return fmt.Sprintf("%s-step-%03d", normalizedDirection, sequenceNo)
}

func (s *serviceImpl) getMigration(
	ctx context.Context,
	pluginID string,
	releaseID int,
	phase catalog.MigrationDirection,
	migrationKey string,
) (*entity.SysPluginMigration, error) {
	var migration *entity.SysPluginMigration
	err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:     pluginID,
			ReleaseId:    releaseID,
			Phase:        phase.String(),
			MigrationKey: migrationKey,
		}).
		Scan(&migration)
	return migration, err
}

func (s *serviceImpl) recordMigration(
	ctx context.Context,
	pluginID string,
	releaseID int,
	phase catalog.MigrationDirection,
	migrationKey string,
	sequenceNo int,
	checksum string,
	executedAt *gtime.Time,
	execErr error,
) error {
	status := catalog.MigrationExecutionStatusSucceeded
	message := ""
	if execErr != nil {
		status = catalog.MigrationExecutionStatusFailed
		message = execErr.Error()
	}

	existing, err := s.getMigration(ctx, pluginID, releaseID, phase, migrationKey)
	if err != nil {
		return err
	}

	data := do.SysPluginMigration{
		PluginId:       pluginID,
		ReleaseId:      releaseID,
		Phase:          phase.String(),
		MigrationKey:   migrationKey,
		ExecutionOrder: sequenceNo,
		Checksum:       checksum,
		Status:         status.String(),
		ErrorMessage:   message,
		ExecutedAt:     executedAt,
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

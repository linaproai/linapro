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
// in sys_plugin_migration. The mock phase is intentionally excluded from this entry
// point because mock data must be loaded transactionally via
// ExecuteManifestMockSQLFilesInTx; callers that want to load mock data MUST go
// through that method instead.
func (s *serviceImpl) ExecuteManifestSQLFiles(
	ctx context.Context,
	manifest *catalog.Manifest,
	direction catalog.MigrationDirection,
) error {
	if manifest == nil {
		return gerror.New("plugin manifest cannot be nil")
	}
	if direction == catalog.MigrationDirectionMock {
		return gerror.New("mock SQL files must be executed via ExecuteManifestMockSQLFilesInTx")
	}

	sqlAssets, err := s.ResolveSQLAssets(manifest, direction)
	if err != nil {
		return err
	}

	for index, asset := range sqlAssets {
		if asset == nil {
			return gerror.New("plugin SQL asset cannot be nil")
		}

		checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(asset.Content)))
		release, err := s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
		if err != nil {
			return err
		}
		if release == nil {
			return gerror.Newf("plugin release record does not exist: %s@%s", manifest.ID, manifest.Version)
		}
		migrationKey := buildMigrationKey(direction, index+1)
		executedAt := gtime.Now()
		_, execErr := g.DB().Exec(ctx, asset.Content)
		if recordErr := s.recordMigration(ctx, manifest.ID, release.Id, direction, migrationKey, index+1, checksum, executedAt, execErr); recordErr != nil {
			return recordErr
		}
		if execErr != nil {
			return gerror.Wrapf(execErr, "execute plugin SQL failed: %s", asset.Key)
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
		var embeddedAssets []*catalog.ArtifactSQLAsset
		switch direction {
		case catalog.MigrationDirectionUninstall:
			embeddedAssets = manifest.RuntimeArtifact.UninstallSQLAssets
		case catalog.MigrationDirectionMock:
			embeddedAssets = manifest.RuntimeArtifact.MockSQLAssets
		default:
			embeddedAssets = manifest.RuntimeArtifact.InstallSQLAssets
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
	switch direction {
	case catalog.MigrationDirectionUninstall:
		relativePaths = s.catalogSvc.ListUninstallSQLPaths(manifest)
	case catalog.MigrationDirectionMock:
		relativePaths = s.catalogSvc.ListMockSQLPaths(manifest)
	default:
		relativePaths = s.catalogSvc.ListInstallSQLPaths(manifest)
	}
	items := make([]*SQLAsset, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		sqlContent, err := s.catalogSvc.ReadSourcePluginAssetContent(manifest, relativePath)
		if err != nil {
			return nil, err
		}
		if sqlContent == "" {
			return nil, gerror.Newf("plugin SQL file is empty: %s", relativePath)
		}
		items = append(items, &SQLAsset{
			Key:     filepath.Base(relativePath),
			Content: sqlContent,
		})
	}
	return items, nil
}

// MockSQLExecutionResult describes the outcome of one mock-data SQL load attempt.
// On failure the caller is expected to roll back the surrounding transaction so
// that every entry in ExecutedFiles ceases to exist alongside the failing file.
type MockSQLExecutionResult struct {
	// ExecutedFiles lists mock SQL filenames that were successfully executed
	// before the load either completed or hit a failure. When Err is non-nil
	// these files are about to be rolled back together with FailedFile.
	ExecutedFiles []string
	// FailedFile names the mock SQL file that triggered the failure. Empty when
	// the load completed successfully or when no mock SQL exists.
	FailedFile string
	// Err carries the underlying database error so callers can surface it to
	// users via the standard bizerr error code wrapping. Nil on success.
	Err error
}

// MockDataLoadError carries the structured details of a rolled-back mock-data
// load so the plugin facade can wrap it once into a stable user-facing bizerr
// regardless of whether the failure surfaced from the source-plugin path or the
// dynamic-plugin reconciler. Use errors.As to recover this type from any
// wrapped chain, then surface PluginID/FailedFile/RolledBackFiles/Cause to the
// caller.
type MockDataLoadError struct {
	// PluginID identifies the plugin whose mock-data load failed.
	PluginID string
	// FailedFile is the mock SQL filename that triggered the failure.
	FailedFile string
	// RolledBackFiles enumerates every mock SQL filename whose effects were
	// reverted, including those that ran successfully prior to FailedFile.
	RolledBackFiles []string
	// Cause is the underlying database/error layer error that triggered the
	// rollback. Surfaced to operators in the user-facing message.
	Cause error
}

// Error implements the error interface for MockDataLoadError.
func (e *MockDataLoadError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause == nil {
		return "plugin mock data load rolled back"
	}
	return "plugin mock data load rolled back: " + e.Cause.Error()
}

// Unwrap returns the underlying database error so errors.Is/errors.As callers
// can introspect or compare the original cause when needed.
func (e *MockDataLoadError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// ExecuteManifestMockSQLFilesInTx executes a plugin's mock-data SQL files inside the
// caller-supplied transaction (carried via ctx) and records each step in
// sys_plugin_migration with phase=mock. The caller MUST run this method inside a
// dao.SysPluginMigration.Transaction(...) closure: on any failure, returning the
// resulting Err from the closure rolls back both the executed mock data rows and
// their migration ledger entries together. The install/uninstall phases are NOT
// affected by the rollback because they execute outside this transaction.
func (s *serviceImpl) ExecuteManifestMockSQLFilesInTx(
	ctx context.Context,
	manifest *catalog.Manifest,
) MockSQLExecutionResult {
	if manifest == nil {
		return MockSQLExecutionResult{Err: gerror.New("plugin manifest cannot be nil")}
	}

	sqlAssets, err := s.ResolveSQLAssets(manifest, catalog.MigrationDirectionMock)
	if err != nil {
		return MockSQLExecutionResult{Err: err}
	}
	if len(sqlAssets) == 0 {
		return MockSQLExecutionResult{}
	}

	release, err := s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return MockSQLExecutionResult{Err: err}
	}
	if release == nil {
		return MockSQLExecutionResult{
			Err: gerror.Newf("plugin release record does not exist: %s@%s", manifest.ID, manifest.Version),
		}
	}

	executed := make([]string, 0, len(sqlAssets))
	for index, asset := range sqlAssets {
		if asset == nil {
			return MockSQLExecutionResult{
				ExecutedFiles: executed,
				Err:           gerror.New("plugin SQL asset cannot be nil"),
			}
		}

		checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(asset.Content)))
		migrationKey := buildMigrationKey(catalog.MigrationDirectionMock, index+1)
		executedAt := gtime.Now()
		if _, execErr := g.DB().Exec(ctx, asset.Content); execErr != nil {
			return MockSQLExecutionResult{
				ExecutedFiles: executed,
				FailedFile:    asset.Key,
				Err:           gerror.Wrapf(execErr, "execute plugin mock SQL failed: %s", asset.Key),
			}
		}
		if recordErr := s.recordMigration(
			ctx,
			manifest.ID,
			release.Id,
			catalog.MigrationDirectionMock,
			migrationKey,
			index+1,
			checksum,
			executedAt,
			nil,
		); recordErr != nil {
			return MockSQLExecutionResult{
				ExecutedFiles: executed,
				FailedFile:    asset.Key,
				Err:           recordErr,
			}
		}
		executed = append(executed, asset.Key)
	}
	return MockSQLExecutionResult{ExecutedFiles: executed}
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

// getMigration loads one previously recorded migration attempt for the given
// release and migration key.
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

// recordMigration upserts the execution result for one SQL migration step so
// install and uninstall attempts remain auditable and re-runnable.
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

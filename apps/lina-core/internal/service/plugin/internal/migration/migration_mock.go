// This file executes plugin mock-data SQL in transactional lifecycle phases and
// reports rollback diagnostics for caller-level user-facing error wrapping.

package migration

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/logger"
)

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

	sqlAssets, err := s.ResolveSQLAssets(manifest, plugintypes.MigrationDirectionMock)
	if err != nil {
		return MockSQLExecutionResult{Err: err}
	}
	if len(sqlAssets) == 0 {
		return MockSQLExecutionResult{}
	}

	release, err := s.storeSvc.GetRelease(ctx, manifest.ID, manifest.Version)
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
		migrationKey := buildMigrationKey(plugintypes.MigrationDirectionMock, index+1)
		executedAt := time.Now()
		if execErr := s.executeSQLAsset(ctx, manifest.ID, plugintypes.MigrationDirectionMock, asset); execErr != nil {
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
			plugintypes.MigrationDirectionMock,
			migrationKey,
			index+1,
			checksum,
			&executedAt,
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

// ExecuteManifestMockSQLFiles executes a plugin's optional mock-data SQL in one
// transaction. On rollback it returns a MockDataLoadError carrying the failed
// file, rolled-back files, and underlying cause.
func (s *serviceImpl) ExecuteManifestMockSQLFiles(ctx context.Context, manifest *catalog.Manifest) error {
	var (
		executedFiles []string
		failedFile    string
		causeErr      error
	)
	txErr := dao.SysPluginMigration.Transaction(ctx, func(txCtx context.Context, _ gdb.TX) error {
		result := s.ExecuteManifestMockSQLFilesInTx(txCtx, manifest)
		executedFiles = append(executedFiles[:0], result.ExecutedFiles...)
		failedFile = result.FailedFile
		if result.Err != nil {
			causeErr = result.Err
			return result.Err
		}
		return nil
	})
	if txErr == nil {
		return nil
	}
	if causeErr == nil {
		causeErr = txErr
	}
	pluginID := ""
	if manifest != nil {
		pluginID = manifest.ID
	}
	logger.Warningf(
		ctx,
		"plugin mock data load rolled back plugin=%s failedFile=%s cause=%v",
		pluginID,
		failedFile,
		causeErr,
	)
	rolledBack := make([]string, 0, len(executedFiles)+1)
	rolledBack = append(rolledBack, executedFiles...)
	if failedFile != "" {
		rolledBack = append(rolledBack, failedFile)
	}
	return &MockDataLoadError{
		PluginID:        pluginID,
		FailedFile:      failedFile,
		RolledBackFiles: rolledBack,
		Cause:           causeErr,
	}
}

// This file centralizes the optional install-time mock-data load helper used
// by both the source-plugin install path and the dynamic-plugin reconciler so
// the bizerr wrapping for a rolled-back load lives in one place.

package plugin

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

// MockDataExecutor abstracts the lifecycle method that executes mock SQL inside
// the caller's transaction. It accepts whatever subset of lifecycle.Service the
// helper actually depends on so tests can provide a thin in-memory stub.
type MockDataExecutor interface {
	// ExecuteManifestMockSQLFilesInTx executes mock-data SQL inside the caller-
	// supplied transaction (carried via ctx). On failure the caller MUST roll
	// back the surrounding transaction; the returned result enumerates the
	// files about to be reverted.
	ExecuteManifestMockSQLFilesInTx(
		ctx context.Context,
		manifest *catalog.Manifest,
	) lifecycle.MockSQLExecutionResult
}

// executeMockDataLoadTransaction wraps the caller-provided executor in a
// dao.SysPluginMigration.Transaction closure so any failure rolls back both
// the executed mock data rows and their migration ledger entries together.
// On rollback returns a *lifecycle.MockDataLoadError carrying the structured
// details. Returns nil on success or when there is no mock data to load.
func executeMockDataLoadTransaction(
	ctx context.Context,
	executor MockDataExecutor,
	manifest *catalog.Manifest,
) error {
	var (
		executedFiles []string
		failedFile    string
		causeErr      error
	)
	txErr := dao.SysPluginMigration.Transaction(ctx, func(txCtx context.Context, _ gdb.TX) error {
		result := executor.ExecuteManifestMockSQLFilesInTx(txCtx, manifest)
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
	logger.Warningf(
		ctx,
		"plugin mock data load rolled back plugin=%s failedFile=%s cause=%v",
		manifest.ID,
		failedFile,
		causeErr,
	)
	rolledBack := make([]string, 0, len(executedFiles)+1)
	rolledBack = append(rolledBack, executedFiles...)
	if failedFile != "" {
		rolledBack = append(rolledBack, failedFile)
	}
	return &lifecycle.MockDataLoadError{
		PluginID:        manifest.ID,
		FailedFile:      failedFile,
		RolledBackFiles: rolledBack,
		Cause:           causeErr,
	}
}

// wrapMockDataLoadError converts a lifecycle.MockDataLoadError into the stable
// user-facing bizerr that carries all parameters into i18n templates. Returns
// the original err unchanged when the chain does not contain a mock-data load
// error so callers can pass through arbitrary install errors safely.
func wrapMockDataLoadError(err error) error {
	if err == nil {
		return nil
	}
	var mockErr *lifecycle.MockDataLoadError
	if !errors.As(err, &mockErr) {
		return err
	}
	causeText := ""
	if mockErr.Cause != nil {
		causeText = mockErr.Cause.Error()
	}
	return bizerr.NewCode(
		CodePluginInstallMockDataFailed,
		bizerr.P("pluginId", mockErr.PluginID),
		bizerr.P("failedFile", mockErr.FailedFile),
		bizerr.P("rolledBackFiles", mockErr.RolledBackFiles),
		bizerr.P("cause", causeText),
	)
}

// isMockDataLoadError reports whether err represents an install that succeeded
// except for the optional mock-data load phase.
func isMockDataLoadError(err error) bool {
	var mockErr *lifecycle.MockDataLoadError
	return errors.As(err, &mockErr)
}

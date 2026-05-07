// Package cmd assembles the Lina core command tree together with shared
// helpers used by host bootstrap and maintenance operations.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/packed"
	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
)

// Main is the root command object for the Lina core CLI.
type Main struct {
	g.Meta `name:"main" root:"http"`
}

// Sensitive command names that require explicit confirmation.
const (
	initCommandName = "init"
	mockCommandName = "mock"
)

// sqlAssetSource identifies where init/mock should load SQL assets from.
type sqlAssetSource string

const (
	// sqlAssetSourceEmbedded loads SQL assets from the packaged embedded manifest.
	sqlAssetSourceEmbedded sqlAssetSource = "embedded"
	// sqlAssetSourceLocal loads SQL assets directly from the local source tree.
	sqlAssetSourceLocal sqlAssetSource = "local"
)

// hostManifestSQLDir is the canonical slash-based directory that stores host SQL delivery files.
const hostManifestSQLDir = "manifest/sql"

// sqlAsset stores one resolved SQL resource ready for execution.
type sqlAsset struct {
	Path    string // Path stores the logical or filesystem path used for logging.
	Content string // Content stores the SQL body to execute.
}

// sqlExecutor executes one SQL statement for shared command SQL processing.
type sqlExecutor func(ctx context.Context, sql string) error

// commandDatabase returns the database used by init/mock SQL execution.
var commandDatabase = func() gdb.DB {
	return g.DB()
}

// requireCommandConfirmation validates the explicit confirmation value for a
// sensitive command before any destructive step is executed.
func requireCommandConfirmation(commandName string, confirmValue string) error {
	expectedValue := expectedCommandConfirmation(commandName)
	if confirmValue == expectedValue {
		return nil
	}
	return gerror.Newf(
		"command %s performs sensitive upgrade or database operations and requires explicit confirmation; use %s or %s",
		commandName,
		makeConfirmationExample(commandName),
		goRunConfirmationExample(commandName),
	)
}

// expectedCommandConfirmation returns the required confirmation token for the
// given command.
func expectedCommandConfirmation(commandName string) string {
	return commandName
}

// makeConfirmationExample returns the safe make invocation for the command.
func makeConfirmationExample(commandName string) string {
	return fmt.Sprintf("make %s confirm=%s", commandName, expectedCommandConfirmation(commandName))
}

// goRunConfirmationExample returns the safe go run invocation for the command.
func goRunConfirmationExample(commandName string) string {
	return fmt.Sprintf("go run main.go %s --confirm=%s", commandName, expectedCommandConfirmation(commandName))
}

// hostInitSQLDir returns the conventional slash-based host SQL directory.
func hostInitSQLDir() string {
	return hostManifestSQLDir
}

// hostMockSQLDir returns the conventional slash-based host mock-data SQL directory.
func hostMockSQLDir() string {
	return path.Join(hostManifestSQLDir, "mock-data")
}

// resolveSQLAssetSource validates the requested SQL asset source and applies the
// runtime default when the caller does not explicitly specify one.
func resolveSQLAssetSource(value string) (sqlAssetSource, error) {
	normalized := sqlAssetSource(strings.TrimSpace(value))
	if normalized == "" {
		return sqlAssetSourceEmbedded, nil
	}
	switch normalized {
	case sqlAssetSourceEmbedded, sqlAssetSourceLocal:
		return normalized, nil
	default:
		return "", gerror.Newf("unsupported SQL asset source: %s; available values are embedded or local", value)
	}
}

// executeSQLAssets runs the provided SQL assets in order, splitting each file
// into executable statements and stopping immediately on the first failure.
func executeSQLAssets(ctx context.Context, assets []sqlAsset) error {
	link, err := currentDatabaseLink(ctx)
	if err != nil {
		return err
	}
	dbDialect, err := dialect.From(link)
	if err != nil {
		return err
	}
	return executeSQLAssetsWithExecutor(ctx, assets, func(ctx context.Context, sql string) error {
		_, err := commandDatabase().Exec(ctx, sql)
		return err
	}, dbDialect)
}

// executeSQLAssetsWithExecutor executes prepared SQL assets statement by
// statement through the provided executor, allowing unit tests to verify
// stop-on-error behavior without a real DB.
func executeSQLAssetsWithExecutor(
	ctx context.Context,
	assets []sqlAsset,
	executor sqlExecutor,
	dbDialect ...dialect.Dialect,
) error {
	var activeDialect dialect.Dialect
	if len(dbDialect) > 0 {
		activeDialect = dbDialect[0]
	}
	for _, asset := range assets {
		content := asset.Content
		if activeDialect != nil {
			translated, err := activeDialect.TranslateDDL(ctx, asset.Path, asset.Content)
			if err != nil {
				return gerror.Wrapf(err, "translate SQL file %s failed", asset.Path)
			}
			content = translated
		}
		statements := splitSQLStatements(content)
		if len(statements) == 0 {
			continue
		}
		baseName := sqlAssetBaseName(asset.Path)
		logger.Infof(ctx, "Executing SQL file: %s", baseName)
		for index, statement := range statements {
			if err := executor(ctx, statement); err != nil {
				statementIndex := index + 1
				logger.Warningf(ctx, "execute %s statement %d: %v", baseName, statementIndex, err)
				return gerror.Wrapf(err, "execute statement %d in SQL file %s failed", statementIndex, baseName)
			}
		}
	}
	return nil
}

// scanLocalSQLAssets loads SQL assets from one source-tree directory.
func scanLocalSQLAssets(ctx context.Context, slashDir string) ([]sqlAsset, error) {
	localDir := filepath.FromSlash(slashDir)
	if !gfile.Exists(localDir) {
		logger.Warningf(ctx, "SQL directory does not exist: %s", localDir)
		return nil, nil
	}
	files, err := gfile.ScanDirFile(localDir, "*.sql", false)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	assets := make([]sqlAsset, 0, len(files))
	for _, file := range files {
		assets = append(assets, sqlAsset{
			Path:    file,
			Content: gfile.GetContents(file),
		})
	}
	return assets, nil
}

// scanEmbeddedSQLAssets loads SQL assets from the packaged manifest embedded in the host binary.
func scanEmbeddedSQLAssets(ctx context.Context, slashDir string) ([]sqlAsset, error) {
	entries, err := fs.ReadDir(packed.Files, slashDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			logger.Warningf(ctx, "embedded SQL directory does not exist: %s", slashDir)
			return nil, nil
		}
		return nil, err
	}

	assets := make([]sqlAsset, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
			continue
		}
		assetPath := path.Join(slashDir, entry.Name())
		content, readErr := fs.ReadFile(packed.Files, assetPath)
		if readErr != nil {
			return nil, readErr
		}
		assets = append(assets, sqlAsset{
			Path:    assetPath,
			Content: string(content),
		})
	}
	sort.SliceStable(assets, func(i int, j int) bool {
		return assets[i].Path < assets[j].Path
	})
	return assets, nil
}

// sqlAssetBaseName returns the basename rendered in command logs and errors.
func sqlAssetBaseName(assetPath string) string {
	normalized := strings.TrimSpace(strings.ReplaceAll(assetPath, `\\`, "/"))
	if normalized == "" {
		return ""
	}
	parts := strings.Split(normalized, "/")
	return parts[len(parts)-1]
}

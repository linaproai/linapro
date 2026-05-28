// Package sqlassets scans and executes host SQL delivery assets for init and
// mock commands while keeping command entry files focused on CLI semantics.
package sqlassets

import (
	"context"
	"errors"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/cmd/internal/dbconfig"
	"lina-core/internal/packed"
	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
)

// Source identifies where init/mock should load SQL assets from.
type Source string

const (
	// SourceEmbedded loads SQL assets from the packaged embedded manifest.
	SourceEmbedded Source = "embedded"
	// SourceLocal loads SQL assets directly from the local source tree.
	SourceLocal Source = "local"
)

// Asset stores one resolved SQL resource ready for execution.
type Asset struct {
	Path    string // Path stores the logical or filesystem path used for logging.
	Content string // Content stores the SQL body to execute.
}

// Executor executes one SQL statement for shared command SQL processing.
type Executor func(ctx context.Context, sql string) error

// commandDatabase returns the database used by init/mock SQL execution.
var commandDatabase = func() gdb.DB {
	return g.DB()
}

// HostInitDir returns the conventional slash-based host SQL directory.
func HostInitDir() string {
	return hostManifestSQLDir
}

// HostMockDir returns the conventional slash-based host mock-data SQL directory.
func HostMockDir() string {
	return path.Join(hostManifestSQLDir, "mock-data")
}

// ResolveSource validates the requested SQL asset source and applies the
// runtime default when the caller does not explicitly specify one.
func ResolveSource(value string) (Source, error) {
	normalized := Source(strings.TrimSpace(value))
	if normalized == "" {
		return SourceEmbedded, nil
	}
	switch normalized {
	case SourceEmbedded, SourceLocal:
		return normalized, nil
	default:
		return "", gerror.Newf("unsupported SQL asset source: %s; available values are embedded or local", value)
	}
}

// ScanInit loads host initialization SQL assets from the selected source.
func ScanInit(ctx context.Context, source Source) ([]Asset, error) {
	switch source {
	case SourceLocal:
		return scanLocal(ctx, HostInitDir())
	case SourceEmbedded:
		return scanEmbedded(ctx, HostInitDir())
	default:
		return nil, gerror.Newf("unsupported init SQL asset source: %s", source)
	}
}

// ScanMock loads host mock-data SQL assets from the selected source.
func ScanMock(ctx context.Context, source Source) ([]Asset, error) {
	switch source {
	case SourceLocal:
		return scanLocal(ctx, HostMockDir())
	case SourceEmbedded:
		return scanEmbedded(ctx, HostMockDir())
	default:
		return nil, gerror.Newf("unsupported mock SQL asset source: %s", source)
	}
}

// Execute runs the provided SQL assets in order, splitting each file into
// executable statements and stopping immediately on the first failure.
func Execute(ctx context.Context, assets []Asset) error {
	link, err := dbconfig.CurrentDatabaseLink(ctx)
	if err != nil {
		return err
	}
	dbDialect, err := dialect.From(link)
	if err != nil {
		return err
	}
	return ExecuteWithExecutor(ctx, assets, func(ctx context.Context, sql string) error {
		_, err := commandDatabase().Exec(ctx, sql)
		return err
	}, dbDialect)
}

// ExecuteWithExecutor executes prepared SQL assets statement by statement
// through the provided executor, allowing unit tests to verify stop-on-error
// behavior without a real DB.
func ExecuteWithExecutor(
	ctx context.Context,
	assets []Asset,
	executor Executor,
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
		statements := Split(content)
		if len(statements) == 0 {
			continue
		}
		baseName := baseName(asset.Path)
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

// Split parses one SQL file content into ordered executable statements while
// ignoring semicolons inside strings and comments.
func Split(content string) []string {
	return dialect.SplitSQLStatements(content)
}

const hostManifestSQLDir = "manifest/sql"

// scanLocal loads SQL assets from one source-tree directory.
func scanLocal(ctx context.Context, slashDir string) ([]Asset, error) {
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

	assets := make([]Asset, 0, len(files))
	for _, file := range files {
		assets = append(assets, Asset{
			Path:    file,
			Content: gfile.GetContents(file),
		})
	}
	return assets, nil
}

// scanEmbedded loads SQL assets from the packaged manifest embedded in the host binary.
func scanEmbedded(ctx context.Context, slashDir string) ([]Asset, error) {
	entries, err := fs.ReadDir(packed.Files, slashDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			logger.Warningf(ctx, "embedded SQL directory does not exist: %s", slashDir)
			return nil, nil
		}
		return nil, err
	}

	assets := make([]Asset, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
			continue
		}
		assetPath := path.Join(slashDir, entry.Name())
		content, readErr := fs.ReadFile(packed.Files, assetPath)
		if readErr != nil {
			return nil, readErr
		}
		assets = append(assets, Asset{
			Path:    assetPath,
			Content: string(content),
		})
	}
	sort.SliceStable(assets, func(i int, j int) bool {
		return assets[i].Path < assets[j].Path
	})
	return assets, nil
}

// baseName returns the basename rendered in command logs and errors.
func baseName(assetPath string) string {
	normalized := strings.TrimSpace(strings.ReplaceAll(assetPath, `\\`, "/"))
	if normalized == "" {
		return ""
	}
	parts := strings.Split(normalized, "/")
	return parts[len(parts)-1]
}

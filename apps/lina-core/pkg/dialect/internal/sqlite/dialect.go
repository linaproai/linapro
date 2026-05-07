// Package sqlite implements LinaPro's internal SQLite dialect behavior.
package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/pkg/logger"
)

// Name is the stable SQLite dialect name.
const Name = "sqlite"

// StartupRuntime is the narrow startup configuration interface needed by the
// SQLite startup hook.
type StartupRuntime interface {
	// OverrideClusterEnabledForDialect locks cluster.enabled in memory for the
	// current process when SQLite cannot support cluster mode.
	OverrideClusterEnabledForDialect(value bool)
}

// TranslateDDL converts the project's MySQL-source SQL subset to SQLite SQL.
func TranslateDDL(ctx context.Context, sourceName string, ddl string) (string, error) {
	return translateDDL(sourceName, ddl)
}

// PrepareDatabase ensures the database parent directory exists and optionally
// deletes existing SQLite database files when rebuild is requested.
func PrepareDatabase(ctx context.Context, link string, rebuild bool) error {
	dbPath, err := PathFromLink(link)
	if err != nil {
		return err
	}
	if strings.HasPrefix(dbPath, "~") {
		return gerror.Newf("SQLite path %s is not supported; use an absolute path or a working-directory relative path", dbPath)
	}

	parentDir := gfile.Dir(dbPath)
	if parentDir == "" || parentDir == "." {
		parentDir = "."
	}
	if rebuild {
		logger.Warningf(ctx, "rebuilding SQLite database %s: deleting existing database files", dbPath)
		for _, path := range DatabaseFiles(dbPath) {
			if gfile.Exists(path) {
				if err = gfile.Remove(path); err != nil {
					return gerror.Wrapf(err, "remove SQLite database file %s before rebuild failed", path)
				}
			}
		}
	}
	if err = gfile.Mkdir(parentDir); err != nil {
		return gerror.Wrapf(err, "create SQLite database parent directory %s failed", parentDir)
	}
	return nil
}

// SupportsCluster reports that SQLite cannot back multi-node coordination.
func SupportsCluster() bool {
	return false
}

// OnStartup locks cluster mode off and prints prominent warnings for SQLite.
func OnStartup(ctx context.Context, link string, runtime StartupRuntime) error {
	if runtime != nil {
		runtime.OverrideClusterEnabledForDialect(false)
	}
	linkText := link
	if linkText == "" {
		linkText = "sqlite::<unknown>"
	}
	logger.Warningf(ctx, "[WARNING] SQLite mode is active (database.default.link = %s)", linkText)
	logger.Warning(ctx, "[WARNING] SQLite mode only supports single-node deployment; cluster.enabled has been forced to false")
	logger.Warning(ctx, "[WARNING] All features run in single-node mode; do not use SQLite mode in production")
	logger.Warning(ctx, "[WARNING] Switch database.default.link back to a MySQL link for multi-node deployment")
	return nil
}

// PathFromLink extracts the database file path from a GoFrame SQLite link.
func PathFromLink(link string) (string, error) {
	normalized := strings.TrimSpace(link)
	if !strings.HasPrefix(normalized, "sqlite:") {
		return "", gerror.New("SQLite link must start with sqlite:")
	}
	path := strings.TrimSpace(strings.TrimPrefix(normalized, "sqlite:"))
	path = strings.TrimPrefix(path, ":")
	path = strings.TrimSpace(path)
	if path == "" {
		return "", gerror.New("SQLite database path is missing from database link")
	}
	if strings.HasPrefix(path, "@file(") && strings.HasSuffix(path, ")") {
		path = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(path, "@file("), ")"))
		if path == "" {
			return "", gerror.New("SQLite database path is missing from database link")
		}
		return path, nil
	}
	return "", gerror.Newf(
		"SQLite link must use GoFrame file syntax sqlite::@file(path), got %s",
		fmt.Sprintf("sqlite:%s", path),
	)
}

// DatabaseFiles returns the primary SQLite file and common WAL sidecar files.
func DatabaseFiles(dbPath string) []string {
	return []string{dbPath, dbPath + "-shm", dbPath + "-wal"}
}

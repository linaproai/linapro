// This file implements the host database initialization command that scans the
// conventional manifest SQL directories.

package cmd

import (
	"context"
	"sort"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/pkg/logger"
)

// InitInput defines the command-line options for the sensitive database
// initialization command.
type InitInput struct {
	g.Meta  `name:"init" brief:"initialize database schema and seed data (DDL + seed DML), requires --confirm=init"`
	Confirm string `name:"confirm" brief:"explicit confirmation value, must be 'init'"`
}

// InitOutput carries the command result placeholder.
type InitOutput struct{}

// Init initializes host and plugin SQL resources after an explicit safety
// confirmation is provided.
func (m *Main) Init(ctx context.Context, in InitInput) (out *InitOutput, err error) {
	if err = requireCommandConfirmation(initCommandName, in.Confirm); err != nil {
		return nil, err
	}
	files, err := scanInitSqlFiles(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "扫描初始化 SQL 文件失败")
	}
	if len(files) == 0 {
		logger.Warning(ctx, "no SQL files found for initialization")
		return
	}
	sort.Strings(files)
	if err = executeSQLFiles(ctx, files); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Database initialization completed.")
	return
}

type sqlExecutor func(ctx context.Context, sql string) error

// executeSQLFiles runs the provided SQL files in order and stops immediately on
// the first execution failure.
func executeSQLFiles(ctx context.Context, files []string) error {
	return executeSQLFilesWithExecutor(ctx, files, func(ctx context.Context, sql string) error {
		_, err := g.DB().Exec(ctx, sql)
		return err
	})
}

// executeSQLFilesWithExecutor reads SQL files and delegates execution to the
// provided executor, which allows unit tests to verify stop-on-error behavior
// without touching a real database.
func executeSQLFilesWithExecutor(ctx context.Context, files []string, executor sqlExecutor) error {
	for _, file := range files {
		sql := gfile.GetContents(file)
		if sql == "" {
			continue
		}
		logger.Infof(ctx, "Executing SQL file: %s", gfile.Basename(file))
		if err := executor(ctx, sql); err != nil {
			logger.Warningf(ctx, "execute %s: %v", gfile.Basename(file), err)
			return gerror.Wrapf(err, "执行 SQL 文件 %s 失败", gfile.Basename(file))
		}
	}
	return nil
}

// scanInitSqlFiles scans the conventional host and plugin initialization SQL
// directories.
func scanInitSqlFiles(ctx context.Context) ([]string, error) {
	var (
		files      = make([]string, 0)
		pluginRoot = gfile.RealPath(gfile.Join("..", "lina-plugins"))
		sqlDir     = hostInitSqlDir()
	)

	if gfile.Exists(sqlDir) {
		coreFiles, err := gfile.ScanDirFile(sqlDir, "*.sql", false)
		if err != nil {
			return nil, err
		}
		files = append(files, coreFiles...)
	} else {
		logger.Warningf(ctx, "SQL directory does not exist: %s", sqlDir)
	}

	if pluginRoot == "" || !gfile.Exists(pluginRoot) {
		return files, nil
	}

	pluginEntries, err := gfile.ScanDir(pluginRoot, "*", false)
	if err != nil {
		return nil, err
	}
	for _, pluginPath := range pluginEntries {
		pluginSqlDir := gfile.Join(pluginPath, "manifest", "sql")
		if !gfile.Exists(pluginSqlDir) {
			continue
		}
		pluginFiles, scanErr := gfile.ScanDirFile(pluginSqlDir, "*.sql", false)
		if scanErr != nil {
			return nil, scanErr
		}
		files = append(files, pluginFiles...)
	}

	return files, nil
}

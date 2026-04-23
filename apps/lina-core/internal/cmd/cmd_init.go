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

// Init initializes host SQL resources after an explicit safety confirmation is
// provided.
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

// scanInitSqlFiles scans the conventional host initialization SQL directory.
func scanInitSqlFiles(ctx context.Context) ([]string, error) {
	var (
		files  = make([]string, 0)
		sqlDir = hostInitSqlDir()
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

	return files, nil
}

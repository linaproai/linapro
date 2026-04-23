// This file implements the host database initialization command with explicit
// SQL asset source selection for development and runtime execution.

package cmd

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/logger"
)

// InitInput defines the command-line options for the sensitive database
// initialization command.
type InitInput struct {
	g.Meta    `name:"init" brief:"initialize database schema and seed data (DDL + seed DML), requires --confirm=init"`
	Confirm   string `name:"confirm" brief:"explicit confirmation value, must be 'init'"`
	SQLSource string `name:"sql-source" brief:"SQL asset source: embedded or local; defaults to embedded"`
}

// InitOutput carries the command result placeholder.
type InitOutput struct{}

// Init initializes host SQL resources after an explicit safety confirmation is
// provided.
func (m *Main) Init(ctx context.Context, in InitInput) (out *InitOutput, err error) {
	if err = requireCommandConfirmation(initCommandName, in.Confirm); err != nil {
		return nil, err
	}
	source, err := resolveSQLAssetSource(in.SQLSource)
	if err != nil {
		return nil, err
	}
	assets, err := scanInitSQLAssets(ctx, source)
	if err != nil {
		return nil, gerror.Wrap(err, "扫描初始化 SQL 文件失败")
	}
	if len(assets) == 0 {
		logger.Warning(ctx, "no SQL files found for initialization")
		return
	}
	if err = executeSQLAssets(ctx, assets); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Database initialization completed.")
	return
}

// scanInitSQLAssets loads host initialization SQL assets from the selected source.
func scanInitSQLAssets(ctx context.Context, source sqlAssetSource) ([]sqlAsset, error) {
	switch source {
	case sqlAssetSourceLocal:
		return scanLocalSQLAssets(ctx, hostInitSQLDir())
	case sqlAssetSourceEmbedded:
		return scanEmbeddedSQLAssets(ctx, hostInitSQLDir())
	default:
		return nil, gerror.Newf("不支持的 init SQL 资源来源: %s", source)
	}
}

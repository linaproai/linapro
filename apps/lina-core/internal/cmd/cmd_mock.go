// This file implements the host mock-data loading command with explicit SQL
// asset source selection for development and runtime execution.

package cmd

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/logger"
)

// MockInput defines the command-line options for the sensitive mock-data
// loading command.
type MockInput struct {
	g.Meta    `name:"mock" brief:"load mock/demo data from manifest/sql/mock-data/, requires --confirm=mock"`
	Confirm   string `name:"confirm" brief:"explicit confirmation value, must be 'mock'"`
	SQLSource string `name:"sql-source" brief:"SQL asset source: embedded or local; defaults to embedded"`
}

// MockOutput carries the command result placeholder.
type MockOutput struct{}

// Mock loads host mock SQL resources after an explicit safety confirmation is
// provided.
func (m *Main) Mock(ctx context.Context, in MockInput) (out *MockOutput, err error) {
	if err = requireCommandConfirmation(mockCommandName, in.Confirm); err != nil {
		return nil, err
	}
	source, err := resolveSQLAssetSource(in.SQLSource)
	if err != nil {
		return nil, err
	}
	assets, err := scanMockSQLAssets(ctx, source)
	if err != nil {
		return nil, gerror.Wrap(err, "scan mock SQL files failed")
	}
	if len(assets) == 0 {
		logger.Warning(ctx, "no mock SQL files found")
		return
	}
	if err = executeSQLAssets(ctx, assets); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Mock data loaded.")
	return
}

// scanMockSQLAssets loads host mock-data SQL assets from the selected source.
func scanMockSQLAssets(ctx context.Context, source sqlAssetSource) ([]sqlAsset, error) {
	switch source {
	case sqlAssetSourceLocal:
		return scanLocalSQLAssets(ctx, hostMockSQLDir())
	case sqlAssetSourceEmbedded:
		return scanEmbeddedSQLAssets(ctx, hostMockSQLDir())
	default:
		return nil, gerror.Newf("unsupported mock SQL asset source: %s", source)
	}
}

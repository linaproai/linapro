// This file implements the host mock-data loading command that scans the
// conventional mock SQL directories.

package cmd

import (
	"context"
	"sort"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/pkg/logger"
)

// MockInput defines the command-line options for the sensitive mock-data
// loading command.
type MockInput struct {
	g.Meta  `name:"mock" brief:"load mock/demo data from manifest/sql/mock-data/, requires --confirm=mock"`
	Confirm string `name:"confirm" brief:"explicit confirmation value, must be 'mock'"`
}

// MockOutput carries the command result placeholder.
type MockOutput struct{}

// Mock loads host and plugin mock SQL resources after an explicit safety
// confirmation is provided.
func (m *Main) Mock(ctx context.Context, in MockInput) (out *MockOutput, err error) {
	if err = requireCommandConfirmation(mockCommandName, in.Confirm); err != nil {
		return nil, err
	}
	files, err := scanMockSqlFiles(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "扫描 Mock SQL 文件失败")
	}
	if len(files) == 0 {
		logger.Warning(ctx, "no mock SQL files found")
		return
	}
	sort.Strings(files)
	if err = executeSQLFiles(ctx, files); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Mock data loaded.")
	return
}

// scanMockSqlFiles scans the conventional host and plugin mock-data SQL
// directories.
func scanMockSqlFiles(ctx context.Context) ([]string, error) {
	var (
		files      = make([]string, 0)
		mockDir    = hostMockSqlDir()
		pluginRoot = gfile.RealPath(gfile.Join("..", "lina-plugins"))
	)

	if gfile.Exists(mockDir) {
		coreFiles, err := gfile.ScanDirFile(mockDir, "*.sql", false)
		if err != nil {
			return nil, err
		}
		files = append(files, coreFiles...)
	} else {
		logger.Warningf(ctx, "mock-data directory does not exist: %s", mockDir)
	}

	if pluginRoot == "" || !gfile.Exists(pluginRoot) {
		return files, nil
	}

	pluginEntries, err := gfile.ScanDir(pluginRoot, "*", false)
	if err != nil {
		return nil, err
	}
	for _, pluginPath := range pluginEntries {
		pluginMockDir := gfile.Join(pluginPath, "manifest", "mock-data")
		if !gfile.Exists(pluginMockDir) {
			continue
		}
		pluginFiles, scanErr := gfile.ScanDirFile(pluginMockDir, "*.sql", false)
		if scanErr != nil {
			return nil, scanErr
		}
		files = append(files, pluginFiles...)
	}

	return files, nil
}

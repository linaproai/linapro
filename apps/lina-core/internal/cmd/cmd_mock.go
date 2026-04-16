package cmd

import (
	"context"
	"sort"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/service/config"
	"lina-core/pkg/logger"
)

type MockInput struct {
	g.Meta `name:"mock" brief:"load mock/demo data from manifest/sql/mock-data/"`
}
type MockOutput struct{}

func (m *Main) Mock(ctx context.Context, in MockInput) (out *MockOutput, err error) {
	sqlDir := config.New().GetInit(ctx).SqlDir
	files, err := scanMockSqlFiles(ctx, sqlDir)
	if err != nil {
		logger.Warningf(ctx, "failed to scan mock SQL files: %v", err)
		return nil, nil
	}
	if len(files) == 0 {
		logger.Warning(ctx, "no mock SQL files found")
		return
	}
	sort.Strings(files)
	execSqlFiles(ctx, files)

	logger.Info(ctx, "Mock data loaded.")
	return
}

func scanMockSqlFiles(ctx context.Context, sqlDir string) ([]string, error) {
	var (
		files      = make([]string, 0)
		mockDir    = gfile.Join(sqlDir, "mock-data")
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

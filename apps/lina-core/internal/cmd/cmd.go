// Package cmd assembles the Lina core command tree together with shared
// helpers used by host bootstrap and maintenance operations.
package cmd

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/pkg/logger"
)

// Main is the root command object for the Lina core CLI.
type Main struct {
	g.Meta `name:"main" root:"http"`
}

// Sensitive command names that require explicit confirmation.
const (
	initCommandName    = "init"
	mockCommandName    = "mock"
	upgradeCommandName = "upgrade"
)

// hostManifestSqlDir is the canonical directory that stores host SQL delivery files.
const hostManifestSqlDir = "manifest/sql"

// sqlExecutor executes one SQL statement for shared command SQL processing.
type sqlExecutor func(ctx context.Context, sql string) error

// requireCommandConfirmation validates the explicit confirmation value for a
// sensitive command before any destructive step is executed.
func requireCommandConfirmation(commandName string, confirmValue string) error {
	expectedValue := expectedCommandConfirmation(commandName)
	if confirmValue == expectedValue {
		return nil
	}
	return gerror.Newf(
		"命令 %s 涉及敏感升级或数据库操作，必须显式确认后才能执行。请使用 %s 或 %s",
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

// hostInitSqlDir returns the conventional host SQL directory.
func hostInitSqlDir() string {
	return hostManifestSqlDir
}

// hostMockSqlDir returns the conventional host mock-data SQL directory.
func hostMockSqlDir() string {
	return gfile.Join(hostManifestSqlDir, "mock-data")
}

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

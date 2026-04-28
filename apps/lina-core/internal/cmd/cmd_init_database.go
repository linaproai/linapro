// This file prepares the configured database before host SQL assets are
// executed by the init command.

package cmd

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/logger"
)

// initDatabaseName is the canonical local framework database initialized by make init.
const initDatabaseName = "linapro"

// parseInitRebuildFlag parses the optional init rebuild flag while treating an
// omitted value as a non-destructive initialization.
func parseInitRebuildFlag(value string) (bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "", "false", "0", "no", "n", "off":
		return false, nil
	case "true", "1", "yes", "y", "on":
		return true, nil
	default:
		return false, gerror.Newf("不支持的 rebuild 参数值: %s，可选值为 true 或 false", value)
	}
}

// prepareInitDatabase creates the canonical database before executing SQL
// assets and drops it first when rebuild is explicitly enabled.
func prepareInitDatabase(ctx context.Context, rebuild bool) (err error) {
	linkVar, err := g.Cfg().Get(ctx, "database.default.link")
	if err != nil {
		return gerror.Wrap(err, "读取数据库连接配置失败")
	}
	if linkVar == nil {
		return gerror.New("数据库连接配置 database.default.link 不能为空")
	}
	link := strings.TrimSpace(linkVar.String())
	if link == "" {
		return gerror.New("数据库连接配置 database.default.link 不能为空")
	}

	databaseName, err := databaseNameFromMySQLLink(link)
	if err != nil {
		return err
	}
	if databaseName != initDatabaseName {
		return gerror.Newf("初始化数据库连接串必须指向 %s，当前为 %s", initDatabaseName, databaseName)
	}

	serverLink, err := serverLinkFromMySQLLink(link)
	if err != nil {
		return err
	}
	quotedName, err := quoteMySQLIdentifier(databaseName)
	if err != nil {
		return err
	}

	db, err := gdb.New(gdb.ConfigNode{Link: serverLink})
	if err != nil {
		return gerror.Wrap(err, "创建数据库初始化连接失败")
	}
	defer func() {
		if closeErr := db.Close(ctx); closeErr != nil && err == nil {
			err = gerror.Wrap(closeErr, "关闭数据库初始化连接失败")
		}
	}()

	if rebuild {
		logger.Warningf(ctx, "rebuilding database %s: dropping existing schema", databaseName)
		if _, err = db.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
			return gerror.Wrapf(err, "重建数据库前删除 %s 失败", databaseName)
		}
	}

	createDatabaseSQL := "CREATE DATABASE IF NOT EXISTS " + quotedName +
		" DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"
	if _, err = db.Exec(ctx, createDatabaseSQL); err != nil {
		return gerror.Wrapf(err, "创建数据库 %s 失败", databaseName)
	}
	return nil
}

// databaseNameFromMySQLLink extracts the configured database name from a
// GoFrame MySQL link string.
func databaseNameFromMySQLLink(link string) (string, error) {
	_, name, _, err := splitMySQLLinkDatabase(link)
	if err != nil {
		return "", err
	}
	return name, nil
}

// serverLinkFromMySQLLink removes the database path segment while preserving
// connection parameters so init can create the target database before use.
func serverLinkFromMySQLLink(link string) (string, error) {
	prefix, _, query, err := splitMySQLLinkDatabase(link)
	if err != nil {
		return "", err
	}
	return prefix + query, nil
}

// splitMySQLLinkDatabase separates one GoFrame MySQL link into the connection
// prefix, database name, and query string.
func splitMySQLLinkDatabase(link string) (prefix string, name string, query string, err error) {
	normalized := strings.TrimSpace(link)
	pathPart := normalized
	if queryIndex := strings.Index(normalized, "?"); queryIndex >= 0 {
		pathPart = normalized[:queryIndex]
		query = normalized[queryIndex:]
	}
	nameStart := strings.LastIndex(pathPart, "/")
	if nameStart < 0 {
		return "", "", "", gerror.New("数据库连接串缺少数据库名称")
	}
	name = strings.TrimSpace(pathPart[nameStart+1:])
	if name == "" {
		return "", "", "", gerror.New("数据库连接串缺少数据库名称")
	}
	return pathPart[:nameStart+1], name, query, nil
}

// quoteMySQLIdentifier safely quotes one MySQL identifier for database-level
// bootstrap statements.
func quoteMySQLIdentifier(identifier string) (string, error) {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return "", gerror.New("MySQL 标识符不能为空")
	}
	if strings.ContainsRune(trimmed, 0) {
		return "", gerror.New("MySQL 标识符不能包含空字符")
	}
	return "`" + strings.ReplaceAll(trimmed, "`", "``") + "`", nil
}

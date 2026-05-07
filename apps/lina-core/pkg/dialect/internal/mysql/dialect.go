// Package mysql implements LinaPro's internal MySQL dialect behavior.
package mysql

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
)

// Name is the stable MySQL dialect name.
const Name = "mysql"

// TranslateDDL leaves MySQL-source SQL unchanged.
func TranslateDDL(ctx context.Context, sourceName string, ddl string) (string, error) {
	return ddl, nil
}

// PrepareDatabase creates the configured MySQL database and optionally drops it
// first when rebuild is explicitly requested.
func PrepareDatabase(ctx context.Context, link string, rebuild bool) (err error) {
	configNode, err := ConfigNodeFromLink(link)
	if err != nil {
		return err
	}
	databaseName := strings.TrimSpace(configNode.Name)

	quotedName, err := QuoteIdentifier(databaseName)
	if err != nil {
		return err
	}

	serverNode := *configNode
	serverNode.Link = ""
	serverNode.Name = ""
	db, err := gdb.New(serverNode)
	if err != nil {
		return gerror.Wrap(err, "create database initialization connection failed")
	}
	defer func() {
		if closeErr := db.Close(ctx); closeErr != nil && err == nil {
			err = gerror.Wrap(closeErr, "close database initialization connection failed")
		}
	}()

	if rebuild {
		logger.Warningf(ctx, "rebuilding database %s: dropping existing schema", databaseName)
		if _, err = db.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
			return gerror.Wrapf(err, "drop database %s before rebuild failed", databaseName)
		}
	}

	createDatabaseSQL := "CREATE DATABASE IF NOT EXISTS " + quotedName +
		" DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"
	if _, err = db.Exec(ctx, createDatabaseSQL); err != nil {
		return gerror.Wrapf(err, "create database %s failed", databaseName)
	}
	return nil
}

// SupportsCluster reports that MySQL can back shared multi-node coordination tables.
func SupportsCluster() bool {
	return true
}

// ConfigNodeFromLink returns the GoFrame-parsed MySQL configuration node.
func ConfigNodeFromLink(link string) (*gdb.ConfigNode, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: link})
	if err != nil {
		return nil, gerror.Wrap(err, "parse MySQL database link failed")
	}
	configNode := db.GetConfig()
	if configNode == nil {
		return nil, gerror.New("database link configuration is empty")
	}
	node := *configNode
	if strings.TrimSpace(node.Name) == "" {
		return nil, gerror.New("database name is missing from database link")
	}
	return &node, nil
}

// QuoteIdentifier safely quotes one MySQL identifier for database-level
// bootstrap statements.
func QuoteIdentifier(identifier string) (string, error) {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return "", gerror.New("MySQL identifier must not be empty")
	}
	if strings.ContainsRune(trimmed, 0) {
		return "", gerror.New("MySQL identifier must not contain NUL bytes")
	}
	return "`" + strings.ReplaceAll(trimmed, "`", "``") + "`", nil
}

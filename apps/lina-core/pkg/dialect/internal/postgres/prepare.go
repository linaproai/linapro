// This file prepares PostgreSQL databases for init SQL execution.

package postgres

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
)

// systemDatabaseName is the PostgreSQL maintenance database used for DROP and
// CREATE DATABASE operations.
const systemDatabaseName = "postgres"

// PrepareDatabase creates the configured PostgreSQL database and optionally
// drops it first when rebuild is explicitly requested.
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

	systemNode := *configNode
	systemNode.Link = ""
	systemNode.Name = systemDatabaseName
	systemDB, err := gdb.New(systemNode)
	if err != nil {
		return gerror.Wrapf(
			err,
			"connect PostgreSQL system database %s at %s:%s as %s failed; PostgreSQL may not be ready",
			systemDatabaseName,
			systemNode.Host,
			systemNode.Port,
			systemNode.User,
		)
	}
	defer func() {
		if closeErr := systemDB.Close(ctx); closeErr != nil && err == nil {
			err = gerror.Wrap(closeErr, "close PostgreSQL database initialization connection failed")
		}
	}()

	exists, err := databaseExists(ctx, systemDB, databaseName)
	if err != nil {
		return gerror.Wrapf(err, "check PostgreSQL database %s existence failed", databaseName)
	}
	if rebuild && exists {
		logger.Warningf(ctx, "rebuilding PostgreSQL database %s: terminating active connections and dropping database", databaseName)
		if _, err = systemDB.Exec(
			ctx,
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()",
			databaseName,
		); err != nil {
			return gerror.Wrapf(err, "terminate active PostgreSQL connections for database %s failed", databaseName)
		}
		if _, err = systemDB.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
			return gerror.Wrapf(err, "drop PostgreSQL database %s before rebuild failed", databaseName)
		}
		exists = false
	}
	if exists {
		return nil
	}
	createDatabaseSQL := "CREATE DATABASE " + quotedName + " ENCODING 'UTF8' LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0"
	if _, err = systemDB.Exec(ctx, createDatabaseSQL); err != nil {
		return gerror.Wrapf(err, "create PostgreSQL database %s failed", databaseName)
	}
	return nil
}

// ConfigNodeFromLink returns the GoFrame-parsed PostgreSQL configuration node.
func ConfigNodeFromLink(link string) (*gdb.ConfigNode, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: link})
	if err != nil {
		return nil, gerror.Wrap(err, "parse PostgreSQL database link failed")
	}
	configNode := db.GetConfig()
	if configNode == nil {
		return nil, gerror.New("database link configuration is empty")
	}
	node := *configNode
	if strings.TrimSpace(node.Name) == "" {
		return nil, gerror.New("database name is missing from PostgreSQL database link")
	}
	return &node, nil
}

// QuoteIdentifier safely quotes one PostgreSQL identifier for database-level
// bootstrap statements.
func QuoteIdentifier(identifier string) (string, error) {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return "", gerror.New("PostgreSQL identifier must not be empty")
	}
	if strings.ContainsRune(trimmed, 0) {
		return "", gerror.New("PostgreSQL identifier must not contain NUL bytes")
	}
	return `"` + strings.ReplaceAll(trimmed, `"`, `""`) + `"`, nil
}

// databaseExists reports whether a PostgreSQL database exists in pg_database.
func databaseExists(ctx context.Context, db gdb.DB, databaseName string) (bool, error) {
	value, err := db.GetValue(ctx, "SELECT 1 FROM pg_database WHERE datname=$1", databaseName)
	if err != nil {
		return false, err
	}
	return !value.IsNil(), nil
}

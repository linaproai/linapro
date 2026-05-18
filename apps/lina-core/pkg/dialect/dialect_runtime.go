// This file implements database dialect resolution and concrete dialect adapters.

package dialect

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	internalpostgres "lina-core/pkg/dialect/internal/postgres"
	internalsqlite "lina-core/pkg/dialect/internal/sqlite"
)

// fromLink resolves one database dialect from the database.default.link prefix.
func fromLink(link string) (Dialect, error) {
	normalized := strings.TrimSpace(link)
	if normalized == "" {
		return nil, bizerr.NewCode(CodeDialectLinkRequired)
	}
	switch {
	case strings.HasPrefix(normalized, pgsqlPrefix):
		return postgresDialect{}, nil
	case strings.HasPrefix(normalized, sqlitePrefix):
		return sqliteDialect{link: normalized}, nil
	case strings.HasPrefix(normalized, "mysql:"):
		return nil, bizerr.NewCode(CodeDialectMySQLUnsupported)
	default:
		prefix := normalized
		if index := strings.Index(prefix, ":"); index >= 0 {
			prefix = prefix[:index+1]
		}
		return nil, bizerr.NewCode(CodeDialectUnsupported, bizerr.P("prefix", prefix))
	}
}

// databaseVersion returns a display-ready database engine and version label for
// the active GoFrame database.
func databaseVersion(ctx context.Context, db gdb.DB) (string, error) {
	dbDialect, err := FromDatabase(db)
	if err != nil {
		return "", err
	}
	return dbDialect.DatabaseVersion(ctx, db)
}

// postgresDialect is the public package wrapper for the internal PostgreSQL dialect.
type postgresDialect struct{}

// Name returns the stable PostgreSQL dialect name.
func (postgresDialect) Name() string {
	return internalpostgres.Name
}

// TranslateDDL leaves PostgreSQL-source SQL unchanged.
func (postgresDialect) TranslateDDL(ctx context.Context, sourceName string, ddl string) (string, error) {
	return internalpostgres.TranslateDDL(ctx, sourceName, ddl)
}

// PrepareDatabase creates the configured PostgreSQL database before init SQL runs.
func (postgresDialect) PrepareDatabase(ctx context.Context, link string, rebuild bool) error {
	return internalpostgres.PrepareDatabase(ctx, link, rebuild)
}

// SupportsCluster reports whether PostgreSQL can back cluster coordination tables.
func (postgresDialect) SupportsCluster() bool {
	return internalpostgres.SupportsCluster()
}

// DatabaseVersion returns the PostgreSQL server version label.
func (postgresDialect) DatabaseVersion(ctx context.Context, db gdb.DB) (string, error) {
	return internalpostgres.DatabaseVersion(ctx, db)
}

// QueryTableMetadata returns existing PostgreSQL table names and comments.
func (postgresDialect) QueryTableMetadata(ctx context.Context, db gdb.DB, schema string, tableNames []string) ([]TableMeta, error) {
	metas, err := internalpostgres.QueryTableMetadata(ctx, db, schema, tableNames)
	if err != nil {
		return nil, err
	}
	result := make([]TableMeta, 0, len(metas))
	for _, meta := range metas {
		result = append(result, TableMeta{
			TableName:    meta.TableName,
			TableComment: meta.TableComment,
		})
	}
	return result, nil
}

// OnStartup has no PostgreSQL-specific startup side effects.
func (postgresDialect) OnStartup(ctx context.Context, runtime RuntimeConfig) error {
	return internalpostgres.OnStartup(ctx, runtime)
}

// sqliteDialect is the public package wrapper for the internal SQLite dialect.
type sqliteDialect struct {
	link string // link stores the source database link for startup diagnostics.
}

// Name returns the stable SQLite dialect name.
func (sqliteDialect) Name() string {
	return internalsqlite.Name
}

// TranslateDDL converts the project's PostgreSQL-source SQL subset to SQLite SQL.
func (sqliteDialect) TranslateDDL(ctx context.Context, sourceName string, ddl string) (string, error) {
	return internalsqlite.TranslateDDL(ctx, sourceName, ddl)
}

// PrepareDatabase prepares the SQLite database file before init SQL runs.
func (sqliteDialect) PrepareDatabase(ctx context.Context, link string, rebuild bool) error {
	return internalsqlite.PrepareDatabase(ctx, link, rebuild)
}

// SupportsCluster reports whether SQLite can back cluster coordination tables.
func (sqliteDialect) SupportsCluster() bool {
	return internalsqlite.SupportsCluster()
}

// DatabaseVersion returns the SQLite library version label.
func (sqliteDialect) DatabaseVersion(ctx context.Context, db gdb.DB) (string, error) {
	return internalsqlite.DatabaseVersion(ctx, db)
}

// QueryTableMetadata returns SQLite table names with empty comments.
func (sqliteDialect) QueryTableMetadata(ctx context.Context, db gdb.DB, schema string, tableNames []string) ([]TableMeta, error) {
	metas, err := internalsqlite.QueryTableMetadata(ctx, db, schema, tableNames)
	if err != nil {
		return nil, err
	}
	result := make([]TableMeta, 0, len(metas))
	for _, meta := range metas {
		result = append(result, TableMeta{
			TableName:    meta.TableName,
			TableComment: meta.TableComment,
		})
	}
	return result, nil
}

// OnStartup applies SQLite-specific startup behavior before cluster services start.
func (d sqliteDialect) OnStartup(ctx context.Context, runtime RuntimeConfig) error {
	return internalsqlite.OnStartup(ctx, d.link, runtime)
}

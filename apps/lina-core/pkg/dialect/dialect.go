// Package dialect provides the stable database-dialect boundary used by host
// bootstrap commands, plugin SQL lifecycle code, and future tooling.
package dialect

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
	internalpostgres "lina-core/pkg/dialect/internal/postgres"
	internalsqlite "lina-core/pkg/dialect/internal/sqlite"
)

// Supported database link prefixes.
const (
	pgsqlPrefix  = "pgsql:"
	sqlitePrefix = "sqlite:"
)

var (
	// CodeDialectLinkRequired reports that database.default.link is empty.
	CodeDialectLinkRequired = bizerr.MustDefine(
		"DIALECT_LINK_REQUIRED",
		"Database link is required",
		gcode.CodeInvalidParameter,
	)
	// CodeDialectUnsupported reports that a database link prefix is not supported.
	CodeDialectUnsupported = bizerr.MustDefine(
		"DIALECT_UNSUPPORTED",
		"Database dialect {prefix} is unsupported; supported prefixes are pgsql: and sqlite:",
		gcode.CodeInvalidParameter,
	)
	// CodeDialectMySQLUnsupported reports that MySQL support has been removed.
	CodeDialectMySQLUnsupported = bizerr.MustDefine(
		"DIALECT_MYSQL_UNSUPPORTED",
		"mysql dialect is no longer supported; supported prefixes are pgsql: and sqlite:",
		gcode.CodeInvalidParameter,
	)
)

// Dialect abstracts database-engine behavior that cannot be delegated to
// GoFrame's query builder.
type Dialect interface {
	// Name returns the stable dialect name used in logs and diagnostics.
	Name() string
	// TranslateDDL converts one PostgreSQL-source SQL asset into SQL executable by
	// this dialect. sourceName is a file path or embedded asset identifier used
	// for error diagnostics.
	TranslateDDL(ctx context.Context, sourceName string, ddl string) (string, error)
	// PrepareDatabase prepares the configured database before init SQL assets run.
	PrepareDatabase(ctx context.Context, link string, rebuild bool) error
	// SupportsCluster reports whether this database can back multi-node
	// coordination state.
	SupportsCluster() bool
	// DatabaseVersion returns a display-ready database engine and version label.
	DatabaseVersion(ctx context.Context, db gdb.DB) (string, error)
	// QueryTableMetadata returns metadata for the named tables that exist in
	// the requested schema. Missing table names are skipped.
	QueryTableMetadata(ctx context.Context, db gdb.DB, schema string, tableNames []string) ([]TableMeta, error)
	// OnStartup applies dialect-specific runtime bootstrap behavior before
	// cluster services start.
	OnStartup(ctx context.Context, runtime RuntimeConfig) error
}

// TableMeta describes portable table metadata exposed through the dialect
// boundary.
type TableMeta struct {
	TableName    string // TableName is the database table identifier.
	TableComment string // TableComment is the optional table comment.
}

// RuntimeConfig is the narrow startup configuration interface needed by
// dialect hooks. Host config.Service adapts to this interface internally.
type RuntimeConfig interface {
	// OverrideClusterEnabledForDialect locks cluster.enabled in memory for the
	// current process when a dialect cannot support cluster mode.
	OverrideClusterEnabledForDialect(value bool)
}

// From resolves one database dialect from the database.default.link prefix.
func From(link string) (Dialect, error) {
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

// FromDatabase resolves one database dialect from the active GoFrame database
// configuration.
func FromDatabase(db gdb.DB) (Dialect, error) {
	link := ""
	if db != nil && db.GetConfig() != nil {
		configNode := db.GetConfig()
		link = strings.TrimSpace(configNode.Link)
		if link == "" && strings.TrimSpace(configNode.Type) != "" {
			link = strings.TrimSpace(configNode.Type) + ":"
		}
	}
	return From(link)
}

// DatabaseVersion returns a display-ready database engine and version label for
// the active GoFrame database.
func DatabaseVersion(ctx context.Context, db gdb.DB) (string, error) {
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

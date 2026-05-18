// Package dialect provides the stable database-dialect boundary used by host
// bootstrap commands, plugin SQL lifecycle code, and future tooling.
package dialect

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
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
	// Name returns the stable dialect name used in logs, diagnostics, cache keys,
	// and startup decisions. The value must not depend on the active connection.
	Name() string
	// TranslateDDL converts one PostgreSQL-source SQL asset into SQL executable by
	// this dialect. sourceName is a file path or embedded asset identifier used
	// only for error diagnostics; ddl is not modified in-place. Implementations
	// return an error when the asset uses SQL outside the supported migration
	// subset, and callers must treat the output as database-specific init SQL.
	TranslateDDL(ctx context.Context, sourceName string, ddl string) (string, error)
	// PrepareDatabase prepares the configured database before init SQL assets run.
	// link is the raw database.default.link value and rebuild controls destructive
	// local reset behavior for initialization flows. Implementations return errors
	// for unreachable servers, invalid filesystem paths, or unsupported rebuild
	// requests; they do not alter routes, permissions, i18n resources, or caches.
	PrepareDatabase(ctx context.Context, link string, rebuild bool) error
	// SupportsCluster reports whether this database can back multi-node
	// coordination state. Callers use this to decide whether cluster-mode cache
	// coordination can be enabled for the active database.
	SupportsCluster() bool
	// DatabaseVersion returns a display-ready database engine and version label
	// for the provided GoFrame database handle. It returns an error when the
	// version query cannot be executed.
	DatabaseVersion(ctx context.Context, db gdb.DB) (string, error)
	// QueryTableMetadata returns metadata for the named tables that exist in
	// the requested schema. Missing table names are skipped and database errors
	// are returned to the caller without fallback data.
	QueryTableMetadata(ctx context.Context, db gdb.DB, schema string, tableNames []string) ([]TableMeta, error)
	// OnStartup applies dialect-specific runtime bootstrap behavior before
	// cluster services start. Implementations may use runtime to adjust in-memory
	// cluster compatibility flags, and return errors for startup blockers.
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
	// current process when a dialect cannot support cluster mode. The call is a
	// startup-only compatibility adjustment and must not mutate persisted config.
	OverrideClusterEnabledForDialect(value bool)
}

// From resolves one database dialect from the database.default.link prefix.
func From(link string) (Dialect, error) {
	return fromLink(link)
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
	return databaseVersion(ctx, db)
}

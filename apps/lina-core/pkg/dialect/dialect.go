// Package dialect provides the stable database-dialect boundary used by host
// bootstrap commands, plugin SQL lifecycle code, and future tooling.
package dialect

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
	internalmysql "lina-core/pkg/dialect/internal/mysql"
	internalsqlite "lina-core/pkg/dialect/internal/sqlite"
)

// Supported database link prefixes.
const (
	mysqlPrefix  = "mysql:"
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
		"Database dialect {prefix} is unsupported; supported prefixes are mysql: and sqlite:",
		gcode.CodeInvalidParameter,
	)
)

// Dialect abstracts database-engine behavior that cannot be delegated to
// GoFrame's query builder.
type Dialect interface {
	// Name returns the stable dialect name used in logs and diagnostics.
	Name() string
	// TranslateDDL converts one MySQL-source SQL asset into SQL executable by
	// this dialect. sourceName is a file path or embedded asset identifier used
	// for error diagnostics.
	TranslateDDL(ctx context.Context, sourceName string, ddl string) (string, error)
	// PrepareDatabase prepares the configured database before init SQL assets run.
	PrepareDatabase(ctx context.Context, link string, rebuild bool) error
	// SupportsCluster reports whether this database can back multi-node
	// coordination state.
	SupportsCluster() bool
	// OnStartup applies dialect-specific runtime bootstrap behavior before
	// cluster services start.
	OnStartup(ctx context.Context, runtime RuntimeConfig) error
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
	case strings.HasPrefix(normalized, mysqlPrefix):
		return mysqlDialect{}, nil
	case strings.HasPrefix(normalized, sqlitePrefix):
		return sqliteDialect{link: normalized}, nil
	default:
		prefix := normalized
		if index := strings.Index(prefix, ":"); index >= 0 {
			prefix = prefix[:index+1]
		}
		return nil, bizerr.NewCode(CodeDialectUnsupported, bizerr.P("prefix", prefix))
	}
}

// mysqlDialect is the public package wrapper for the internal MySQL dialect.
type mysqlDialect struct{}

// Name returns the stable MySQL dialect name.
func (mysqlDialect) Name() string {
	return internalmysql.Name
}

// TranslateDDL leaves MySQL-source SQL unchanged.
func (mysqlDialect) TranslateDDL(ctx context.Context, sourceName string, ddl string) (string, error) {
	return internalmysql.TranslateDDL(ctx, sourceName, ddl)
}

// PrepareDatabase creates the configured MySQL database before init SQL runs.
func (mysqlDialect) PrepareDatabase(ctx context.Context, link string, rebuild bool) error {
	return internalmysql.PrepareDatabase(ctx, link, rebuild)
}

// SupportsCluster reports whether MySQL can back cluster coordination tables.
func (mysqlDialect) SupportsCluster() bool {
	return internalmysql.SupportsCluster()
}

// OnStartup has no MySQL-specific startup side effects.
func (mysqlDialect) OnStartup(ctx context.Context, runtime RuntimeConfig) error {
	return nil
}

// sqliteDialect is the public package wrapper for the internal SQLite dialect.
type sqliteDialect struct {
	link string // link stores the source database link for startup diagnostics.
}

// Name returns the stable SQLite dialect name.
func (sqliteDialect) Name() string {
	return internalsqlite.Name
}

// TranslateDDL converts the project's MySQL-source SQL subset to SQLite SQL.
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

// OnStartup applies SQLite-specific startup behavior before cluster services start.
func (d sqliteDialect) OnStartup(ctx context.Context, runtime RuntimeConfig) error {
	return internalsqlite.OnStartup(ctx, d.link, runtime)
}

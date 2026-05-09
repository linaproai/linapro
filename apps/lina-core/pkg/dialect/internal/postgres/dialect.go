// Package postgres implements LinaPro's internal PostgreSQL dialect behavior.
package postgres

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Name is the stable PostgreSQL dialect name.
const Name = "postgres"

// TranslateDDL leaves PostgreSQL-source SQL unchanged.
func TranslateDDL(ctx context.Context, sourceName string, ddl string) (string, error) {
	return ddl, nil
}

// SupportsCluster reports that PostgreSQL can back shared multi-node
// coordination tables.
func SupportsCluster() bool {
	return true
}

// DatabaseVersion returns the PostgreSQL server version label.
func DatabaseVersion(ctx context.Context, db gdb.DB) (string, error) {
	if db == nil {
		return "", gerror.New("database connection is required")
	}
	result, err := db.GetValue(ctx, "SELECT version()")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.String()), nil
}

// OnStartup has no PostgreSQL-specific startup side effects.
func OnStartup(ctx context.Context, runtime any) error {
	return nil
}

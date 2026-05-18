// This file queries PostgreSQL table metadata for host services.

package postgres

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
)

// TableMeta carries PostgreSQL table metadata needed by the public dialect
// wrapper.
type TableMeta struct {
	TableName    string
	TableComment string
}

// QueryTableMetadata returns existing PostgreSQL table names and comments.
func QueryTableMetadata(ctx context.Context, db gdb.DB, schema string, tableNames []string) ([]TableMeta, error) {
	if len(tableNames) == 0 {
		return []TableMeta{}, nil
	}
	normalizedSchema := strings.TrimSpace(schema)
	if normalizedSchema == "" {
		normalizedSchema = "public"
	}
	rows, err := db.GetAll(
		ctx,
		`SELECT t.table_name, COALESCE(obj_description(c.oid), '') AS table_comment
FROM information_schema.tables t
LEFT JOIN pg_namespace n ON n.nspname = t.table_schema
LEFT JOIN pg_class c ON c.relname = t.table_name AND c.relnamespace = n.oid AND c.relkind = 'r'
WHERE t.table_schema = ? AND t.table_name IN(?)
ORDER BY t.table_name`,
		normalizedSchema,
		tableNames,
	)
	if err != nil {
		return nil, err
	}

	metas := make([]TableMeta, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		tableName := strings.TrimSpace(row["table_name"].String())
		if tableName == "" {
			continue
		}
		metas = append(metas, TableMeta{
			TableName:    tableName,
			TableComment: strings.TrimSpace(row["table_comment"].String()),
		})
	}
	return metas, nil
}

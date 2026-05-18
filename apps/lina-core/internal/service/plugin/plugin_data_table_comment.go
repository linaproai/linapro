// This file resolves best-effort human-readable comments for data host-service
// table authorizations so review UIs can show more than raw table names.

package plugin

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
)

// dataTableMetadataSchema is the host schema used by PostgreSQL metadata
// lookups. SQLite ignores it, and other dialects may map it internally.
const dataTableMetadataSchema = "public"

// ResolveDataTableComments resolves host-side table comments for the given
// data-table names. It degrades to an empty map when metadata lookup is
// unavailable so plugin list APIs are not blocked by optional schema comments.
func (s *serviceImpl) ResolveDataTableComments(ctx context.Context, tables []string) map[string]string {
	normalizedTables := normalizeDataTableNames(tables)
	if len(normalizedTables) == 0 {
		return map[string]string{}
	}

	db := g.DB()
	dbDialect, err := dialect.FromDatabase(db)
	if err != nil {
		logger.Warningf(ctx, "resolve plugin data table comments skipped: %v", err)
		return map[string]string{}
	}

	metas, err := dbDialect.QueryTableMetadata(ctx, db, dataTableMetadataSchema, normalizedTables)
	if err != nil {
		logger.Warningf(ctx, "resolve plugin data table comments failed schema=%s tables=%v err=%v", dataTableMetadataSchema, normalizedTables, err)
		return map[string]string{}
	}
	return dataTableCommentsFromMetadata(metas)
}

// dataTableCommentsFromMetadata maps dialect table metadata to the comment map
// consumed by plugin governance projections.
func dataTableCommentsFromMetadata(metas []dialect.TableMeta) map[string]string {
	comments := make(map[string]string, len(metas))
	for _, meta := range metas {
		tableName := strings.TrimSpace(meta.TableName)
		tableComment := strings.TrimSpace(meta.TableComment)
		if tableName == "" || tableComment == "" {
			continue
		}
		comments[tableName] = tableComment
	}
	return comments
}

// normalizeDataTableNames trims blanks and de-duplicates table names before
// they are used in metadata lookup queries.
func normalizeDataTableNames(tables []string) []string {
	if len(tables) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(tables))
	normalized := make([]string, 0, len(tables))
	for _, table := range tables {
		name := strings.TrimSpace(table)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	return normalized
}

// This file resolves best-effort human-readable comments for data host-service
// table authorizations so review UIs can show more than raw table names.

package plugin

import (
	"context"
	"strings"

	"lina-core/pkg/logger"
	plugindbhost "lina-core/pkg/plugindb/host"
)

const pluginDataDriverTypePrefix = "plugin-data-"

// ResolveDataTableComments resolves host-side table comments for the given
// data-table names. It degrades to an empty map when metadata lookup is
// unavailable so plugin list APIs are not blocked by optional schema comments.
func (s *serviceImpl) ResolveDataTableComments(ctx context.Context, tables []string) map[string]string {
	normalizedTables := normalizeDataTableNames(tables)
	if len(normalizedTables) == 0 {
		return map[string]string{}
	}

	db, err := plugindbhost.DB()
	if err != nil {
		logger.Warningf(ctx, "resolve plugin data table comments skipped: %v", err)
		return map[string]string{}
	}

	dbType := normalizePluginDataMetadataDBType("")
	if config := db.GetConfig(); config != nil {
		dbType = normalizePluginDataMetadataDBType(config.Type)
	}
	switch dbType {
	case "mariadb", "mysql", "tidb":
	default:
		return map[string]string{}
	}

	schema := strings.TrimSpace(db.GetSchema())
	if schema == "" {
		if config := db.GetConfig(); config != nil {
			schema = strings.TrimSpace(config.Name)
		}
	}
	if schema == "" {
		return map[string]string{}
	}

	records, err := db.Model("information_schema.TABLES").
		Fields("TABLE_NAME AS table_name", "TABLE_COMMENT AS table_comment").
		Where("TABLE_SCHEMA", schema).
		WhereIn("TABLE_NAME", normalizedTables).
		All()
	if err != nil {
		logger.Warningf(ctx, "resolve plugin data table comments failed schema=%s tables=%v err=%v", schema, normalizedTables, err)
		return map[string]string{}
	}

	comments := make(map[string]string, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		tableName := strings.TrimSpace(record["table_name"].String())
		tableComment := strings.TrimSpace(record["table_comment"].String())
		if tableName == "" || tableComment == "" {
			continue
		}
		comments[tableName] = tableComment
	}
	return comments
}

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

func normalizePluginDataMetadataDBType(dbType string) string {
	normalized := strings.ToLower(strings.TrimSpace(dbType))
	if strings.HasPrefix(normalized, pluginDataDriverTypePrefix) {
		return strings.TrimPrefix(normalized, pluginDataDriverTypePrefix)
	}
	return normalized
}

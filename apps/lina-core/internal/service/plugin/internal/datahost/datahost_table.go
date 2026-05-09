// This file builds table-backed data service contracts directly from the
// authorized table name and live schema metadata.

package datahost

import (
	"context"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
)

// BuildAuthorizedTableContract synthesizes one internal resource contract for a
// directly authorized data table.
func BuildAuthorizedTableContract(
	ctx context.Context,
	table string,
	methods []string,
) (*catalog.ResourceSpec, error) {
	normalizedTable := strings.TrimSpace(table)
	if normalizedTable == "" {
		return nil, gerror.New("data service table cannot be empty")
	}

	db, err := getPluginDataDB()
	if err != nil {
		return nil, err
	}
	// Dynamic-plugin install/uninstall SQL can create, drop, or reshape the
	// governed table at runtime, so cached field metadata must be refreshed
	// before rebuilding one authorized table contract.
	if err = db.GetCore().ClearTableFields(ctx, normalizedTable); err != nil {
		return nil, err
	}
	tableFields, err := db.TableFields(ctx, normalizedTable)
	if err != nil {
		return nil, err
	}
	if len(tableFields) == 0 {
		return nil, gerror.Newf("data service table does not exist or is not readable: %s", normalizedTable)
	}

	orderedFields := sortTableFields(tableFields)
	identityColumns, err := listIdentityColumns(ctx, db, normalizedTable)
	if err != nil {
		return nil, err
	}
	var (
		fields         = make([]*catalog.ResourceField, 0, len(orderedFields))
		filters        = make([]*catalog.ResourceQuery, 0, len(orderedFields))
		writableFields = make([]string, 0, len(orderedFields))
	)

	keyField := ""
	orderByColumn := ""
	for _, field := range orderedFields {
		if field == nil {
			continue
		}
		columnName := strings.TrimSpace(field.Name)
		if columnName == "" {
			continue
		}
		fieldName := buildTableFieldAlias(columnName)
		fields = append(fields, &catalog.ResourceField{
			Name:   fieldName,
			Column: columnName,
		})
		filters = append(filters, &catalog.ResourceQuery{
			Param:    fieldName,
			Column:   columnName,
			Operator: catalog.ResourceFilterOperatorEQ.String(),
		})
		if strings.EqualFold(strings.TrimSpace(field.Key), "PRI") && keyField == "" {
			keyField = fieldName
		}
		if orderByColumn == "" && strings.EqualFold(strings.TrimSpace(field.Key), "PRI") {
			orderByColumn = columnName
		}
		if !isAutoManagedTableField(field, identityColumns) {
			writableFields = append(writableFields, fieldName)
		}
	}

	if keyField == "" {
		if _, ok := tableFields["id"]; ok {
			keyField = "id"
		}
	}
	if keyField == "" {
		return nil, gerror.Newf("data service table %s is missing a recognizable primary key", normalizedTable)
	}
	if orderByColumn == "" {
		orderByColumn = keyField
	}

	return &catalog.ResourceSpec{
		Key:            normalizedTable,
		Type:           catalog.ResourceSpecTypeTableList.String(),
		Table:          normalizedTable,
		Fields:         fields,
		Filters:        filters,
		Operations:     normalizeAuthorizedTableMethods(methods),
		KeyField:       keyField,
		WritableFields: writableFields,
		Access:         catalog.ResourceAccessModeBoth.String(),
		OrderBy: catalog.ResourceOrderBySpec{
			Column:    orderByColumn,
			Direction: catalog.ResourceOrderDirectionASC.String(),
		},
	}, nil
}

// sortTableFields returns table fields in schema order with name tie-breaking.
func sortTableFields(fields map[string]*gdb.TableField) []*gdb.TableField {
	ordered := make([]*gdb.TableField, 0, len(fields))
	for _, field := range fields {
		if field != nil {
			ordered = append(ordered, field)
		}
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Index == ordered[j].Index {
			return strings.TrimSpace(ordered[i].Name) < strings.TrimSpace(ordered[j].Name)
		}
		return ordered[i].Index < ordered[j].Index
	})
	return ordered
}

// normalizeAuthorizedTableMethods trims, lowercases, sorts, and returns authorized methods.
func normalizeAuthorizedTableMethods(methods []string) []string {
	allowed := make([]string, 0, len(methods))
	for _, method := range methods {
		normalized := strings.ToLower(strings.TrimSpace(method))
		if normalized != "" {
			allowed = append(allowed, normalized)
		}
	}
	if len(allowed) == 0 {
		return []string{}
	}
	sort.Strings(allowed)
	return allowed
}

// isAutoManagedTableField reports whether the host manages the field automatically.
func isAutoManagedTableField(field *gdb.TableField, identityColumns map[string]struct{}) bool {
	if field == nil {
		return true
	}
	name := strings.ToLower(strings.TrimSpace(field.Name))
	extra := strings.ToLower(strings.TrimSpace(field.Extra))
	if _, ok := identityColumns[name]; ok {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(field.Key), "PRI") && strings.Contains(extra, "auto_increment") {
		return true
	}
	switch name {
	case "created_at", "updated_at", "deleted_at":
		return true
	default:
		return false
	}
}

// listIdentityColumns returns PostgreSQL identity columns for table-backed
// resource contracts. Other drivers either expose auto-increment through field
// metadata or have no identity columns to hide from guest write contracts.
func listIdentityColumns(ctx context.Context, db gdb.DB, table string) (map[string]struct{}, error) {
	columns := make(map[string]struct{})
	if db == nil || strings.TrimSpace(table) == "" {
		return columns, nil
	}
	config := db.GetConfig()
	if config == nil {
		return columns, nil
	}
	dbType := strings.ToLower(strings.TrimSpace(config.Type))
	if dbType != "pgsql" && dbType != "plugin-data-pgsql" {
		return columns, nil
	}
	records, err := db.GetAll(
		ctx,
		`SELECT column_name FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = $1 AND is_identity = 'YES'`,
		table,
	)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if record == nil {
			continue
		}
		columnName := strings.ToLower(strings.TrimSpace(record["column_name"].String()))
		if columnName != "" {
			columns[columnName] = struct{}{}
		}
	}
	return columns, nil
}

// buildTableFieldAlias converts snake_case columns into logical camelCase field names.
func buildTableFieldAlias(columnName string) string {
	normalized := strings.TrimSpace(strings.ToLower(columnName))
	if normalized == "" || !strings.Contains(normalized, "_") {
		return normalized
	}
	parts := strings.Split(normalized, "_")
	if len(parts) == 0 {
		return normalized
	}
	builder := strings.Builder{}
	builder.WriteString(parts[0])
	for _, part := range parts[1:] {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	return builder.String()
}

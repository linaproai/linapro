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
		return nil, gerror.New("data service table 不能为空")
	}

	db, err := getPluginDataDB()
	if err != nil {
		return nil, err
	}
	tableFields, err := db.TableFields(ctx, normalizedTable)
	if err != nil {
		return nil, err
	}
	if len(tableFields) == 0 {
		return nil, gerror.Newf("data service table 不存在或不可读: %s", normalizedTable)
	}

	orderedFields := sortTableFields(tableFields)
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
		if !isAutoManagedTableField(field) {
			writableFields = append(writableFields, fieldName)
		}
	}

	if keyField == "" {
		if _, ok := tableFields["id"]; ok {
			keyField = "id"
		}
	}
	if keyField == "" {
		return nil, gerror.Newf("data service table %s 缺少可识别主键", normalizedTable)
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

func isAutoManagedTableField(field *gdb.TableField) bool {
	if field == nil {
		return true
	}
	name := strings.ToLower(strings.TrimSpace(field.Name))
	extra := strings.ToLower(strings.TrimSpace(field.Extra))
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

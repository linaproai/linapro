// This file applies typed plugindb query plans to governed data service
// requests.

package datahost

import (
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/plugindb/shared"
)

func decodeDataListPlan(table string, request *pluginbridge.HostServiceDataListRequest) (*shared.DataQueryPlan, error) {
	plan, err := pluginbridge.DecodeHostServiceDataListPlan(request)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		request = normalizeDataListRequest(request)
		plan = &shared.DataQueryPlan{
			Table:  strings.TrimSpace(table),
			Action: shared.DataPlanActionList,
			Page: &shared.DataPagination{
				PageNum:  request.PageNum,
				PageSize: request.PageSize,
			},
		}
		for field, value := range request.Filters {
			if strings.TrimSpace(value) == "" {
				continue
			}
			filter, filterErr := shared.NewEQFilter(field, value)
			if filterErr != nil {
				return nil, filterErr
			}
			plan.Filters = append(plan.Filters, filter)
		}
		return plan, nil
	}
	if strings.TrimSpace(plan.Table) == "" {
		plan.Table = strings.TrimSpace(table)
	}
	if strings.TrimSpace(plan.Table) != strings.TrimSpace(table) {
		return nil, gerror.Newf("plugindb query plan table mismatch: %s != %s", plan.Table, table)
	}
	if plan.Action == "" {
		plan.Action = shared.DataPlanActionList
	}
	if plan.Action != shared.DataPlanActionList && plan.Action != shared.DataPlanActionCount {
		return nil, gerror.Newf("plugindb list request action is invalid: %s", plan.Action)
	}
	if plan.Action == shared.DataPlanActionList {
		if plan.Page == nil {
			plan.Page = &shared.DataPagination{PageNum: defaultDataListPageNum, PageSize: defaultDataListPageSize}
		}
		if plan.Page.PageNum <= 0 {
			plan.Page.PageNum = defaultDataListPageNum
		}
		if plan.Page.PageSize <= 0 {
			plan.Page.PageSize = defaultDataListPageSize
		}
		if plan.Page.PageSize > maxDataListPageSize {
			plan.Page.PageSize = maxDataListPageSize
		}
	}
	return plan, shared.ValidateDataQueryPlan(plan)
}

func decodeDataGetPlan(table string, request *pluginbridge.HostServiceDataGetRequest) (*shared.DataQueryPlan, error) {
	plan, err := pluginbridge.DecodeHostServiceDataGetPlan(request)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		plan = &shared.DataQueryPlan{Table: strings.TrimSpace(table), Action: shared.DataPlanActionGet}
	}
	if strings.TrimSpace(plan.Table) == "" {
		plan.Table = strings.TrimSpace(table)
	}
	if strings.TrimSpace(plan.Table) != strings.TrimSpace(table) {
		return nil, gerror.Newf("plugindb get request table mismatch: %s != %s", plan.Table, table)
	}
	if plan.Action == "" {
		plan.Action = shared.DataPlanActionGet
	}
	if plan.Action != shared.DataPlanActionGet {
		return nil, gerror.Newf("plugindb get request action is invalid: %s", plan.Action)
	}
	if request != nil && len(plan.KeyJSON) == 0 {
		plan.KeyJSON = append([]byte(nil), request.KeyJSON...)
	}
	if len(plan.KeyJSON) == 0 {
		return nil, gerror.New("data key 不能为空")
	}
	return plan, shared.ValidateDataQueryPlan(plan)
}

func applyPlanFilters(model *gdb.Model, resource *catalog.ResourceSpec, filters []*shared.DataFilter) (*gdb.Model, error) {
	if model == nil || resource == nil || len(filters) == 0 {
		return model, nil
	}
	for _, filter := range filters {
		if err := shared.ValidateDataFilter(filter); err != nil {
			return nil, err
		}
		column := resolveResourceFieldColumn(resource, filter.Field)
		if column == "" {
			return nil, gerror.Newf("plugindb filter field 未授权: %s", filter.Field)
		}
		switch filter.Operator {
		case shared.DataFilterOperatorEQ:
			value, err := shared.UnmarshalValueJSON(filter.ValueJSON)
			if err != nil {
				return nil, err
			}
			model = model.Where(column, value)
		case shared.DataFilterOperatorIN:
			values, err := shared.UnmarshalValuesJSON(filter.ValuesJSON)
			if err != nil {
				return nil, err
			}
			if len(values) == 0 {
				return nil, gerror.Newf("plugindb filter %s requires at least one value", filter.Operator)
			}
			model = model.WhereIn(column, values)
		case shared.DataFilterOperatorLike:
			value, err := shared.UnmarshalValueJSON(filter.ValueJSON)
			if err != nil {
				return nil, err
			}
			model = model.WhereLike(column, "%"+fmt.Sprint(value)+"%")
		default:
			return nil, gerror.Newf("plugindb filter operator 不支持: %s", filter.Operator)
		}
	}
	return model, nil
}

func buildPlanFieldArgs(resource *catalog.ResourceSpec, selected []string) ([]any, error) {
	if len(selected) == 0 {
		return buildResourceFieldArgs(resource), nil
	}
	fields := make([]any, 0, len(selected))
	seen := make(map[string]struct{}, len(selected))
	for _, fieldName := range selected {
		normalizedField := strings.TrimSpace(fieldName)
		if normalizedField == "" {
			return nil, gerror.New("plugindb selected field cannot be empty")
		}
		if _, ok := seen[normalizedField]; ok {
			continue
		}
		seen[normalizedField] = struct{}{}
		column := resolveResourceFieldColumn(resource, normalizedField)
		if column == "" {
			return nil, gerror.Newf("plugindb selected field 未授权: %s", normalizedField)
		}
		fields = append(fields, fmt.Sprintf("%s AS %s", column, normalizedField))
	}
	return fields, nil
}

func buildPlanOrderBy(resource *catalog.ResourceSpec, orders []*shared.DataOrder) (string, error) {
	if len(orders) == 0 {
		return buildResourceOrderBy(resource), nil
	}
	parts := make([]string, 0, len(orders))
	for _, order := range orders {
		if err := shared.ValidateDataOrder(order); err != nil {
			return "", err
		}
		column := resolveResourceFieldColumn(resource, order.Field)
		if column == "" {
			return "", gerror.Newf("plugindb order field 未授权: %s", order.Field)
		}
		direction := "ASC"
		if order.Direction == shared.DataOrderDirectionDESC {
			direction = "DESC"
		}
		parts = append(parts, column+" "+direction)
	}
	return strings.Join(parts, ", "), nil
}

func buildResourceRecordWithSelection(recordMap map[string]interface{}, resource *catalog.ResourceSpec, selected []string) map[string]interface{} {
	if len(selected) == 0 {
		return buildResourceRecord(recordMap, resource)
	}
	row := make(map[string]interface{}, len(selected))
	seen := make(map[string]struct{}, len(selected))
	for _, fieldName := range selected {
		normalizedField := strings.TrimSpace(fieldName)
		if normalizedField == "" {
			continue
		}
		if _, ok := seen[normalizedField]; ok {
			continue
		}
		seen[normalizedField] = struct{}{}
		row[normalizedField] = normalizeResourceValue(recordMap[normalizedField])
	}
	return row
}

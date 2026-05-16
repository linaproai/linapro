// This file implements guest-side governed data query builder methods.

package plugindb

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugindb/shared"
)

// Table starts one single-table governed query builder.
func (db *DB) Table(table string) *Query {
	return &Query{
		table: strings.TrimSpace(table),
		plan:  &shared.DataQueryPlan{Table: strings.TrimSpace(table)},
	}
}

// Fields requests one field projection.
func (q *Query) Fields(fields ...string) *Query {
	if q.err != nil {
		return q
	}
	for _, field := range fields {
		normalized := strings.TrimSpace(field)
		if normalized == "" {
			q.err = gerror.New("plugindb fields contains an empty field name")
			return q
		}
		q.plan.Fields = append(q.plan.Fields, normalized)
	}
	return q
}

// Where appends one typed filter clause.
func (q *Query) Where(field string, operator shared.DataFilterOperator, value any) *Query {
	if q.err != nil {
		return q
	}
	normalizedField := strings.TrimSpace(field)
	if normalizedField == "" {
		q.err = gerror.New("plugindb where field cannot be empty")
		return q
	}
	if !operator.IsValid() {
		q.err = gerror.Newf("plugindb where operator is invalid: %s", operator)
		return q
	}
	var (
		filter *shared.DataFilter
		err    error
	)
	switch operator {
	case shared.DataFilterOperatorEQ:
		filter, err = shared.NewEQFilter(normalizedField, value)
	case shared.DataFilterOperatorIN:
		filter, err = shared.NewINFilter(normalizedField, value)
	case shared.DataFilterOperatorLike:
		filter, err = shared.NewLikeFilter(normalizedField, value)
	default:
		err = gerror.Newf("plugindb where operator is unsupported: %s", operator)
	}
	if err != nil {
		q.err = err
		return q
	}
	q.plan.Filters = append(q.plan.Filters, filter)
	return q
}

// WhereEq appends one equality filter.
func (q *Query) WhereEq(field string, value any) *Query {
	return q.Where(field, shared.DataFilterOperatorEQ, value)
}

// WhereIn appends one list-membership filter.
func (q *Query) WhereIn(field string, values any) *Query {
	return q.Where(field, shared.DataFilterOperatorIN, values)
}

// WhereLike appends one wildcard filter.
func (q *Query) WhereLike(field string, value any) *Query {
	return q.Where(field, shared.DataFilterOperatorLike, value)
}

// WhereKey sets the key used by get/update/delete operations.
func (q *Query) WhereKey(key any) *Query {
	if q.err != nil {
		return q
	}
	keyJSON, err := shared.MarshalValueJSON(key)
	if err != nil {
		q.err = err
		return q
	}
	q.plan.KeyJSON = keyJSON
	return q
}

// Order appends one typed order clause.
func (q *Query) Order(field string, direction shared.DataOrderDirection) *Query {
	if q.err != nil {
		return q
	}
	normalizedField := strings.TrimSpace(field)
	if normalizedField == "" {
		q.err = gerror.New("plugindb order field cannot be empty")
		return q
	}
	if !direction.IsValid() {
		q.err = gerror.Newf("plugindb order direction is invalid: %s", direction)
		return q
	}
	q.plan.Orders = append(q.plan.Orders, &shared.DataOrder{Field: normalizedField, Direction: direction})
	return q
}

// OrderAsc appends one ascending order clause.
func (q *Query) OrderAsc(field string) *Query {
	return q.Order(field, shared.DataOrderDirectionASC)
}

// OrderDesc appends one descending order clause.
func (q *Query) OrderDesc(field string) *Query {
	return q.Order(field, shared.DataOrderDirectionDESC)
}

// Page applies one paging window.
func (q *Query) Page(pageNum int32, pageSize int32) *Query {
	if q.err != nil {
		return q
	}
	q.plan.Page = &shared.DataPagination{PageNum: pageNum, PageSize: pageSize}
	return q
}

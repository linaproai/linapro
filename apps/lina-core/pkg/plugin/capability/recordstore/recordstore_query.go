// This file implements guest-side governed record store query builder methods.

package recordstore

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	dataplan "lina-core/pkg/plugin/capability/recordstore/internal/plan"
)

// Table starts one single-table governed query builder.
func (db *DB) Table(table string) *Query {
	var invoker HostServiceInvoker
	if db != nil {
		invoker = db.invoker
	}
	return &Query{
		table:   strings.TrimSpace(table),
		plan:    &dataplan.QueryPlan{Table: strings.TrimSpace(table)},
		invoker: invoker,
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
			q.err = gerror.New("record store capability fields contains an empty field name")
			return q
		}
		q.plan.Fields = append(q.plan.Fields, normalized)
	}
	return q
}

// where appends one typed filter clause.
func (q *Query) where(field string, operator dataplan.FilterOperator, value any) *Query {
	if q.err != nil {
		return q
	}
	normalizedField := strings.TrimSpace(field)
	if normalizedField == "" {
		q.err = gerror.New("record store capability where field cannot be empty")
		return q
	}
	if !operator.IsValid() {
		q.err = gerror.Newf("record store capability where operator is invalid: %s", operator)
		return q
	}
	var (
		filter *dataplan.Filter
		err    error
	)
	switch operator {
	case dataplan.FilterOperatorEQ:
		filter, err = dataplan.NewEQFilter(normalizedField, value)
	case dataplan.FilterOperatorIN:
		filter, err = dataplan.NewINFilter(normalizedField, value)
	case dataplan.FilterOperatorLike:
		filter, err = dataplan.NewLikeFilter(normalizedField, value)
	default:
		err = gerror.Newf("record store capability where operator is unsupported: %s", operator)
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
	return q.where(field, dataplan.FilterOperatorEQ, value)
}

// WhereIn appends one list-membership filter.
func (q *Query) WhereIn(field string, values any) *Query {
	return q.where(field, dataplan.FilterOperatorIN, values)
}

// WhereLike appends one wildcard filter.
func (q *Query) WhereLike(field string, value any) *Query {
	return q.where(field, dataplan.FilterOperatorLike, value)
}

// WhereKey sets the key used by get/update/delete operations.
func (q *Query) WhereKey(key any) *Query {
	if q.err != nil {
		return q
	}
	keyJSON, err := dataplan.MarshalValueJSON(key)
	if err != nil {
		q.err = err
		return q
	}
	q.plan.KeyJSON = keyJSON
	return q
}

// order appends one typed order clause.
func (q *Query) order(field string, direction dataplan.OrderDirection) *Query {
	if q.err != nil {
		return q
	}
	normalizedField := strings.TrimSpace(field)
	if normalizedField == "" {
		q.err = gerror.New("record store capability order field cannot be empty")
		return q
	}
	if !direction.IsValid() {
		q.err = gerror.Newf("record store capability order direction is invalid: %s", direction)
		return q
	}
	q.plan.Orders = append(q.plan.Orders, &dataplan.Order{Field: normalizedField, Direction: direction})
	return q
}

// OrderAsc appends one ascending order clause.
func (q *Query) OrderAsc(field string) *Query {
	return q.order(field, dataplan.OrderDirectionASC)
}

// OrderDesc appends one descending order clause.
func (q *Query) OrderDesc(field string) *Query {
	return q.order(field, dataplan.OrderDirectionDESC)
}

// Page applies one paging window.
func (q *Query) Page(pageNum int32, pageSize int32) *Query {
	if q.err != nil {
		return q
	}
	q.plan.Page = &dataplan.Pagination{PageNum: pageNum, PageSize: pageSize}
	return q
}

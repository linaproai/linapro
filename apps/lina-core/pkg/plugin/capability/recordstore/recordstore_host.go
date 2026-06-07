// This file exposes structured record store query-plan contracts shared by guest
// builders and host-side governed table execution.

package recordstore

import dataplan "lina-core/pkg/plugin/capability/recordstore/internal/plan"

// QueryPlan represents one governed typed record store query plan.
type QueryPlan = dataplan.QueryPlan

// Filter represents one governed typed filter clause.
type Filter = dataplan.Filter

// Order represents one governed typed order clause.
type Order = dataplan.Order

// Pagination represents one governed typed page window.
type Pagination = dataplan.Pagination

// PlanAction represents one governed record store plan action.
type PlanAction = dataplan.PlanAction

// FilterOperator represents one governed filter operator.
type FilterOperator = dataplan.FilterOperator

// OrderDirection represents one governed order direction.
type OrderDirection = dataplan.OrderDirection

const (
	// PlanActionList lists records from one authorized table.
	PlanActionList = dataplan.PlanActionList
	// PlanActionGet reads one record by key from one authorized table.
	PlanActionGet = dataplan.PlanActionGet
	// PlanActionCount counts records from one authorized table.
	PlanActionCount = dataplan.PlanActionCount
	// OrderDirectionDESC orders records in descending order.
	OrderDirectionDESC = dataplan.OrderDirectionDESC
	// FilterOperatorEQ compares one field by equality.
	FilterOperatorEQ = dataplan.FilterOperatorEQ
	// FilterOperatorIN compares one field against a value list.
	FilterOperatorIN = dataplan.FilterOperatorIN
	// FilterOperatorLike compares one field by wildcard matching.
	FilterOperatorLike = dataplan.FilterOperatorLike
)

// UnmarshalQueryPlanJSON decodes one governed typed query plan.
func UnmarshalQueryPlanJSON(data []byte) (*QueryPlan, error) {
	return dataplan.UnmarshalQueryPlanJSON(data)
}

// ValidateQueryPlan validates one governed typed query plan.
func ValidateQueryPlan(plan *QueryPlan) error {
	return dataplan.ValidateQueryPlan(plan)
}

// ValidateFilter validates one governed typed filter clause.
func ValidateFilter(filter *Filter) error {
	return dataplan.ValidateFilter(filter)
}

// ValidateOrder validates one governed typed order clause.
func ValidateOrder(order *Order) error {
	return dataplan.ValidateOrder(order)
}

// UnmarshalValueJSON decodes one JSON-encoded scalar or object value.
func UnmarshalValueJSON(data []byte) (any, error) {
	return dataplan.UnmarshalValueJSON(data)
}

// UnmarshalValuesJSON decodes one list of JSON-encoded values.
func UnmarshalValuesJSON(items [][]byte) ([]any, error) {
	return dataplan.UnmarshalValuesJSON(items)
}

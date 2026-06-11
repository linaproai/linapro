// This file defines the typed record store capability enums shared by guest builders and
// host-side execution components.

package plan

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// PlanAction represents one governed record store plan action.
type PlanAction string

// Record store plan action constants.
const (
	// PlanActionList lists records from one authorized table.
	PlanActionList PlanAction = "list"
	// PlanActionGet reads one record by key from one authorized table.
	PlanActionGet PlanAction = "get"
	// PlanActionCount counts records from one authorized table.
	PlanActionCount PlanAction = "count"
	// PlanActionCreate inserts one record into one authorized table.
	PlanActionCreate PlanAction = "create"
	// PlanActionUpdate updates one record in one authorized table.
	PlanActionUpdate PlanAction = "update"
	// PlanActionDelete deletes one record from one authorized table.
	PlanActionDelete PlanAction = "delete"
	// PlanActionTransaction executes one structured mutation transaction.
	PlanActionTransaction PlanAction = "transaction"
)

// FilterOperator represents one supported filter operator.
type FilterOperator string

// Record store filter operator constants.
const (
	// FilterOperatorEQ compares one field by equality.
	FilterOperatorEQ FilterOperator = "eq"
	// FilterOperatorIN compares one field against a value list.
	FilterOperatorIN FilterOperator = "in"
	// FilterOperatorLike compares one field by wildcard matching.
	FilterOperatorLike FilterOperator = "like"
)

// OrderDirection represents one supported order direction.
type OrderDirection string

// Record store order direction constants.
const (
	// OrderDirectionASC orders records in ascending order.
	OrderDirectionASC OrderDirection = "asc"
	// OrderDirectionDESC orders records in descending order.
	OrderDirectionDESC OrderDirection = "desc"
)

// MutationAction represents one transaction mutation action.
type MutationAction string

// Mutation action constants.
const (
	// MutationActionCreate enqueues one insert mutation.
	MutationActionCreate MutationAction = "create"
	// MutationActionUpdate enqueues one update mutation.
	MutationActionUpdate MutationAction = "update"
	// MutationActionDelete enqueues one delete mutation.
	MutationActionDelete MutationAction = "delete"
)

// AccessMode represents one runtime access requirement for a table contract.
type AccessMode string

// Access mode constants.
const (
	// AccessModeRequest requires a request-bound execution context.
	AccessModeRequest AccessMode = "request"
	// AccessModeSystem allows a system-bound execution context.
	AccessModeSystem AccessMode = "system"
	// AccessModeBoth allows both request-bound and system-bound execution contexts.
	AccessModeBoth AccessMode = "both"
)

// String returns the string representation of the plan action.
func (value PlanAction) String() string { return string(value) }

// IsValid reports whether the plan action is one of the supported constants.
func (value PlanAction) IsValid() bool {
	switch value {
	case PlanActionList,
		PlanActionGet,
		PlanActionCount,
		PlanActionCreate,
		PlanActionUpdate,
		PlanActionDelete,
		PlanActionTransaction:
		return true
	default:
		return false
	}
}

// ParsePlanAction parses one raw value into a typed plan action.
func ParsePlanAction(value string) (PlanAction, error) {
	normalized := PlanAction(strings.ToLower(strings.TrimSpace(value)))
	if !normalized.IsValid() {
		return "", gerror.Newf("invalid record store plan action: %s", value)
	}
	return normalized, nil
}

// String returns the string representation of the filter operator.
func (value FilterOperator) String() string { return string(value) }

// IsValid reports whether the filter operator is supported.
func (value FilterOperator) IsValid() bool {
	switch value {
	case FilterOperatorEQ, FilterOperatorIN, FilterOperatorLike:
		return true
	default:
		return false
	}
}

// ParseFilterOperator parses one raw value into a typed filter operator.
func ParseFilterOperator(value string) (FilterOperator, error) {
	normalized := FilterOperator(strings.ToLower(strings.TrimSpace(value)))
	if !normalized.IsValid() {
		return "", gerror.Newf("invalid record store filter operator: %s", value)
	}
	return normalized, nil
}

// String returns the string representation of the order direction.
func (value OrderDirection) String() string { return string(value) }

// IsValid reports whether the order direction is supported.
func (value OrderDirection) IsValid() bool {
	switch value {
	case OrderDirectionASC, OrderDirectionDESC:
		return true
	default:
		return false
	}
}

// ParseOrderDirection parses one raw value into a typed order direction.
func ParseOrderDirection(value string) (OrderDirection, error) {
	normalized := OrderDirection(strings.ToLower(strings.TrimSpace(value)))
	if !normalized.IsValid() {
		return "", gerror.Newf("invalid record store order direction: %s", value)
	}
	return normalized, nil
}

// String returns the string representation of the mutation action.
func (value MutationAction) String() string { return string(value) }

// IsValid reports whether the mutation action is supported.
func (value MutationAction) IsValid() bool {
	switch value {
	case MutationActionCreate, MutationActionUpdate, MutationActionDelete:
		return true
	default:
		return false
	}
}

// ParseMutationAction parses one raw value into a typed mutation action.
func ParseMutationAction(value string) (MutationAction, error) {
	normalized := MutationAction(strings.ToLower(strings.TrimSpace(value)))
	if !normalized.IsValid() {
		return "", gerror.Newf("invalid record store mutation action: %s", value)
	}
	return normalized, nil
}

// String returns the string representation of the access mode.
func (value AccessMode) String() string { return string(value) }

// IsValid reports whether the access mode is supported.
func (value AccessMode) IsValid() bool {
	switch value {
	case AccessModeRequest, AccessModeSystem, AccessModeBoth:
		return true
	default:
		return false
	}
}

// ParseAccessMode parses one raw value into a typed access mode.
func ParseAccessMode(value string) (AccessMode, error) {
	normalized := AccessMode(strings.ToLower(strings.TrimSpace(value)))
	if !normalized.IsValid() {
		return "", gerror.Newf("invalid record store access mode: %s", value)
	}
	return normalized, nil
}

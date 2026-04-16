// This file defines the typed plugindb enums shared by guest builders and
// host-side execution components.

package shared

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// DataPlanAction represents one governed data plan action.
type DataPlanAction string

const (
	// DataPlanActionList lists records from one authorized table.
	DataPlanActionList DataPlanAction = "list"
	// DataPlanActionGet reads one record by key from one authorized table.
	DataPlanActionGet DataPlanAction = "get"
	// DataPlanActionCount counts records from one authorized table.
	DataPlanActionCount DataPlanAction = "count"
	// DataPlanActionCreate inserts one record into one authorized table.
	DataPlanActionCreate DataPlanAction = "create"
	// DataPlanActionUpdate updates one record in one authorized table.
	DataPlanActionUpdate DataPlanAction = "update"
	// DataPlanActionDelete deletes one record from one authorized table.
	DataPlanActionDelete DataPlanAction = "delete"
	// DataPlanActionTransaction executes one structured mutation transaction.
	DataPlanActionTransaction DataPlanAction = "transaction"
)

// DataFilterOperator represents one supported filter operator.
type DataFilterOperator string

const (
	// DataFilterOperatorEQ compares one field by equality.
	DataFilterOperatorEQ DataFilterOperator = "eq"
	// DataFilterOperatorIN compares one field against a value list.
	DataFilterOperatorIN DataFilterOperator = "in"
	// DataFilterOperatorLike compares one field by wildcard matching.
	DataFilterOperatorLike DataFilterOperator = "like"
)

// DataOrderDirection represents one supported order direction.
type DataOrderDirection string

const (
	// DataOrderDirectionASC orders records in ascending order.
	DataOrderDirectionASC DataOrderDirection = "asc"
	// DataOrderDirectionDESC orders records in descending order.
	DataOrderDirectionDESC DataOrderDirection = "desc"
)

// DataMutationAction represents one transaction mutation action.
type DataMutationAction string

const (
	// DataMutationActionCreate enqueues one insert mutation.
	DataMutationActionCreate DataMutationAction = "create"
	// DataMutationActionUpdate enqueues one update mutation.
	DataMutationActionUpdate DataMutationAction = "update"
	// DataMutationActionDelete enqueues one delete mutation.
	DataMutationActionDelete DataMutationAction = "delete"
)

// DataAccessMode represents one runtime access requirement for a table contract.
type DataAccessMode string

const (
	// DataAccessModeRequest requires a request-bound execution context.
	DataAccessModeRequest DataAccessMode = "request"
	// DataAccessModeSystem allows a system-bound execution context.
	DataAccessModeSystem DataAccessMode = "system"
	// DataAccessModeBoth allows both request-bound and system-bound execution contexts.
	DataAccessModeBoth DataAccessMode = "both"
)

// String returns the string representation of the plan action.
func (value DataPlanAction) String() string { return string(value) }

// IsValid reports whether the plan action is one of the supported constants.
func (value DataPlanAction) IsValid() bool {
	switch value {
	case DataPlanActionList,
		DataPlanActionGet,
		DataPlanActionCount,
		DataPlanActionCreate,
		DataPlanActionUpdate,
		DataPlanActionDelete,
		DataPlanActionTransaction:
		return true
	default:
		return false
	}
}

// ParseDataPlanAction parses one raw value into a typed plan action.
func ParseDataPlanAction(value string) (DataPlanAction, error) {
	normalized := DataPlanAction(strings.ToLower(strings.TrimSpace(value)))
	if !normalized.IsValid() {
		return "", gerror.Newf("invalid data plan action: %s", value)
	}
	return normalized, nil
}

// String returns the string representation of the filter operator.
func (value DataFilterOperator) String() string { return string(value) }

// IsValid reports whether the filter operator is supported.
func (value DataFilterOperator) IsValid() bool {
	switch value {
	case DataFilterOperatorEQ, DataFilterOperatorIN, DataFilterOperatorLike:
		return true
	default:
		return false
	}
}

// ParseDataFilterOperator parses one raw value into a typed filter operator.
func ParseDataFilterOperator(value string) (DataFilterOperator, error) {
	normalized := DataFilterOperator(strings.ToLower(strings.TrimSpace(value)))
	if !normalized.IsValid() {
		return "", gerror.Newf("invalid data filter operator: %s", value)
	}
	return normalized, nil
}

// String returns the string representation of the order direction.
func (value DataOrderDirection) String() string { return string(value) }

// IsValid reports whether the order direction is supported.
func (value DataOrderDirection) IsValid() bool {
	switch value {
	case DataOrderDirectionASC, DataOrderDirectionDESC:
		return true
	default:
		return false
	}
}

// ParseDataOrderDirection parses one raw value into a typed order direction.
func ParseDataOrderDirection(value string) (DataOrderDirection, error) {
	normalized := DataOrderDirection(strings.ToLower(strings.TrimSpace(value)))
	if !normalized.IsValid() {
		return "", gerror.Newf("invalid data order direction: %s", value)
	}
	return normalized, nil
}

// String returns the string representation of the mutation action.
func (value DataMutationAction) String() string { return string(value) }

// IsValid reports whether the mutation action is supported.
func (value DataMutationAction) IsValid() bool {
	switch value {
	case DataMutationActionCreate, DataMutationActionUpdate, DataMutationActionDelete:
		return true
	default:
		return false
	}
}

// ParseDataMutationAction parses one raw value into a typed mutation action.
func ParseDataMutationAction(value string) (DataMutationAction, error) {
	normalized := DataMutationAction(strings.ToLower(strings.TrimSpace(value)))
	if !normalized.IsValid() {
		return "", gerror.Newf("invalid data mutation action: %s", value)
	}
	return normalized, nil
}

// String returns the string representation of the access mode.
func (value DataAccessMode) String() string { return string(value) }

// IsValid reports whether the access mode is supported.
func (value DataAccessMode) IsValid() bool {
	switch value {
	case DataAccessModeRequest, DataAccessModeSystem, DataAccessModeBoth:
		return true
	default:
		return false
	}
}

// ParseDataAccessMode parses one raw value into a typed access mode.
func ParseDataAccessMode(value string) (DataAccessMode, error) {
	normalized := DataAccessMode(strings.ToLower(strings.TrimSpace(value)))
	if !normalized.IsValid() {
		return "", gerror.Newf("invalid data access mode: %s", value)
	}
	return normalized, nil
}

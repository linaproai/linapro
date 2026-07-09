// This file defines plugin resource declaration value objects and normalizers
// that are shared by catalog validation, integration, and data host dispatch.

package plugintypes

import "strings"

// ResourceSpecType defines the supported plugin backend resource declaration type.
type ResourceSpecType string

// ResourceFilterOperator defines supported resource filter operators.
type ResourceFilterOperator string

// ResourceOrderDirection defines supported ordering directions in resource specs.
type ResourceOrderDirection string

// ResourceOperation defines the supported structured data operations for one resource.
type ResourceOperation string

// ResourceAccessMode defines which execution contexts may invoke one resource.
type ResourceAccessMode string

const (
	// ResourceSpecTypeTableList declares a table-list backend resource.
	ResourceSpecTypeTableList ResourceSpecType = "table-list"

	// ResourceFilterOperatorEQ declares equality filtering.
	ResourceFilterOperatorEQ ResourceFilterOperator = "eq"
	// ResourceFilterOperatorLike declares LIKE filtering.
	ResourceFilterOperatorLike ResourceFilterOperator = "like"
	// ResourceFilterOperatorGTEDate declares lower-bound date filtering.
	ResourceFilterOperatorGTEDate ResourceFilterOperator = "gte-date"
	// ResourceFilterOperatorLTEDate declares upper-bound date filtering.
	ResourceFilterOperatorLTEDate ResourceFilterOperator = "lte-date"

	// ResourceOrderDirectionASC declares ascending order.
	ResourceOrderDirectionASC ResourceOrderDirection = "asc"
	// ResourceOrderDirectionDESC declares descending order.
	ResourceOrderDirectionDESC ResourceOrderDirection = "desc"

	// ResourceOperationQuery declares list/query access.
	ResourceOperationQuery ResourceOperation = "query"
	// ResourceOperationGet declares single-record access.
	ResourceOperationGet ResourceOperation = "get"
	// ResourceOperationCreate declares create access.
	ResourceOperationCreate ResourceOperation = "create"
	// ResourceOperationUpdate declares update access.
	ResourceOperationUpdate ResourceOperation = "update"
	// ResourceOperationDelete declares delete access.
	ResourceOperationDelete ResourceOperation = "delete"
	// ResourceOperationTransaction declares transactional access.
	ResourceOperationTransaction ResourceOperation = "transaction"

	// ResourceAccessModeRequest allows request-context access.
	ResourceAccessModeRequest ResourceAccessMode = "request"
	// ResourceAccessModeSystem allows system-context access.
	ResourceAccessModeSystem ResourceAccessMode = "system"
	// ResourceAccessModeBoth allows request and system access.
	ResourceAccessModeBoth ResourceAccessMode = "both"
)

// String returns the canonical resource spec type value.
func (value ResourceSpecType) String() string { return string(value) }

// String returns the canonical resource filter-operator value.
func (value ResourceFilterOperator) String() string { return string(value) }

// String returns the canonical resource order-direction value.
func (value ResourceOrderDirection) String() string { return string(value) }

// String returns the canonical resource operation value.
func (value ResourceOperation) String() string { return string(value) }

// String returns the canonical resource access-mode value.
func (value ResourceAccessMode) String() string { return string(value) }

// NormalizeResourceSpecType maps a raw string to the canonical ResourceSpecType constant.
func NormalizeResourceSpecType(value string) ResourceSpecType {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ResourceSpecTypeTableList.String():
		return ResourceSpecTypeTableList
	default:
		return ResourceSpecType("")
	}
}

// NormalizeResourceFilterOperator maps a raw string to the canonical ResourceFilterOperator constant.
func NormalizeResourceFilterOperator(value string) ResourceFilterOperator {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ResourceFilterOperatorEQ.String():
		return ResourceFilterOperatorEQ
	case ResourceFilterOperatorLike.String():
		return ResourceFilterOperatorLike
	case ResourceFilterOperatorGTEDate.String():
		return ResourceFilterOperatorGTEDate
	case ResourceFilterOperatorLTEDate.String():
		return ResourceFilterOperatorLTEDate
	default:
		return ResourceFilterOperator("")
	}
}

// NormalizeResourceOrderDirection maps a raw string to the canonical ResourceOrderDirection constant.
func NormalizeResourceOrderDirection(value string) ResourceOrderDirection {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ResourceOrderDirectionASC.String():
		return ResourceOrderDirectionASC
	case ResourceOrderDirectionDESC.String():
		return ResourceOrderDirectionDESC
	default:
		return ResourceOrderDirection("")
	}
}

// NormalizeResourceOperation maps a raw string to the canonical ResourceOperation constant.
func NormalizeResourceOperation(value string) ResourceOperation {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ResourceOperationQuery.String():
		return ResourceOperationQuery
	case ResourceOperationGet.String():
		return ResourceOperationGet
	case ResourceOperationCreate.String():
		return ResourceOperationCreate
	case ResourceOperationUpdate.String():
		return ResourceOperationUpdate
	case ResourceOperationDelete.String():
		return ResourceOperationDelete
	case ResourceOperationTransaction.String():
		return ResourceOperationTransaction
	default:
		return ResourceOperation("")
	}
}

// NormalizeResourceAccessMode maps a raw string to the canonical ResourceAccessMode constant.
func NormalizeResourceAccessMode(value string) ResourceAccessMode {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", ResourceAccessModeRequest.String():
		return ResourceAccessModeRequest
	case ResourceAccessModeSystem.String():
		return ResourceAccessModeSystem
	case ResourceAccessModeBoth.String():
		return ResourceAccessModeBoth
	default:
		return ResourceAccessMode("")
	}
}

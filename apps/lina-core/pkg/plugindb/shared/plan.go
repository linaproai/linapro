// This file defines the typed plugindb query-plan model shared by guest and
// host implementations.

package shared

// DataQueryPlan represents one governed single-table data request.
type DataQueryPlan struct {
	// Table is the authorized target table name.
	Table string `json:"table"`
	// Action is the governed action to execute.
	Action DataPlanAction `json:"action"`
	// Fields contains the requested field projection.
	Fields []string `json:"fields,omitempty"`
	// Filters contains the requested filter clauses.
	Filters []*DataFilter `json:"filters,omitempty"`
	// Orders contains the requested order-by clauses.
	Orders []*DataOrder `json:"orders,omitempty"`
	// Page contains the optional paging window.
	Page *DataPagination `json:"page,omitempty"`
	// KeyJSON contains the JSON-encoded key value for get/update/delete.
	KeyJSON []byte `json:"keyJson,omitempty"`
	// RecordJSON contains the JSON-encoded input record for create/update.
	RecordJSON []byte `json:"recordJson,omitempty"`
	// Transaction contains the structured transaction payload.
	Transaction *DataTransactionPlan `json:"transaction,omitempty"`
}

// DataFilter represents one field-level filter clause.
type DataFilter struct {
	// Field is the logical field name declared by the governed table contract.
	Field string `json:"field"`
	// Operator is the typed comparison operator.
	Operator DataFilterOperator `json:"operator"`
	// ValueJSON contains one JSON-encoded scalar value.
	ValueJSON []byte `json:"valueJson,omitempty"`
	// ValuesJSON contains one or more JSON-encoded values for list operators.
	ValuesJSON [][]byte `json:"valuesJson,omitempty"`
}

// DataOrder represents one order-by clause.
type DataOrder struct {
	// Field is the logical field name declared by the governed table contract.
	Field string `json:"field"`
	// Direction is the typed order direction.
	Direction DataOrderDirection `json:"direction"`
}

// DataPagination represents one requested page window.
type DataPagination struct {
	// PageNum is the 1-based page number.
	PageNum int32 `json:"pageNum,omitempty"`
	// PageSize is the requested page size.
	PageSize int32 `json:"pageSize,omitempty"`
}

// DataTransactionPlan represents one structured mutation transaction payload.
type DataTransactionPlan struct {
	// Operations is the ordered list of mutation operations.
	Operations []*DataMutationPlan `json:"operations,omitempty"`
}

// DataMutationPlan represents one transaction mutation operation.
type DataMutationPlan struct {
	// Action is the typed mutation action.
	Action DataMutationAction `json:"action"`
	// KeyJSON contains the JSON-encoded key value for update/delete.
	KeyJSON []byte `json:"keyJson,omitempty"`
	// RecordJSON contains the JSON-encoded input record for create/update.
	RecordJSON []byte `json:"recordJson,omitempty"`
}

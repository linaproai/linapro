// Package plan defines the typed record store capability query-plan model shared by guest
// helpers and host-side execution components.
package plan

// QueryPlan represents one governed single-table record store request.
type QueryPlan struct {
	// Table is the authorized target table name.
	Table string `json:"table"`
	// Action is the governed action to execute.
	Action PlanAction `json:"action"`
	// Fields contains the requested field projection.
	Fields []string `json:"fields,omitempty"`
	// Filters contains the requested filter clauses.
	Filters []*Filter `json:"filters,omitempty"`
	// Orders contains the requested order-by clauses.
	Orders []*Order `json:"orders,omitempty"`
	// Page contains the optional paging window.
	Page *Pagination `json:"page,omitempty"`
	// KeyJSON contains the JSON-encoded key value for get/update/delete.
	KeyJSON []byte `json:"keyJson,omitempty"`
	// RecordJSON contains the JSON-encoded input record for create/update.
	RecordJSON []byte `json:"recordJson,omitempty"`
	// Transaction contains the structured transaction payload.
	Transaction *TransactionPlan `json:"transaction,omitempty"`
}

// Filter represents one field-level filter clause.
type Filter struct {
	// Field is the logical field name declared by the governed table contract.
	Field string `json:"field"`
	// Operator is the typed comparison operator.
	Operator FilterOperator `json:"operator"`
	// ValueJSON contains one JSON-encoded scalar value.
	ValueJSON []byte `json:"valueJson,omitempty"`
	// ValuesJSON contains one or more JSON-encoded values for list operators.
	ValuesJSON [][]byte `json:"valuesJson,omitempty"`
}

// Order represents one order-by clause.
type Order struct {
	// Field is the logical field name declared by the governed table contract.
	Field string `json:"field"`
	// Direction is the typed order direction.
	Direction OrderDirection `json:"direction"`
}

// Pagination represents one requested page window.
type Pagination struct {
	// PageNum is the 1-based page number.
	PageNum int32 `json:"pageNum,omitempty"`
	// PageSize is the requested page size.
	PageSize int32 `json:"pageSize,omitempty"`
}

// TransactionPlan represents one structured mutation transaction payload.
type TransactionPlan struct {
	// Operations is the ordered list of mutation operations.
	Operations []*MutationPlan `json:"operations,omitempty"`
}

// MutationPlan represents one transaction mutation operation.
type MutationPlan struct {
	// Action is the typed mutation action.
	Action MutationAction `json:"action"`
	// KeyJSON contains the JSON-encoded key value for update/delete.
	KeyJSON []byte `json:"keyJson,omitempty"`
	// RecordJSON contains the JSON-encoded input record for create/update.
	RecordJSON []byte `json:"recordJson,omitempty"`
}

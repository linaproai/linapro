// Package plugindb exposes a governed ORM-style facade for dynamic plugins.
package plugindb

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugindb/shared"
)

// DB exposes the guest-side governed data builder entry.
type DB struct{}

// Query represents one single-table governed query builder.
type Query struct {
	table string
	plan  *shared.DataQueryPlan
	err   error
}

// MutationResult represents one governed mutation result.
type MutationResult struct {
	// AffectedRows is the number of rows affected by the mutation.
	AffectedRows int64
	// Key is the optional decoded key returned by the host.
	Key any
	// Record is the optional decoded record snapshot returned by the host.
	Record map[string]any
}

// Tx represents one governed mutation transaction builder.
type Tx struct {
	table      string
	operations []*shared.DataMutationPlan
	err        error
}

// TxQuery represents one transaction-scoped table mutation builder.
type TxQuery struct {
	tx      *Tx
	table   string
	keyJSON []byte
	err     error
}

// Open returns one governed data facade for the current plugin.
func Open() *DB {
	return &DB{}
}

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

// Table selects the single transaction table and returns one mutation builder.
func (tx *Tx) Table(table string) *TxQuery {
	normalizedTable := strings.TrimSpace(table)
	if tx.err == nil {
		switch {
		case normalizedTable == "":
			tx.err = gerror.New("plugindb transaction table cannot be empty")
		case tx.table == "":
			tx.table = normalizedTable
		case tx.table != normalizedTable:
			tx.err = gerror.Newf("plugindb transaction only supports one table per transaction: %s != %s", tx.table, normalizedTable)
		}
	}
	return &TxQuery{tx: tx, table: normalizedTable, err: tx.err}
}

// WhereKey sets the key used by update/delete in a transaction.
func (q *TxQuery) WhereKey(key any) *TxQuery {
	if q.err != nil {
		return q
	}
	keyJSON, err := shared.MarshalValueJSON(key)
	if err != nil {
		q.err = err
		q.tx.err = err
		return q
	}
	q.keyJSON = keyJSON
	return q
}

// Insert appends one insert mutation to the transaction.
func (q *TxQuery) Insert(record map[string]any) (*MutationResult, error) {
	return q.enqueueMutation(shared.DataMutationActionCreate, nil, record)
}

// Update appends one update mutation to the transaction.
func (q *TxQuery) Update(record map[string]any) (*MutationResult, error) {
	return q.enqueueMutation(shared.DataMutationActionUpdate, q.keyJSON, record)
}

// Delete appends one delete mutation to the transaction.
func (q *TxQuery) Delete() (*MutationResult, error) {
	return q.enqueueMutation(shared.DataMutationActionDelete, q.keyJSON, nil)
}

func (q *TxQuery) enqueueMutation(action shared.DataMutationAction, keyJSON []byte, record map[string]any) (*MutationResult, error) {
	if q.err != nil {
		return nil, q.err
	}
	if q.tx == nil {
		return nil, gerror.New("plugindb transaction query is not initialized")
	}
	if !action.IsValid() {
		return nil, gerror.Newf("plugindb mutation action is invalid: %s", action)
	}
	if (action == shared.DataMutationActionUpdate || action == shared.DataMutationActionDelete) && len(keyJSON) == 0 {
		return nil, gerror.New("plugindb update/delete in transaction requires WhereKey")
	}
	recordJSON, err := shared.MarshalValueJSON(record)
	if err != nil {
		q.err = err
		q.tx.err = err
		return nil, err
	}
	q.tx.operations = append(q.tx.operations, &shared.DataMutationPlan{
		Action:     action,
		KeyJSON:    append([]byte(nil), keyJSON...),
		RecordJSON: recordJSON,
	})
	return &MutationResult{}, nil
}

// This file implements guest-side governed data transaction builder methods.

package plugindb

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugindb/shared"
)

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

// enqueueMutation validates and appends one structured mutation plan to the
// surrounding single-table transaction builder.
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

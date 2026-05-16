// Package plugindb exposes a governed ORM-style facade for dynamic plugins.
package plugindb

import "lina-core/pkg/plugindb/shared"

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

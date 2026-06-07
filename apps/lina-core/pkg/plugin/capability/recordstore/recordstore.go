// Package recordstore exposes a governed ORM-style facade for dynamic plugins.
package recordstore

import dataplan "lina-core/pkg/plugin/capability/recordstore/internal/plan"

// HostServiceInvoker dispatches one structured host-service request for record
// store execution. It lets pluginbridge/guest inject transport without making
// recordstore import the bridge guest package.
type HostServiceInvoker func(service string, method string, resourceRef string, table string, payload []byte) ([]byte, error)

// DB exposes the guest-side governed record store builder entry.
type DB struct {
	invoker HostServiceInvoker
}

// Query represents one single-table governed query builder.
type Query struct {
	table   string
	plan    *dataplan.QueryPlan
	err     error
	invoker HostServiceInvoker
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
	operations []*dataplan.MutationPlan
	err        error
	invoker    HostServiceInvoker
}

// TxQuery represents one transaction-scoped table mutation builder.
type TxQuery struct {
	tx      *Tx
	table   string
	keyJSON []byte
	err     error
}

// Open returns one governed record store facade for the current plugin.
func Open() *DB {
	return &DB{}
}

// OpenWithHostServiceInvoker returns one governed record store facade backed by
// the supplied host-service invoker.
func OpenWithHostServiceInvoker(invoker HostServiceInvoker) *DB {
	return &DB{invoker: invoker}
}

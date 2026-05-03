// This file creates the MySQL MEMORY backend implementation used by the
// default kvcache service adapter.

package mysqlmemory

// MySQLMemoryBackend implements KV cache operations using the host sys_kv_cache
// MEMORY table.
type MySQLMemoryBackend struct{}

// NewMySQLMemoryBackend creates one MySQL MEMORY backend implementation.
func NewMySQLMemoryBackend() *MySQLMemoryBackend {
	return &MySQLMemoryBackend{}
}

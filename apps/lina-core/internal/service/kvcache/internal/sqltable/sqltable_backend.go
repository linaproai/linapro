// This file creates the SQL table backend implementation used by the
// default kvcache service adapter.

package sqltable

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
)

// SQLTableBackend implements KV cache operations using the host sys_kv_cache
// SQL table.
type SQLTableBackend struct {
	db gdb.DB // db optionally overrides the default DAO database for tests.
}

// NewSQLTableBackend creates one SQL table backend implementation.
func NewSQLTableBackend() *SQLTableBackend {
	return &SQLTableBackend{}
}

// NewSQLTableBackendWithDB creates one backend bound to a specific database.
// It is intended for package tests that verify the backend against multiple
// SQL dialects without mutating the process-wide GoFrame database config.
func NewSQLTableBackendWithDB(db gdb.DB) *SQLTableBackend {
	return &SQLTableBackend{db: db}
}

// model returns the sys_kv_cache model for the backend's active database.
func (b *SQLTableBackend) model(ctx context.Context) *gdb.Model {
	if b != nil && b.db != nil {
		return b.db.Model(dao.SysKvCache.Table()).Safe().Ctx(ctx)
	}
	return dao.SysKvCache.Ctx(ctx)
}

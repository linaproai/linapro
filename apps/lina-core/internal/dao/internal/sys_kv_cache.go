// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysKvCacheDao is the data access object for the table sys_kv_cache.
type SysKvCacheDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysKvCacheColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysKvCacheColumns defines and stores column names for the table sys_kv_cache.
type SysKvCacheColumns struct {
	Id         string // Primary key ID
	TenantId   string // Owning tenant ID, 0 means PLATFORM
	OwnerType  string // Owner type: plugin=dynamic plugin, module=host module
	OwnerKey   string // Owner key: plugin ID or module name
	Namespace  string // Cache namespace mapped to the host-cache resource identifier
	CacheKey   string // Cache key
	ValueKind  string // Value type: 1=string, 2=integer
	ValueBytes string // Cache byte value used by get/set
	ValueInt   string // Cache integer value used by incr
	ExpireAt   string // Expiration time, NULL means never expires
	CreatedAt  string // Creation time
	UpdatedAt  string // Update time
}

// sysKvCacheColumns holds the columns for the table sys_kv_cache.
var sysKvCacheColumns = SysKvCacheColumns{
	Id:         "id",
	TenantId:   "tenant_id",
	OwnerType:  "owner_type",
	OwnerKey:   "owner_key",
	Namespace:  "namespace",
	CacheKey:   "cache_key",
	ValueKind:  "value_kind",
	ValueBytes: "value_bytes",
	ValueInt:   "value_int",
	ExpireAt:   "expire_at",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
}

// NewSysKvCacheDao creates and returns a new DAO object for table data access.
func NewSysKvCacheDao(handlers ...gdb.ModelHandler) *SysKvCacheDao {
	return &SysKvCacheDao{
		group:    "default",
		table:    "sys_kv_cache",
		columns:  sysKvCacheColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysKvCacheDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysKvCacheDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysKvCacheDao) Columns() SysKvCacheColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysKvCacheDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysKvCacheDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *SysKvCacheDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

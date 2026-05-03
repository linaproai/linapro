// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysCacheRevisionDao is the data access object for the table sys_cache_revision.
type SysCacheRevisionDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  SysCacheRevisionColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// SysCacheRevisionColumns defines and stores column names for the table sys_cache_revision.
type SysCacheRevisionColumns struct {
	Id        string // Primary key ID
	Domain    string // Cache domain, such as runtime-config, permission-access, or plugin-runtime
	Scope     string // Explicit invalidation scope, such as global, plugin:<id>, locale:<locale>, or user:<id>
	Revision  string // Monotonic cache revision for this domain and scope
	Reason    string // Latest change reason used for diagnostics
	CreatedAt string // Creation time
	UpdatedAt string // Update time
}

// sysCacheRevisionColumns holds the columns for the table sys_cache_revision.
var sysCacheRevisionColumns = SysCacheRevisionColumns{
	Id:        "id",
	Domain:    "domain",
	Scope:     "scope",
	Revision:  "revision",
	Reason:    "reason",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewSysCacheRevisionDao creates and returns a new DAO object for table data access.
func NewSysCacheRevisionDao(handlers ...gdb.ModelHandler) *SysCacheRevisionDao {
	return &SysCacheRevisionDao{
		group:    "default",
		table:    "sys_cache_revision",
		columns:  sysCacheRevisionColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysCacheRevisionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysCacheRevisionDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysCacheRevisionDao) Columns() SysCacheRevisionColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysCacheRevisionDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysCacheRevisionDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysCacheRevisionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

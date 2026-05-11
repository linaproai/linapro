// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysJobGroupDao is the data access object for the table sys_job_group.
type SysJobGroupDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysJobGroupColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysJobGroupColumns defines and stores column names for the table sys_job_group.
type SysJobGroupColumns struct {
	Id        string // Job group ID
	TenantId  string // Owning tenant ID, 0 means PLATFORM
	Code      string // Group code
	Name      string // Group name
	Remark    string // Remark
	SortOrder string // Display order
	IsDefault string // Default group flag: 1=yes, 0=no
	CreatedAt string // Creation time
	UpdatedAt string // Update time
	DeletedAt string // Deletion time
}

// sysJobGroupColumns holds the columns for the table sys_job_group.
var sysJobGroupColumns = SysJobGroupColumns{
	Id:        "id",
	TenantId:  "tenant_id",
	Code:      "code",
	Name:      "name",
	Remark:    "remark",
	SortOrder: "sort_order",
	IsDefault: "is_default",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewSysJobGroupDao creates and returns a new DAO object for table data access.
func NewSysJobGroupDao(handlers ...gdb.ModelHandler) *SysJobGroupDao {
	return &SysJobGroupDao{
		group:    "default",
		table:    "sys_job_group",
		columns:  sysJobGroupColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysJobGroupDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysJobGroupDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysJobGroupDao) Columns() SysJobGroupColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysJobGroupDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysJobGroupDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysJobGroupDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

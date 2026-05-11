// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserDeptDao is the data access object for the table plugin_org_center_user_dept.
type UserDeptDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  UserDeptColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// UserDeptColumns defines and stores column names for the table plugin_org_center_user_dept.
type UserDeptColumns struct {
	TenantId string // Owning tenant ID, 0 means PLATFORM
	UserId   string // User ID
	DeptId   string // Department ID
}

// userDeptColumns holds the columns for the table plugin_org_center_user_dept.
var userDeptColumns = UserDeptColumns{
	TenantId: "tenant_id",
	UserId:   "user_id",
	DeptId:   "dept_id",
}

// NewUserDeptDao creates and returns a new DAO object for table data access.
func NewUserDeptDao(handlers ...gdb.ModelHandler) *UserDeptDao {
	return &UserDeptDao{
		group:    "default",
		table:    "plugin_org_center_user_dept",
		columns:  userDeptColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserDeptDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserDeptDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserDeptDao) Columns() UserDeptColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserDeptDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserDeptDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserDeptDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserPostDao is the data access object for the table plugin_org_center_user_post.
type UserPostDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  UserPostColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// UserPostColumns defines and stores column names for the table plugin_org_center_user_post.
type UserPostColumns struct {
	TenantId string // Owning tenant ID, 0 means PLATFORM
	UserId   string // User ID
	PostId   string // Post ID
}

// userPostColumns holds the columns for the table plugin_org_center_user_post.
var userPostColumns = UserPostColumns{
	TenantId: "tenant_id",
	UserId:   "user_id",
	PostId:   "post_id",
}

// NewUserPostDao creates and returns a new DAO object for table data access.
func NewUserPostDao(handlers ...gdb.ModelHandler) *UserPostDao {
	return &UserPostDao{
		group:    "default",
		table:    "plugin_org_center_user_post",
		columns:  userPostColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserPostDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserPostDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserPostDao) Columns() UserPostColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserPostDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserPostDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserPostDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

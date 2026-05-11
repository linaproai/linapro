// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserMembershipDao is the data access object for the table plugin_multi_tenant_user_membership.
type UserMembershipDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  UserMembershipColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// UserMembershipColumns defines and stores column names for the table plugin_multi_tenant_user_membership.
type UserMembershipColumns struct {
	Id        string //
	UserId    string //
	TenantId  string //
	Status    string //
	JoinedAt  string //
	CreatedBy string //
	UpdatedBy string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// userMembershipColumns holds the columns for the table plugin_multi_tenant_user_membership.
var userMembershipColumns = UserMembershipColumns{
	Id:        "id",
	UserId:    "user_id",
	TenantId:  "tenant_id",
	Status:    "status",
	JoinedAt:  "joined_at",
	CreatedBy: "created_by",
	UpdatedBy: "updated_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewUserMembershipDao creates and returns a new DAO object for table data access.
func NewUserMembershipDao(handlers ...gdb.ModelHandler) *UserMembershipDao {
	return &UserMembershipDao{
		group:    "default",
		table:    "plugin_multi_tenant_user_membership",
		columns:  userMembershipColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserMembershipDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserMembershipDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserMembershipDao) Columns() UserMembershipColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserMembershipDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserMembershipDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserMembershipDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TenantDao is the data access object for the table plugin_multi_tenant_tenant.
type TenantDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  TenantColumns      // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// TenantColumns defines and stores column names for the table plugin_multi_tenant_tenant.
type TenantColumns struct {
	Id        string //
	Code      string //
	Name      string //
	Status    string //
	Remark    string //
	CreatedBy string //
	UpdatedBy string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// tenantColumns holds the columns for the table plugin_multi_tenant_tenant.
var tenantColumns = TenantColumns{
	Id:        "id",
	Code:      "code",
	Name:      "name",
	Status:    "status",
	Remark:    "remark",
	CreatedBy: "created_by",
	UpdatedBy: "updated_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewTenantDao creates and returns a new DAO object for table data access.
func NewTenantDao(handlers ...gdb.ModelHandler) *TenantDao {
	return &TenantDao{
		group:    "default",
		table:    "plugin_multi_tenant_tenant",
		columns:  tenantColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TenantDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TenantDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TenantDao) Columns() TenantColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TenantDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TenantDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TenantDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

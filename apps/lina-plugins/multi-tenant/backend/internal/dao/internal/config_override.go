// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ConfigOverrideDao is the data access object for the table plugin_multi_tenant_config_override.
type ConfigOverrideDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  ConfigOverrideColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// ConfigOverrideColumns defines and stores column names for the table plugin_multi_tenant_config_override.
type ConfigOverrideColumns struct {
	Id          string //
	TenantId    string //
	ConfigKey   string //
	ConfigValue string //
	Enabled     string //
	CreatedAt   string //
	UpdatedAt   string //
	DeletedAt   string //
}

// configOverrideColumns holds the columns for the table plugin_multi_tenant_config_override.
var configOverrideColumns = ConfigOverrideColumns{
	Id:          "id",
	TenantId:    "tenant_id",
	ConfigKey:   "config_key",
	ConfigValue: "config_value",
	Enabled:     "enabled",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
}

// NewConfigOverrideDao creates and returns a new DAO object for table data access.
func NewConfigOverrideDao(handlers ...gdb.ModelHandler) *ConfigOverrideDao {
	return &ConfigOverrideDao{
		group:    "default",
		table:    "plugin_multi_tenant_config_override",
		columns:  configOverrideColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ConfigOverrideDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ConfigOverrideDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ConfigOverrideDao) Columns() ConfigOverrideColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ConfigOverrideDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ConfigOverrideDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ConfigOverrideDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

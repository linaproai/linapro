// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysPluginStateDao is the data access object for the table sys_plugin_state.
type SysPluginStateDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  SysPluginStateColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// SysPluginStateColumns defines and stores column names for the table sys_plugin_state.
type SysPluginStateColumns struct {
	Id         string // Primary key ID
	PluginId   string // Plugin unique identifier (kebab-case)
	TenantId   string // Plugin state tenant ID, 0 means platform/global state
	StateKey   string // State key
	StateValue string // State value with JSON support
	Enabled    string // Whether the plugin is enabled for the tenant
	CreatedAt  string // Creation time
	UpdatedAt  string // Update time
}

// sysPluginStateColumns holds the columns for the table sys_plugin_state.
var sysPluginStateColumns = SysPluginStateColumns{
	Id:         "id",
	PluginId:   "plugin_id",
	TenantId:   "tenant_id",
	StateKey:   "state_key",
	StateValue: "state_value",
	Enabled:    "enabled",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
}

// NewSysPluginStateDao creates and returns a new DAO object for table data access.
func NewSysPluginStateDao(handlers ...gdb.ModelHandler) *SysPluginStateDao {
	return &SysPluginStateDao{
		group:    "default",
		table:    "sys_plugin_state",
		columns:  sysPluginStateColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysPluginStateDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysPluginStateDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysPluginStateDao) Columns() SysPluginStateColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysPluginStateDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysPluginStateDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysPluginStateDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

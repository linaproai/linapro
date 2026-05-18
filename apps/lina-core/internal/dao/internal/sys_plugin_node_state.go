// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysPluginNodeStateDao is the data access object for the table sys_plugin_node_state.
type SysPluginNodeStateDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  SysPluginNodeStateColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// SysPluginNodeStateColumns defines and stores column names for the table sys_plugin_node_state.
type SysPluginNodeStateColumns struct {
	Id              string // Primary key ID
	PluginId        string // Plugin unique identifier (kebab-case)
	ReleaseId       string // Owning plugin release ID
	NodeKey         string // Node unique identifier
	DesiredState    string // Node desired state
	CurrentState    string // Node current state
	Generation      string // Plugin generation number
	LastHeartbeatAt string // Last heartbeat time
	ErrorMessage    string // Node error message
	CreatedAt       string // Creation time
	UpdatedAt       string // Update time
}

// sysPluginNodeStateColumns holds the columns for the table sys_plugin_node_state.
var sysPluginNodeStateColumns = SysPluginNodeStateColumns{
	Id:              "id",
	PluginId:        "plugin_id",
	ReleaseId:       "release_id",
	NodeKey:         "node_key",
	DesiredState:    "desired_state",
	CurrentState:    "current_state",
	Generation:      "generation",
	LastHeartbeatAt: "last_heartbeat_at",
	ErrorMessage:    "error_message",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
}

// NewSysPluginNodeStateDao creates and returns a new DAO object for table data access.
func NewSysPluginNodeStateDao(handlers ...gdb.ModelHandler) *SysPluginNodeStateDao {
	return &SysPluginNodeStateDao{
		group:    "default",
		table:    "sys_plugin_node_state",
		columns:  sysPluginNodeStateColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysPluginNodeStateDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysPluginNodeStateDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysPluginNodeStateDao) Columns() SysPluginNodeStateColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysPluginNodeStateDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysPluginNodeStateDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysPluginNodeStateDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysPluginResourceRefDao is the data access object for the table sys_plugin_resource_ref.
type SysPluginResourceRefDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  SysPluginResourceRefColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// SysPluginResourceRefColumns defines and stores column names for the table sys_plugin_resource_ref.
type SysPluginResourceRefColumns struct {
	Id           string // Primary key ID
	PluginId     string // Plugin unique identifier (kebab-case)
	ReleaseId    string // Owning plugin release ID
	ResourceType string // Resource type: manifest/sql/frontend/menu/permission, etc.
	ResourceKey  string // Resource unique key
	ResourcePath string // Resource location metadata, empty by default and without concrete frontend or SQL paths
	OwnerType    string // Host object type: file/menu/route/slot, etc.
	OwnerKey     string // Stable host object identifier
	Remark       string // Remark
	CreatedAt    string // Creation time
	UpdatedAt    string // Update time
	DeletedAt    string // Deletion time
}

// sysPluginResourceRefColumns holds the columns for the table sys_plugin_resource_ref.
var sysPluginResourceRefColumns = SysPluginResourceRefColumns{
	Id:           "id",
	PluginId:     "plugin_id",
	ReleaseId:    "release_id",
	ResourceType: "resource_type",
	ResourceKey:  "resource_key",
	ResourcePath: "resource_path",
	OwnerType:    "owner_type",
	OwnerKey:     "owner_key",
	Remark:       "remark",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewSysPluginResourceRefDao creates and returns a new DAO object for table data access.
func NewSysPluginResourceRefDao(handlers ...gdb.ModelHandler) *SysPluginResourceRefDao {
	return &SysPluginResourceRefDao{
		group:    "default",
		table:    "sys_plugin_resource_ref",
		columns:  sysPluginResourceRefColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysPluginResourceRefDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysPluginResourceRefDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysPluginResourceRefDao) Columns() SysPluginResourceRefColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysPluginResourceRefDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysPluginResourceRefDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysPluginResourceRefDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

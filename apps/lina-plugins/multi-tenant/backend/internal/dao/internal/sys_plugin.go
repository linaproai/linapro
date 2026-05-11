// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysPluginDao is the data access object for the table sys_plugin.
type SysPluginDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysPluginColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysPluginColumns defines and stores column names for the table sys_plugin.
type SysPluginColumns struct {
	Id                      string // Primary key ID
	PluginId                string // Plugin unique identifier (kebab-case)
	Name                    string // Plugin name
	Version                 string // Plugin version
	Type                    string // Plugin top-level type: source/dynamic
	ScopeNature             string // Plugin scope nature: platform_only or tenant_aware
	InstallMode             string // Plugin install mode: global or tenant_scoped
	AutoEnableForNewTenants string // Platform policy: whether installed and enabled tenant-scoped plugins are enabled for new tenants automatically
	Installed               string // Installation status: 1=installed, 0=not installed
	Status                  string // Enablement status: 1=enabled, 0=disabled
	DesiredState            string // Host desired state: uninstalled/installed/enabled
	CurrentState            string // Host current state: uninstalled/installed/enabled/reconciling/failed
	Generation              string // Current host generation number
	ReleaseId               string // Current active host release ID
	ManifestPath            string // Plugin manifest file path
	Checksum                string // Plugin package checksum
	InstalledAt             string // Installation time
	EnabledAt               string // Last enabled time
	DisabledAt              string // Last disabled time
	Remark                  string // Remark
	CreatedAt               string // Creation time
	UpdatedAt               string // Update time
	DeletedAt               string // Deletion time
}

// sysPluginColumns holds the columns for the table sys_plugin.
var sysPluginColumns = SysPluginColumns{
	Id:                      "id",
	PluginId:                "plugin_id",
	Name:                    "name",
	Version:                 "version",
	Type:                    "type",
	ScopeNature:             "scope_nature",
	InstallMode:             "install_mode",
	AutoEnableForNewTenants: "auto_enable_for_new_tenants",
	Installed:               "installed",
	Status:                  "status",
	DesiredState:            "desired_state",
	CurrentState:            "current_state",
	Generation:              "generation",
	ReleaseId:               "release_id",
	ManifestPath:            "manifest_path",
	Checksum:                "checksum",
	InstalledAt:             "installed_at",
	EnabledAt:               "enabled_at",
	DisabledAt:              "disabled_at",
	Remark:                  "remark",
	CreatedAt:               "created_at",
	UpdatedAt:               "updated_at",
	DeletedAt:               "deleted_at",
}

// NewSysPluginDao creates and returns a new DAO object for table data access.
func NewSysPluginDao(handlers ...gdb.ModelHandler) *SysPluginDao {
	return &SysPluginDao{
		group:    "default",
		table:    "sys_plugin",
		columns:  sysPluginColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysPluginDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysPluginDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysPluginDao) Columns() SysPluginColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysPluginDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysPluginDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysPluginDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

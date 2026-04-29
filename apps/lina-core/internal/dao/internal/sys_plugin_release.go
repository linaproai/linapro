// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysPluginReleaseDao is the data access object for the table sys_plugin_release.
type SysPluginReleaseDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  SysPluginReleaseColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// SysPluginReleaseColumns defines and stores column names for the table sys_plugin_release.
type SysPluginReleaseColumns struct {
	Id               string // Primary key ID
	PluginId         string // Plugin unique identifier (kebab-case)
	ReleaseVersion   string // Plugin version
	Type             string // Plugin top-level type: source/dynamic
	RuntimeKind      string // Runtime artifact type (currently only wasm)
	SchemaVersion    string // plugin.yaml manifest schema version
	MinHostVersion   string // Minimum compatible host version
	MaxHostVersion   string // Maximum compatible host version
	Status           string // Release status: prepared/installed/active/uninstalled/failed
	ManifestPath     string // Plugin manifest path
	PackagePath      string // Plugin source directory or runtime artifact path
	Checksum         string // Plugin manifest or artifact checksum
	ManifestSnapshot string // Plugin manifest and resource summary snapshot in YAML, without concrete SQL or frontend file paths
	CreatedAt        string // Creation time
	UpdatedAt        string // Update time
	DeletedAt        string // Deletion time
}

// sysPluginReleaseColumns holds the columns for the table sys_plugin_release.
var sysPluginReleaseColumns = SysPluginReleaseColumns{
	Id:               "id",
	PluginId:         "plugin_id",
	ReleaseVersion:   "release_version",
	Type:             "type",
	RuntimeKind:      "runtime_kind",
	SchemaVersion:    "schema_version",
	MinHostVersion:   "min_host_version",
	MaxHostVersion:   "max_host_version",
	Status:           "status",
	ManifestPath:     "manifest_path",
	PackagePath:      "package_path",
	Checksum:         "checksum",
	ManifestSnapshot: "manifest_snapshot",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
	DeletedAt:        "deleted_at",
}

// NewSysPluginReleaseDao creates and returns a new DAO object for table data access.
func NewSysPluginReleaseDao(handlers ...gdb.ModelHandler) *SysPluginReleaseDao {
	return &SysPluginReleaseDao{
		group:    "default",
		table:    "sys_plugin_release",
		columns:  sysPluginReleaseColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysPluginReleaseDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysPluginReleaseDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysPluginReleaseDao) Columns() SysPluginReleaseColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysPluginReleaseDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysPluginReleaseDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysPluginReleaseDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

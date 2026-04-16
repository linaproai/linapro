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
	Id               string // 主键ID
	PluginId         string // 插件唯一标识（kebab-case）
	ReleaseVersion   string // 插件版本号
	Type             string // 插件一级类型（source/dynamic）
	RuntimeKind      string // 运行时产物类型（当前仅 wasm）
	SchemaVersion    string // plugin.yaml 清单 schema 版本
	MinHostVersion   string // 宿主最小兼容版本
	MaxHostVersion   string // 宿主最大兼容版本
	Status           string // release 状态（prepared/installed/active/uninstalled/failed）
	ManifestPath     string // 插件清单路径
	PackagePath      string // 插件源码目录或运行时产物路径
	Checksum         string // 插件清单或产物校验值
	ManifestSnapshot string // 插件清单与资源摘要快照（YAML，不保存具体 SQL/前端文件路径）
	CreatedAt        string // 创建时间
	UpdatedAt        string // 更新时间
	DeletedAt        string // 删除时间
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

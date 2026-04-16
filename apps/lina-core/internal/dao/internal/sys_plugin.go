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
	Id           string // 主键ID
	PluginId     string // 插件唯一标识（kebab-case）
	Name         string // 插件名称
	Version      string // 插件版本号
	Type         string // 插件一级类型（source/dynamic）
	Installed    string // 安装状态（1=已安装 0=未安装）
	Status       string // 启用状态（1=启用 0=禁用）
	DesiredState string // 宿主期望状态（uninstalled/installed/enabled）
	CurrentState string // 宿主当前状态（uninstalled/installed/enabled/reconciling/failed）
	Generation   string // 宿主当前生效代际号
	ReleaseId    string // 宿主当前生效 release ID
	ManifestPath string // 插件清单文件路径
	Checksum     string // 插件包校验值
	InstalledAt  string // 安装时间
	EnabledAt    string // 最后一次启用时间
	DisabledAt   string // 最后一次禁用时间
	Remark       string // 备注
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
	DeletedAt    string // 删除时间
}

// sysPluginColumns holds the columns for the table sys_plugin.
var sysPluginColumns = SysPluginColumns{
	Id:           "id",
	PluginId:     "plugin_id",
	Name:         "name",
	Version:      "version",
	Type:         "type",
	Installed:    "installed",
	Status:       "status",
	DesiredState: "desired_state",
	CurrentState: "current_state",
	Generation:   "generation",
	ReleaseId:    "release_id",
	ManifestPath: "manifest_path",
	Checksum:     "checksum",
	InstalledAt:  "installed_at",
	EnabledAt:    "enabled_at",
	DisabledAt:   "disabled_at",
	Remark:       "remark",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
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

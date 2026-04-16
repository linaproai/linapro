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
	Id           string // 主键ID
	PluginId     string // 插件唯一标识（kebab-case）
	ReleaseId    string // 所属插件 release ID
	ResourceType string // 资源类型（manifest/sql/frontend/menu/permission 等）
	ResourceKey  string // 资源唯一键
	ResourcePath string // 资源定位补充信息（默认留空，不保存具体前端/SQL 路径）
	OwnerType    string // 宿主对象类型（file/menu/route/slot 等）
	OwnerKey     string // 宿主对象稳定标识
	Remark       string // 备注
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
	DeletedAt    string // 删除时间
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

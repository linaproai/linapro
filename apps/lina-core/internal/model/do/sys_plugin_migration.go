// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginMigration is the golang structure of table sys_plugin_migration for DAO operations like Where/Data.
type SysPluginMigration struct {
	g.Meta         `orm:"table:sys_plugin_migration, do:true"`
	Id             any         // 主键ID
	PluginId       any         // 插件唯一标识（kebab-case）
	ReleaseId      any         // 所属插件 release ID
	Phase          any         // 迁移阶段（install/uninstall/upgrade/rollback）
	MigrationKey   any         // 迁移执行键（如 install-step-001，不保存具体 SQL 路径）
	Checksum       any         // 迁移文件校验值
	ExecutionOrder any         // 执行顺序（从1开始）
	Status         any         // 执行状态（pending/succeeded/failed/skipped）
	ExecutedAt     *gtime.Time // 执行时间
	ErrorMessage   any         // 失败原因或补充说明
	CreatedAt      *gtime.Time // 创建时间
	UpdatedAt      *gtime.Time // 更新时间
}

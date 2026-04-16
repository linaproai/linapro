// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginMigration is the golang structure for table sys_plugin_migration.
type SysPluginMigration struct {
	Id             int         `json:"id"             orm:"id"              description:"主键ID"`
	PluginId       string      `json:"pluginId"       orm:"plugin_id"       description:"插件唯一标识（kebab-case）"`
	ReleaseId      int         `json:"releaseId"      orm:"release_id"      description:"所属插件 release ID"`
	Phase          string      `json:"phase"          orm:"phase"           description:"迁移阶段（install/uninstall/upgrade/rollback）"`
	MigrationKey   string      `json:"migrationKey"   orm:"migration_key"   description:"迁移执行键（如 install-step-001，不保存具体 SQL 路径）"`
	Checksum       string      `json:"checksum"       orm:"checksum"        description:"迁移文件校验值"`
	ExecutionOrder int         `json:"executionOrder" orm:"execution_order" description:"执行顺序（从1开始）"`
	Status         string      `json:"status"         orm:"status"          description:"执行状态（pending/succeeded/failed/skipped）"`
	ExecutedAt     *gtime.Time `json:"executedAt"     orm:"executed_at"     description:"执行时间"`
	ErrorMessage   string      `json:"errorMessage"   orm:"error_message"   description:"失败原因或补充说明"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:"创建时间"`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"      description:"更新时间"`
}

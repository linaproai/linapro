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
	Id             any         // Primary key ID
	PluginId       any         // Plugin unique identifier (kebab-case)
	ReleaseId      any         // Owning plugin release ID
	Phase          any         // Migration phase: install/uninstall/upgrade/rollback
	MigrationKey   any         // Migration execution key such as install-step-001, without concrete SQL path
	Checksum       any         // Migration file checksum
	ExecutionOrder any         // Execution order starting from 1
	Status         any         // Execution status: pending/succeeded/failed/skipped
	ExecutedAt     *gtime.Time // Execution time
	ErrorMessage   any         // Failure reason or additional description
	CreatedAt      *gtime.Time // Creation time
	UpdatedAt      *gtime.Time // Update time
}

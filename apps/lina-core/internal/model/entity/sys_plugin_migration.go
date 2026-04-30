// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginMigration is the golang structure for table sys_plugin_migration.
type SysPluginMigration struct {
	Id             int         `json:"id"             orm:"id"              description:"Primary key ID"`
	PluginId       string      `json:"pluginId"       orm:"plugin_id"       description:"Plugin unique identifier (kebab-case)"`
	ReleaseId      int         `json:"releaseId"      orm:"release_id"      description:"Owning plugin release ID"`
	Phase          string      `json:"phase"          orm:"phase"           description:"Migration phase: install/uninstall/upgrade/rollback/mock"`
	MigrationKey   string      `json:"migrationKey"   orm:"migration_key"   description:"Migration execution key such as install-step-001, without concrete SQL path"`
	Checksum       string      `json:"checksum"       orm:"checksum"        description:"Migration file checksum"`
	ExecutionOrder int         `json:"executionOrder" orm:"execution_order" description:"Execution order starting from 1"`
	Status         string      `json:"status"         orm:"status"          description:"Execution status: pending/succeeded/failed/skipped"`
	ExecutedAt     *gtime.Time `json:"executedAt"     orm:"executed_at"     description:"Execution time"`
	ErrorMessage   string      `json:"errorMessage"   orm:"error_message"   description:"Failure reason or additional description"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:"Creation time"`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"      description:"Update time"`
}

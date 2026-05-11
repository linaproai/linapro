// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ConfigOverride is the golang structure for table config_override.
type ConfigOverride struct {
	Id          int64       `json:"id"          orm:"id"           description:""`
	TenantId    int64       `json:"tenantId"    orm:"tenant_id"    description:""`
	ConfigKey   string      `json:"configKey"   orm:"config_key"   description:""`
	ConfigValue string      `json:"configValue" orm:"config_value" description:""`
	Enabled     bool        `json:"enabled"     orm:"enabled"      description:""`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:""`
	DeletedAt   *gtime.Time `json:"deletedAt"   orm:"deleted_at"   description:""`
}

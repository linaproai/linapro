// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ConfigOverride is the golang structure of table plugin_multi_tenant_config_override for DAO operations like Where/Data.
type ConfigOverride struct {
	g.Meta      `orm:"table:plugin_multi_tenant_config_override, do:true"`
	Id          any         //
	TenantId    any         //
	ConfigKey   any         //
	ConfigValue any         //
	Enabled     any         //
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
	DeletedAt   *gtime.Time //
}

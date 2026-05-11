// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysConfig is the golang structure of table sys_config for DAO operations like Where/Data.
type SysConfig struct {
	g.Meta    `orm:"table:sys_config, do:true"`
	Id        any         // Config parameter ID
	TenantId  any         // Owning tenant ID, 0 means PLATFORM default
	Name      any         // Config parameter name
	Key       any         // Config parameter key
	Value     any         // Config parameter value
	IsBuiltin any         // Built-in record flag: 1=yes, 0=no
	Remark    any         // Remark
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Modification time
	DeletedAt *gtime.Time // Deletion time
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysKvCache is the golang structure of table sys_kv_cache for DAO operations like Where/Data.
type SysKvCache struct {
	g.Meta     `orm:"table:sys_kv_cache, do:true"`
	Id         any         // Primary key ID
	OwnerType  any         // Owner type: plugin=dynamic plugin, module=host module
	OwnerKey   any         // Owner key: plugin ID or module name
	Namespace  any         // Cache namespace mapped to the host-cache resource identifier
	CacheKey   any         // Cache key
	ValueKind  any         // Value type: 1=string, 2=integer
	ValueBytes any         // Cache byte value used by get/set
	ValueInt   any         // Cache integer value used by incr
	ExpireAt   *gtime.Time // Expiration time, NULL means never expires
	CreatedAt  *gtime.Time // Creation time
	UpdatedAt  *gtime.Time // Update time
}

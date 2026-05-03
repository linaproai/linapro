// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysCacheRevision is the golang structure of table sys_cache_revision for DAO operations like Where/Data.
type SysCacheRevision struct {
	g.Meta    `orm:"table:sys_cache_revision, do:true"`
	Id        any         // Primary key ID
	Domain    any         // Cache domain, such as runtime-config, permission-access, or plugin-runtime
	Scope     any         // Explicit invalidation scope, such as global, plugin:<id>, locale:<locale>, or user:<id>
	Revision  any         // Monotonic cache revision for this domain and scope
	Reason    any         // Latest change reason used for diagnostics
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Update time
}

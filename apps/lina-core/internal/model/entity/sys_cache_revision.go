// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysCacheRevision is the golang structure for table sys_cache_revision.
type SysCacheRevision struct {
	Id        int64       `json:"id"        orm:"id"         description:"Primary key ID"`
	TenantId  int         `json:"tenantId"  orm:"tenant_id"  description:"Owning tenant ID, 0 means PLATFORM"`
	Domain    string      `json:"domain"    orm:"domain"     description:"Cache domain, such as runtime-config, permission-access, or plugin-runtime"`
	Scope     string      `json:"scope"     orm:"scope"      description:"Explicit invalidation scope, such as global, plugin:<id>, locale:<locale>, or user:<id>"`
	Revision  int64       `json:"revision"  orm:"revision"   description:"Monotonic cache revision for this domain and scope"`
	Reason    string      `json:"reason"    orm:"reason"     description:"Latest change reason used for diagnostics"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
}

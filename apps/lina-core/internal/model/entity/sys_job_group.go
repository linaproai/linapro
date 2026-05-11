// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJobGroup is the golang structure for table sys_job_group.
type SysJobGroup struct {
	Id        int64       `json:"id"        orm:"id"         description:"Job group ID"`
	TenantId  int         `json:"tenantId"  orm:"tenant_id"  description:"Owning tenant ID, 0 means PLATFORM"`
	Code      string      `json:"code"      orm:"code"       description:"Group code"`
	Name      string      `json:"name"      orm:"name"       description:"Group name"`
	Remark    string      `json:"remark"    orm:"remark"     description:"Remark"`
	SortOrder int         `json:"sortOrder" orm:"sort_order" description:"Display order"`
	IsDefault int         `json:"isDefault" orm:"is_default" description:"Default group flag: 1=yes, 0=no"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}

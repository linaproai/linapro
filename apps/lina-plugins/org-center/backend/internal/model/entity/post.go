// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Post is the golang structure for table post.
type Post struct {
	Id        int         `json:"id"        orm:"id"         description:"Post ID"`
	TenantId  int         `json:"tenantId"  orm:"tenant_id"  description:"Owning tenant ID, 0 means PLATFORM"`
	DeptId    int         `json:"deptId"    orm:"dept_id"    description:"Owning department ID"`
	Code      string      `json:"code"      orm:"code"       description:"Post code"`
	Name      string      `json:"name"      orm:"name"       description:"Post name"`
	Sort      int         `json:"sort"      orm:"sort"       description:"Display order"`
	Status    int         `json:"status"    orm:"status"     description:"Status: 0=disabled, 1=enabled"`
	Remark    string      `json:"remark"    orm:"remark"     description:"Remark"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}

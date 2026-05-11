// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Dept is the golang structure for table dept.
type Dept struct {
	Id        int         `json:"id"        orm:"id"         description:"Department ID"`
	TenantId  int         `json:"tenantId"  orm:"tenant_id"  description:"Owning tenant ID, 0 means PLATFORM"`
	ParentId  int         `json:"parentId"  orm:"parent_id"  description:"Parent department ID"`
	Ancestors string      `json:"ancestors" orm:"ancestors"  description:"Ancestor list"`
	Name      string      `json:"name"      orm:"name"       description:"Department name"`
	Code      string      `json:"code"      orm:"code"       description:"Department code"`
	OrderNum  int         `json:"orderNum"  orm:"order_num"  description:"Display order"`
	Leader    int         `json:"leader"    orm:"leader"     description:"Leader user ID"`
	Phone     string      `json:"phone"     orm:"phone"      description:"Contact phone number"`
	Email     string      `json:"email"     orm:"email"      description:"Email address"`
	Status    int         `json:"status"    orm:"status"     description:"Status: 0=disabled, 1=enabled"`
	Remark    string      `json:"remark"    orm:"remark"     description:"Remark"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}

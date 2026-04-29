// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysRole is the golang structure for table sys_role.
type SysRole struct {
	Id        int         `json:"id"        orm:"id"         description:"Role ID"`
	Name      string      `json:"name"      orm:"name"       description:"Role name"`
	Key       string      `json:"key"       orm:"key"        description:"Permission key"`
	Sort      int         `json:"sort"      orm:"sort"       description:"Display order"`
	DataScope int         `json:"dataScope" orm:"data_scope" description:"Data scope: 1=all, 2=department, 3=self"`
	Status    int         `json:"status"    orm:"status"     description:"Status: 0=disabled, 1=enabled"`
	Remark    string      `json:"remark"    orm:"remark"     description:"Remark"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}

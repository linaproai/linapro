// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysUser is the golang structure for table sys_user.
type SysUser struct {
	Id        int         `json:"id"        orm:"id"         description:"User ID"`
	TenantId  int         `json:"tenantId"  orm:"tenant_id"  description:"Primary/default tenant ID, 0 means PLATFORM"`
	Username  string      `json:"username"  orm:"username"   description:"Username"`
	Password  string      `json:"password"  orm:"password"   description:"Password"`
	Nickname  string      `json:"nickname"  orm:"nickname"   description:"User nickname"`
	Email     string      `json:"email"     orm:"email"      description:"Email address"`
	Phone     string      `json:"phone"     orm:"phone"      description:"Mobile phone number"`
	Sex       int         `json:"sex"       orm:"sex"        description:"Gender: 0=unknown, 1=male, 2=female"`
	Avatar    string      `json:"avatar"    orm:"avatar"     description:"Avatar URL"`
	Status    int         `json:"status"    orm:"status"     description:"Status: 0=disabled, 1=enabled"`
	Remark    string      `json:"remark"    orm:"remark"     description:"Remark"`
	LoginDate *gtime.Time `json:"loginDate" orm:"login_date" description:"Last login time"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}

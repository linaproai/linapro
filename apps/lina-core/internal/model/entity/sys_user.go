// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysUser is the golang structure for table sys_user.
type SysUser struct {
	Id        int         `json:"id"        orm:"id"         description:"用户ID"`
	Username  string      `json:"username"  orm:"username"   description:"用户账号"`
	Password  string      `json:"password"  orm:"password"   description:"密码"`
	Nickname  string      `json:"nickname"  orm:"nickname"   description:"用户昵称"`
	Email     string      `json:"email"     orm:"email"      description:"邮箱"`
	Phone     string      `json:"phone"     orm:"phone"      description:"手机号码"`
	Sex       int         `json:"sex"       orm:"sex"        description:"性别（0未知 1男 2女）"`
	Avatar    string      `json:"avatar"    orm:"avatar"     description:"头像地址"`
	Status    int         `json:"status"    orm:"status"     description:"状态（0停用 1正常）"`
	Remark    string      `json:"remark"    orm:"remark"     description:"备注"`
	LoginDate *gtime.Time `json:"loginDate" orm:"login_date" description:"最后登录时间"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"删除时间"`
}

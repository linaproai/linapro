// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysLocker is the golang structure for table sys_locker.
type SysLocker struct {
	Id         int         `json:"id"         orm:"id"          description:"Primary key ID"`
	Name       string      `json:"name"       orm:"name"        description:"Lock name, unique identifier"`
	Reason     string      `json:"reason"     orm:"reason"      description:"Reason for acquiring the lock"`
	Holder     string      `json:"holder"     orm:"holder"      description:"Lock holder identifier (node name)"`
	ExpireTime *gtime.Time `json:"expireTime" orm:"expire_time" description:"Lock expiration time"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:"Creation time"`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:"Update time"`
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysLocker is the golang structure for table sys_locker.
type SysLocker struct {
	Id         int        `json:"id"         orm:"id"          description:"Primary key ID"`
	Name       string     `json:"name"       orm:"name"        description:"Lock name, unique identifier"`
	Reason     string     `json:"reason"     orm:"reason"      description:"Reason for acquiring the lock"`
	Holder     string     `json:"holder"     orm:"holder"      description:"Lock holder identifier (node name)"`
	ExpireTime *time.Time `json:"expireTime" orm:"expire_time" description:"Lock expiration time"`
	CreatedAt  *time.Time `json:"createdAt"  orm:"created_at"  description:"Creation time"`
	UpdatedAt  *time.Time `json:"updatedAt"  orm:"updated_at"  description:"Update time"`
}

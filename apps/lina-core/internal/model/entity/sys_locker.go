// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysLocker is the golang structure for table sys_locker.
type SysLocker struct {
	Id         uint        `json:"id"         orm:"id"          description:"主键ID"`
	Name       string      `json:"name"       orm:"name"        description:"锁名称，唯一标识"`
	Reason     string      `json:"reason"     orm:"reason"      description:"获取锁的原因"`
	Holder     string      `json:"holder"     orm:"holder"      description:"锁持有者标识（节点名）"`
	ExpireTime *gtime.Time `json:"expireTime" orm:"expire_time" description:"锁过期时间"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:"创建时间"`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:"更新时间"`
}

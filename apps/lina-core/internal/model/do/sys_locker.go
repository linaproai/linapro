// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysLocker is the golang structure of table sys_locker for DAO operations like Where/Data.
type SysLocker struct {
	g.Meta     `orm:"table:sys_locker, do:true"`
	Id         any         // 主键ID
	Name       any         // 锁名称，唯一标识
	Reason     any         // 获取锁的原因
	Holder     any         // 锁持有者标识（节点名）
	ExpireTime *gtime.Time // 锁过期时间
	CreatedAt  *gtime.Time // 创建时间
	UpdatedAt  *gtime.Time // 更新时间
}

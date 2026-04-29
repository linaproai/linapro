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
	Id         any         // Primary key ID
	Name       any         // Lock name, unique identifier
	Reason     any         // Reason for acquiring the lock
	Holder     any         // Lock holder identifier (node name)
	ExpireTime *gtime.Time // Lock expiration time
	CreatedAt  *gtime.Time // Creation time
	UpdatedAt  *gtime.Time // Update time
}

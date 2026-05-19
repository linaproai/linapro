// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysLocker is the golang structure of table sys_locker for DAO operations like Where/Data.
type SysLocker struct {
	g.Meta     `orm:"table:sys_locker, do:true"`
	Id         any        // Primary key ID
	Name       any        // Lock name, unique identifier
	Reason     any        // Reason for acquiring the lock
	Holder     any        // Lock holder identifier (node name)
	ExpireTime *time.Time // Lock expiration time
	CreatedAt  *time.Time // Creation time
	UpdatedAt  *time.Time // Update time
}

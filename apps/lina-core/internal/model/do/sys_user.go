// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysUser is the golang structure of table sys_user for DAO operations like Where/Data.
type SysUser struct {
	g.Meta    `orm:"table:sys_user, do:true"`
	Id        any        // User ID
	TenantId  any        // Primary/default tenant ID, 0 means PLATFORM
	Username  any        // Username
	Password  any        // Password
	Nickname  any        // User nickname
	Email     any        // Email address
	Phone     any        // Mobile phone number
	Sex       any        // Gender: 0=unknown, 1=male, 2=female
	Avatar    any        // Avatar URL
	Status    any        // Status: 0=disabled, 1=enabled
	Remark    any        // Remark
	LoginDate *time.Time // Last login time
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
	DeletedAt *time.Time // Deletion time
}

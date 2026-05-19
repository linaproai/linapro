// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysOnlineSession is the golang structure of table sys_online_session for DAO operations like Where/Data.
type SysOnlineSession struct {
	g.Meta         `orm:"table:sys_online_session, do:true"`
	TenantId       any        // Owning tenant ID, 0 means PLATFORM
	TokenId        any        // Session token ID (UUID)
	UserId         any        // User ID
	Username       any        // Login account
	DeptName       any        // Department name
	Ip             any        // Login IP
	Browser        any        // Browser
	Os             any        // Operating system
	LoginTime      *time.Time // Login time
	LastActiveTime *time.Time // Last active time
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysOnlineSession is the golang structure of table sys_online_session for DAO operations like Where/Data.
type SysOnlineSession struct {
	g.Meta         `orm:"table:sys_online_session, do:true"`
	TokenId        any         // Session token ID (UUID)
	UserId         any         // User ID
	Username       any         // Login account
	DeptName       any         // Department name
	Ip             any         // Login IP
	Browser        any         // Browser
	Os             any         // Operating system
	LoginTime      *gtime.Time // Login time
	LastActiveTime *gtime.Time // Last active time
}

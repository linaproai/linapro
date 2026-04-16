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
	TokenId        any         // 会话Token ID（UUID）
	UserId         any         // 用户ID
	Username       any         // 登录账号
	DeptName       any         // 部门名称
	Ip             any         // 登录IP
	Browser        any         // 浏览器
	Os             any         // 操作系统
	LoginTime      *gtime.Time // 登录时间
	LastActiveTime *gtime.Time // 最后活跃时间
}

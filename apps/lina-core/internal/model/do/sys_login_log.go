// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysLoginLog is the golang structure of table sys_login_log for DAO operations like Where/Data.
type SysLoginLog struct {
	g.Meta    `orm:"table:sys_login_log, do:true"`
	Id        any         // 日志ID
	UserName  any         // 登录账号
	Status    any         // 登录状态（0成功 1失败）
	Ip        any         // 登录IP地址
	Browser   any         // 浏览器类型
	Os        any         // 操作系统
	Msg       any         // 提示消息
	LoginTime *gtime.Time // 登录时间
}

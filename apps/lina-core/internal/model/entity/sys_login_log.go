// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysLoginLog is the golang structure for table sys_login_log.
type SysLoginLog struct {
	Id        int         `json:"id"        orm:"id"         description:"日志ID"`
	UserName  string      `json:"userName"  orm:"user_name"  description:"登录账号"`
	Status    int         `json:"status"    orm:"status"     description:"登录状态（0成功 1失败）"`
	Ip        string      `json:"ip"        orm:"ip"         description:"登录IP地址"`
	Browser   string      `json:"browser"   orm:"browser"    description:"浏览器类型"`
	Os        string      `json:"os"        orm:"os"         description:"操作系统"`
	Msg       string      `json:"msg"       orm:"msg"        description:"提示消息"`
	LoginTime *gtime.Time `json:"loginTime" orm:"login_time" description:"登录时间"`
}

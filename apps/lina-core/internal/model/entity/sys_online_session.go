// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysOnlineSession is the golang structure for table sys_online_session.
type SysOnlineSession struct {
	TokenId        string      `json:"tokenId"        orm:"token_id"         description:"会话Token ID（UUID）"`
	UserId         int         `json:"userId"         orm:"user_id"          description:"用户ID"`
	Username       string      `json:"username"       orm:"username"         description:"登录账号"`
	DeptName       string      `json:"deptName"       orm:"dept_name"        description:"部门名称"`
	Ip             string      `json:"ip"             orm:"ip"               description:"登录IP"`
	Browser        string      `json:"browser"        orm:"browser"          description:"浏览器"`
	Os             string      `json:"os"             orm:"os"               description:"操作系统"`
	LoginTime      *gtime.Time `json:"loginTime"      orm:"login_time"       description:"登录时间"`
	LastActiveTime *gtime.Time `json:"lastActiveTime" orm:"last_active_time" description:"最后活跃时间"`
}

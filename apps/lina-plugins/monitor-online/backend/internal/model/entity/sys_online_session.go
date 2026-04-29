// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysOnlineSession is the golang structure for table sys_online_session.
type SysOnlineSession struct {
	TokenId        string      `json:"tokenId"        orm:"token_id"         description:"Session token ID (UUID)"`
	UserId         int         `json:"userId"         orm:"user_id"          description:"User ID"`
	Username       string      `json:"username"       orm:"username"         description:"Login account"`
	DeptName       string      `json:"deptName"       orm:"dept_name"        description:"Department name"`
	Ip             string      `json:"ip"             orm:"ip"               description:"Login IP"`
	Browser        string      `json:"browser"        orm:"browser"          description:"Browser"`
	Os             string      `json:"os"             orm:"os"               description:"Operating system"`
	LoginTime      *gtime.Time `json:"loginTime"      orm:"login_time"       description:"Login time"`
	LastActiveTime *gtime.Time `json:"lastActiveTime" orm:"last_active_time" description:"Last active time"`
}

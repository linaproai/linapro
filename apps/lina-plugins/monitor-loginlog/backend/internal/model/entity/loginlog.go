// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Loginlog is the golang structure for table loginlog.
type Loginlog struct {
	Id        int         `json:"id"        orm:"id"         description:"Log ID"`
	UserName  string      `json:"userName"  orm:"user_name"  description:"Login account"`
	Status    int         `json:"status"    orm:"status"     description:"Login status: 0=succeeded, 1=failed"`
	Ip        string      `json:"ip"        orm:"ip"         description:"Login IP address"`
	Browser   string      `json:"browser"   orm:"browser"    description:"Browser type"`
	Os        string      `json:"os"        orm:"os"         description:"Operating system"`
	Msg       string      `json:"msg"       orm:"msg"        description:"Prompt message"`
	LoginTime *gtime.Time `json:"loginTime" orm:"login_time" description:"Login time"`
}

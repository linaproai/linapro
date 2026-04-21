package v1

import "github.com/gogf/gf/v2/os/gtime"

// LoginLogEntity represents one login-log record returned by plugin APIs.
type LoginLogEntity struct {
	Id        int         `json:"id" dc:"日志ID" eg:"1"`
	UserName  string      `json:"userName" dc:"登录账号" eg:"admin"`
	Status    int         `json:"status" dc:"登录状态：0=成功 1=失败" eg:"0"`
	Ip        string      `json:"ip" dc:"登录IP地址" eg:"127.0.0.1"`
	Browser   string      `json:"browser" dc:"浏览器类型" eg:"Chrome 120.0"`
	Os        string      `json:"os" dc:"操作系统" eg:"macOS"`
	Msg       string      `json:"msg" dc:"提示消息" eg:"登录成功"`
	LoginTime *gtime.Time `json:"loginTime" dc:"登录时间" eg:"2025-01-01 12:00:00"`
}

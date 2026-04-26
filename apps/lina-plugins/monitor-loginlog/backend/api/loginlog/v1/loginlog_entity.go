package v1

import "github.com/gogf/gf/v2/os/gtime"

// LoginLogEntity represents one login-log record returned by plugin APIs.
type LoginLogEntity struct {
	Id        int         `json:"id" dc:"Log ID" eg:"1"`
	UserName  string      `json:"userName" dc:"Login account" eg:"admin"`
	Status    int         `json:"status" dc:"Login status: 0=success 1=failed" eg:"0"`
	Ip        string      `json:"ip" dc:"Login IP address" eg:"127.0.0.1"`
	Browser   string      `json:"browser" dc:"Browser type" eg:"Chrome 120.0"`
	Os        string      `json:"os" dc:"Operating system" eg:"macOS"`
	Msg       string      `json:"msg" dc:"Message" eg:"Login succeeded"`
	LoginTime *gtime.Time `json:"loginTime" dc:"Login time" eg:"2025-01-01 12:00:00"`
}

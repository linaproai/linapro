package v1

import "github.com/gogf/gf/v2/frame/g"

// Online User List API

// OnlineListReq defines the request for listing online users.
type OnlineListReq struct {
	g.Meta   `path:"/monitor/online/list" method:"get" tags:"系统监控" summary:"在线用户列表" dc:"分页查询当前在线用户会话，支持按用户名和IP地址模糊过滤" permission:"monitor:online:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Username string `json:"username" dc:"按用户名模糊过滤，不传则查询全部" eg:"admin"`
	Ip       string `json:"ip" dc:"按IP地址模糊过滤，不传则查询全部" eg:"127.0.0.1"`
}

// OnlineListRes is the online user list response.
type OnlineListRes struct {
	Items []*OnlineUserItem `json:"items" dc:"在线用户列表" eg:"[]"`
	Total int               `json:"total" dc:"在线用户总数" eg:"5"`
}

// OnlineUserItem represents an online user item.
type OnlineUserItem struct {
	TokenId   string `json:"tokenId" dc:"会话Token ID" eg:"abc123"`
	Username  string `json:"username" dc:"登录账号" eg:"admin"`
	DeptName  string `json:"deptName" dc:"部门名称" eg:"研发部"`
	Ip        string `json:"ip" dc:"登录IP" eg:"127.0.0.1"`
	Browser   string `json:"browser" dc:"浏览器" eg:"Chrome 120.0"`
	Os        string `json:"os" dc:"操作系统" eg:"Windows 10"`
	LoginTime string `json:"loginTime" dc:"登录时间" eg:"2025-01-01 12:00:00"`
}

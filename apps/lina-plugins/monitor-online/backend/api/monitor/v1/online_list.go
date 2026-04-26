package v1

import "github.com/gogf/gf/v2/frame/g"

// Online User List API

// OnlineListReq defines the request for listing online users.
type OnlineListReq struct {
	g.Meta   `path:"/monitor/online/list" method:"get" tags:"System Monitoring" summary:"Online user list" dc:"Query current online user sessions by page, with fuzzy filtering by username and IP address." permission:"monitor:online:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	Username string `json:"username" dc:"Fuzzy filtering by username, queries all records when omitted" eg:"admin"`
	Ip       string `json:"ip" dc:"Fuzzy filtering by IP address, queries all records when omitted" eg:"127.0.0.1"`
}

// OnlineListRes is the online user list response.
type OnlineListRes struct {
	Items []*OnlineUserItem `json:"items" dc:"Online user list" eg:"[]"`
	Total int               `json:"total" dc:"Total number of online users" eg:"5"`
}

// OnlineUserItem represents an online user item.
type OnlineUserItem struct {
	TokenId   string `json:"tokenId" dc:"Session Token ID" eg:"abc123"`
	Username  string `json:"username" dc:"Login account" eg:"admin"`
	DeptName  string `json:"deptName" dc:"Department name" eg:"R&D Department"`
	Ip        string `json:"ip" dc:"Login IP" eg:"127.0.0.1"`
	Browser   string `json:"browser" dc:"Browser" eg:"Chrome 120.0"`
	Os        string `json:"os" dc:"Operating system" eg:"Windows 10"`
	LoginTime string `json:"loginTime" dc:"Login time" eg:"2025-01-01 12:00:00"`
}

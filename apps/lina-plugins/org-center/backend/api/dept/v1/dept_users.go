package v1

import "github.com/gogf/gf/v2/frame/g"

// UsersReq returns users belonging to a dept (for leader selection).
type UsersReq struct {
	g.Meta  `path:"/dept/{id}/users" method:"get" tags:"Department Management" summary:"Get department user list" dc:"Get users under the specified department and its child departments for organization-aware selectors. Department ID 0 returns all users, and keyword search matches username or nickname." permission:"system:dept:query"`
	Id      int    `json:"id" dc:"Department ID, 0 means querying all users, when greater than 0, querying users of this department and all its sub-departments" eg:"100"`
	Keyword string `json:"keyword" dc:"Search keywords and perform fuzzy matching by username or nickname" eg:"Zhang"`
	Limit   int    `json:"limit" d:"10" dc:"The maximum number of returned items, the default is 10, used to limit the amount of data in the drop-down list" eg:"10"`
}

// DeptUser represents a user in a department.
type DeptUser struct {
	Id       int    `json:"id" dc:"User ID" eg:"1"`
	Username string `json:"username" dc:"User login account" eg:"zhangsan"`
	Nickname string `json:"nickname" dc:"User nickname, used for front-end display" eg:"Zhang San"`
}

// UsersRes defines the response for querying department users.
type UsersRes struct {
	List []*DeptUser `json:"list" dc:"Department user list" eg:"[]"`
}

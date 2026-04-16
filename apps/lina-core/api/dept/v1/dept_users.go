package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Dept Users API

// UsersReq returns users belonging to a dept (for leader selection).
// When Id=0, returns all users. When Id>0, returns users in the dept and all its sub-depts.
type UsersReq struct {
	g.Meta  `path:"/dept/{id}/users" method:"get" tags:"部门管理" summary:"获取部门用户列表" dc:"获取指定部门及其所有子部门下的用户列表，供管理工作台负责人选择器等按组织结构选人场景复用。当部门ID为0时返回所有用户，支持按用户名或昵称进行关键词搜索" permission:"system:dept:query"`
	Id      int    `json:"id" dc:"部门ID，0表示查询所有用户，大于0时查询该部门及其所有子部门的用户" eg:"100"`
	Keyword string `json:"keyword" dc:"搜索关键词，按用户名或昵称进行模糊匹配" eg:"张"`
	Limit   int    `json:"limit" d:"10" dc:"最大返回条数，默认为10，用于限制下拉列表的数据量" eg:"10"`
}

// DeptUser represents a user in a department.
type DeptUser struct {
	Id       int    `json:"id" dc:"用户ID" eg:"1"`
	Username string `json:"username" dc:"用户登录账号" eg:"zhangsan"`
	Nickname string `json:"nickname" dc:"用户昵称，用于前端显示" eg:"张三"`
}

// UsersRes defines the response for querying department users.
type UsersRes struct {
	List []*DeptUser `json:"list" dc:"部门用户列表" eg:"[]"`
}

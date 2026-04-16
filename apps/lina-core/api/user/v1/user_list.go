package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying the user list.
type ListReq struct {
	g.Meta         `path:"/user" method:"get" tags:"用户管理" summary:"获取用户列表" dc:"分页查询用户列表，支持按用户名、昵称、状态、手机号、性别、部门、创建时间等条件筛选，支持自定义排序" permission:"system:user:query"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Username       string `json:"username" dc:"按用户名筛选（模糊匹配）" eg:"admin"`
	Nickname       string `json:"nickname" dc:"按昵称筛选（模糊匹配）" eg:"管理员"`
	Status         *int   `json:"status" dc:"按状态筛选：1=正常 0=停用" eg:"1"`
	Phone          string `json:"phone" dc:"按手机号筛选（模糊匹配）" eg:"138"`
	Sex            *int   `json:"sex" dc:"按性别筛选：0=未知 1=男 2=女" eg:"1"`
	DeptId         *int   `json:"deptId" dc:"按部门ID筛选（包含子部门）" eg:"100"`
	BeginTime      string `json:"beginTime" dc:"按创建时间起始筛选" eg:"2025-01-01"`
	EndTime        string `json:"endTime" dc:"按创建时间结束筛选" eg:"2025-12-31"`
	OrderBy        string `json:"orderBy" dc:"排序字段：id,username,nickname,phone,email,status,created_at" eg:"id"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向：asc或desc" eg:"desc"`
}

// ListItem represents a single user in the user list.
type ListItem struct {
	*entity.SysUser
	DeptId    int      `json:"deptId" dc:"部门ID" eg:"100"`
	DeptName  string   `json:"deptName" dc:"部门名称" eg:"技术部"`
	RoleIds   []int    `json:"roleIds" dc:"角色ID列表" eg:"[1,2]"`
	RoleNames []string `json:"roleNames" dc:"角色名称列表" eg:"[\"管理员\",\"普通用户\"]"`
}

// ListRes is the response structure for user list query.
type ListRes struct {
	List  []*ListItem `json:"list" dc:"用户列表" eg:"[]"`
	Total int         `json:"total" dc:"总条数" eg:"100"`
}

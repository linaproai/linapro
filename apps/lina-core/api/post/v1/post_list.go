package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying the post list.
type ListReq struct {
	g.Meta   `path:"/post" method:"get" tags:"岗位管理" summary:"获取岗位列表" dc:"分页查询岗位列表，支持按部门、编码、名称、状态等条件筛选" permission:"system:post:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	DeptId   *int   `json:"deptId" dc:"按部门ID筛选" eg:"100"`
	Code     string `json:"code" dc:"按岗位编码筛选（模糊匹配）" eg:"ceo"`
	Name     string `json:"name" dc:"按岗位名称筛选（模糊匹配）" eg:"总经理"`
	Status   *int   `json:"status" dc:"按状态筛选：1=正常 0=停用" eg:"1"`
}

// ListRes is the response for post list
type ListRes struct {
	List  []*entity.SysPost `json:"list" dc:"岗位列表" eg:"[]"`
	Total int               `json:"total" dc:"总条数" eg:"20"`
}

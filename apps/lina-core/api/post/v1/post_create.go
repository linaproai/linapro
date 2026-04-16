package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CreateReq defines the request for creating a post.
type CreateReq struct {
	g.Meta `path:"/post" method:"post" tags:"岗位管理" summary:"创建岗位" dc:"在指定部门下创建一个新岗位，岗位编码在系统中必须唯一" permission:"system:post:add"`
	DeptId int    `json:"deptId" v:"required#请选择所属部门" dc:"所属部门ID" eg:"100"`
	Code   string `json:"code" v:"required#请输入岗位编码" dc:"岗位编码（唯一）" eg:"dev"`
	Name   string `json:"name" v:"required#请输入岗位名称" dc:"岗位名称" eg:"开发工程师"`
	Sort   *int   `json:"sort" d:"0" dc:"排序号，数值越小排序越靠前" eg:"1"`
	Status *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用" eg:"1"`
	Remark string `json:"remark" dc:"备注" eg:"负责系统开发"`
}

// CreateRes defines the response for creating a post.
type CreateRes struct {
	Id int `json:"id" dc:"岗位ID" eg:"1"`
}

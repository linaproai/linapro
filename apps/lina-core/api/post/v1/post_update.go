package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UpdateReq defines the request for updating a post.
type UpdateReq struct {
	g.Meta `path:"/post/{id}" method:"put" tags:"岗位管理" summary:"更新岗位" dc:"更新指定岗位的信息，所有字段均为可选更新" permission:"system:post:edit"`
	Id     int     `json:"id" v:"required" dc:"岗位ID" eg:"1"`
	DeptId *int    `json:"deptId" dc:"所属部门ID" eg:"100"`
	Code   *string `json:"code" dc:"岗位编码（唯一）" eg:"dev"`
	Name   *string `json:"name" dc:"岗位名称" eg:"开发工程师"`
	Sort   *int    `json:"sort" dc:"排序号" eg:"1"`
	Status *int    `json:"status" dc:"状态：1=正常 0=停用" eg:"1"`
	Remark *string `json:"remark" dc:"备注" eg:"负责系统开发"`
}

// UpdateRes defines the response for updating a post.
type UpdateRes struct{}

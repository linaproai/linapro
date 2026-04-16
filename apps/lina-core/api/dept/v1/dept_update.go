package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UpdateReq defines the request for updating a department.
type UpdateReq struct {
	g.Meta   `path:"/dept/{id}" method:"put" tags:"部门管理" summary:"更新部门" dc:"更新指定部门的信息，仅传入需要修改的字段即可，未传入的字段保持不变。不允许将部门的父级设为自身或其子部门" permission:"system:dept:edit"`
	Id       int     `json:"id" v:"required" dc:"部门ID" eg:"100"`
	ParentId *int    `json:"parentId" dc:"父级部门ID，0表示顶级部门，不能设置为自身或其下级部门" eg:"0"`
	Name     *string `json:"name" dc:"部门名称，同一父级下不可重复" eg:"研发中心"`
	Code     *string `json:"code" dc:"部门编码，系统内唯一标识" eg:"RD"`
	OrderNum *int    `json:"orderNum" dc:"排序号，数值越小越靠前" eg:"2"`
	Leader   *int    `json:"leader" dc:"负责人用户ID，关联系统用户表" eg:"1"`
	Phone    *string `json:"phone" dc:"部门联系电话" eg:"021-66666666"`
	Email    *string `json:"email" dc:"部门联系邮箱" eg:"rd@company.com"`
	Status   *int    `json:"status" dc:"部门状态：1=正常，0=停用" eg:"1"`
	Remark   *string `json:"remark" dc:"备注信息" eg:"负责公司核心产品研发"`
}

// UpdateRes defines the response for updating a department.
type UpdateRes struct{}

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CreateReq defines the request for creating a department.
type CreateReq struct {
	g.Meta   `path:"/dept" method:"post" tags:"部门管理" summary:"创建部门" dc:"创建一个新部门，支持设置父级部门形成树形层级结构，部门编码在系统内须唯一" permission:"system:dept:add"`
	ParentId int    `json:"parentId" d:"0" dc:"父级部门ID，0表示顶级部门" eg:"100"`
	Name     string `json:"name" v:"required#请输入部门名称" dc:"部门名称，同一父级下不可重复" eg:"技术部"`
	Code     string `json:"code" dc:"部门编码，系统内唯一标识，用于与外部系统对接" eg:"TECH"`
	OrderNum *int   `json:"orderNum" d:"0" dc:"排序号，数值越小越靠前，同级部门按此字段升序排列" eg:"1"`
	Leader   *int   `json:"leader" dc:"负责人用户ID，关联系统用户表" eg:"1"`
	Phone    string `json:"phone" dc:"部门联系电话" eg:"021-88888888"`
	Email    string `json:"email" dc:"部门联系邮箱" eg:"tech@company.com"`
	Status   *int   `json:"status" d:"1" dc:"部门状态：1=正常，0=停用。停用后该部门及其子部门的用户将无法登录" eg:"1"`
	Remark   string `json:"remark" dc:"备注信息" eg:"负责公司技术研发工作"`
}

// CreateRes defines the response for creating a department.
type CreateRes struct {
	Id int `json:"id" dc:"新创建的部门ID" eg:"110"`
}

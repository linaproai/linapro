// This file declares department DTO entities exposed by the org-center API.

package v1

import "github.com/gogf/gf/v2/os/gtime"

// DeptEntity mirrors the plugin_org_center_dept table shape returned through plugin APIs.
type DeptEntity struct {
	Id        int         `json:"id" dc:"部门ID" eg:"100"`
	ParentId  int         `json:"parentId" dc:"父级部门ID，0表示顶级部门" eg:"0"`
	Ancestors string      `json:"ancestors" dc:"祖级路径，逗号分隔的部门ID链路" eg:"0,100"`
	Name      string      `json:"name" dc:"部门名称" eg:"技术部"`
	Code      string      `json:"code" dc:"部门编码" eg:"TECH"`
	OrderNum  int         `json:"orderNum" dc:"排序号" eg:"1"`
	Leader    int         `json:"leader" dc:"负责人用户ID" eg:"1"`
	Phone     string      `json:"phone" dc:"联系电话" eg:"021-88888888"`
	Email     string      `json:"email" dc:"联系邮箱" eg:"tech@company.com"`
	Status    int         `json:"status" dc:"部门状态：1=正常 0=停用" eg:"1"`
	Remark    string      `json:"remark" dc:"备注信息" eg:"负责公司技术研发工作"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"创建时间" eg:"2026-04-21 10:00:00"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"更新时间" eg:"2026-04-21 10:30:00"`
	DeletedAt *gtime.Time `json:"deletedAt" dc:"软删除时间，未删除时为空" eg:"null"`
}

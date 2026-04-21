// This file declares post DTO entities exposed by the org-center API.

package v1

import "github.com/gogf/gf/v2/os/gtime"

// PostEntity mirrors the plugin_org_center_post table shape returned through plugin APIs.
type PostEntity struct {
	Id        int         `json:"id" dc:"岗位ID" eg:"1"`
	DeptId    int         `json:"deptId" dc:"所属部门ID" eg:"100"`
	Code      string      `json:"code" dc:"岗位编码" eg:"dev"`
	Name      string      `json:"name" dc:"岗位名称" eg:"开发工程师"`
	Sort      int         `json:"sort" dc:"排序号" eg:"1"`
	Status    int         `json:"status" dc:"状态：1=正常 0=停用" eg:"1"`
	Remark    string      `json:"remark" dc:"备注" eg:"负责系统开发"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"创建时间" eg:"2026-04-21 10:00:00"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"更新时间" eg:"2026-04-21 10:30:00"`
	DeletedAt *gtime.Time `json:"deletedAt" dc:"软删除时间，未删除时为空" eg:"null"`
}

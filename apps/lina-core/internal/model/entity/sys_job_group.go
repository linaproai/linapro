// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJobGroup is the golang structure for table sys_job_group.
type SysJobGroup struct {
	Id        uint64      `json:"id"        orm:"id"         description:"任务分组ID"`
	Code      string      `json:"code"      orm:"code"       description:"分组编码"`
	Name      string      `json:"name"      orm:"name"       description:"分组名称"`
	Remark    string      `json:"remark"    orm:"remark"     description:"备注"`
	SortOrder int         `json:"sortOrder" orm:"sort_order" description:"显示排序"`
	IsDefault int         `json:"isDefault" orm:"is_default" description:"是否默认分组（1=是 0=否）"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"删除时间"`
}

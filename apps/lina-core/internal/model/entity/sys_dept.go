// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDept is the golang structure for table sys_dept.
type SysDept struct {
	Id        int         `json:"id"        orm:"id"         description:"部门ID"`
	ParentId  int         `json:"parentId"  orm:"parent_id"  description:"父部门ID"`
	Ancestors string      `json:"ancestors" orm:"ancestors"  description:"祖级列表"`
	Name      string      `json:"name"      orm:"name"       description:"部门名称"`
	Code      string      `json:"code"      orm:"code"       description:"部门编码"`
	OrderNum  int         `json:"orderNum"  orm:"order_num"  description:"显示排序"`
	Leader    int         `json:"leader"    orm:"leader"     description:"负责人用户ID"`
	Phone     string      `json:"phone"     orm:"phone"      description:"联系电话"`
	Email     string      `json:"email"     orm:"email"      description:"邮箱"`
	Status    int         `json:"status"    orm:"status"     description:"状态（0停用 1正常）"`
	Remark    string      `json:"remark"    orm:"remark"     description:"备注"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"删除时间"`
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDictType is the golang structure for table sys_dict_type.
type SysDictType struct {
	Id        int         `json:"id"        orm:"id"         description:"字典类型ID"`
	Name      string      `json:"name"      orm:"name"       description:"字典名称"`
	Type      string      `json:"type"      orm:"type"       description:"字典类型"`
	Status    int         `json:"status"    orm:"status"     description:"状态（0停用 1正常）"`
	Remark    string      `json:"remark"    orm:"remark"     description:"备注"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
}

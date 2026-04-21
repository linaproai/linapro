// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDictData is the golang structure for table sys_dict_data.
type SysDictData struct {
	Id        int         `json:"id"        orm:"id"         description:"字典数据ID"`
	DictType  string      `json:"dictType"  orm:"dict_type"  description:"字典类型"`
	Label     string      `json:"label"     orm:"label"      description:"字典标签"`
	Value     string      `json:"value"     orm:"value"      description:"字典键值"`
	Sort      int         `json:"sort"      orm:"sort"       description:"显示排序"`
	TagStyle  string      `json:"tagStyle"  orm:"tag_style"  description:"标签样式（primary/success/danger/warning等）"`
	CssClass  string      `json:"cssClass"  orm:"css_class"  description:"CSS样式类名"`
	Status    int         `json:"status"    orm:"status"     description:"状态（0停用 1正常）"`
	Remark    string      `json:"remark"    orm:"remark"     description:"备注"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
}

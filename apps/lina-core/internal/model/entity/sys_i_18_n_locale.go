// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysI18NLocale is the golang structure for table sys_i18n_locale.
type SysI18NLocale struct {
	Id         uint64      `json:"id"         orm:"id"          description:"语言ID"`
	Locale     string      `json:"locale"     orm:"locale"      description:"语言编码，如 zh-CN、en-US"`
	Name       string      `json:"name"       orm:"name"        description:"语言名称默认展示值"`
	NativeName string      `json:"nativeName" orm:"native_name" description:"语言原生名称"`
	Sort       int         `json:"sort"       orm:"sort"        description:"显示排序"`
	Status     int         `json:"status"     orm:"status"      description:"状态（0=停用 1=启用）"`
	IsDefault  int         `json:"isDefault"  orm:"is_default"  description:"是否默认语言（0=否 1=是）"`
	Remark     string      `json:"remark"     orm:"remark"      description:"备注"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:"创建时间"`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:"更新时间"`
	DeletedAt  *gtime.Time `json:"deletedAt"  orm:"deleted_at"  description:"删除时间"`
}

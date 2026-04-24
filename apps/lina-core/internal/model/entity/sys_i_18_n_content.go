// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysI18NContent is the golang structure for table sys_i18n_content.
type SysI18NContent struct {
	Id           uint64      `json:"id"           orm:"id"            description:"内容ID"`
	BusinessType string      `json:"businessType" orm:"business_type" description:"业务类型"`
	BusinessId   string      `json:"businessId"   orm:"business_id"   description:"业务主键或稳定业务标识"`
	Field        string      `json:"field"        orm:"field"         description:"业务字段名"`
	Locale       string      `json:"locale"       orm:"locale"        description:"语言编码"`
	ContentType  string      `json:"contentType"  orm:"content_type"  description:"内容类型（plain/markdown/html/json）"`
	Content      string      `json:"content"      orm:"content"       description:"多语言内容值"`
	Status       int         `json:"status"       orm:"status"        description:"状态（0=停用 1=启用）"`
	Remark       string      `json:"remark"       orm:"remark"        description:"备注"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"创建时间"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`
	DeletedAt    *gtime.Time `json:"deletedAt"    orm:"deleted_at"    description:"删除时间"`
}

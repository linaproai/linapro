// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysI18NMessage is the golang structure for table sys_i18n_message.
type SysI18NMessage struct {
	Id           uint64      `json:"id"           orm:"id"            description:"消息ID"`
	Locale       string      `json:"locale"       orm:"locale"        description:"语言编码"`
	MessageKey   string      `json:"messageKey"   orm:"message_key"   description:"翻译键"`
	MessageValue string      `json:"messageValue" orm:"message_value" description:"翻译值"`
	ScopeType    string      `json:"scopeType"    orm:"scope_type"    description:"作用域类型（host/project/plugin/business）"`
	ScopeKey     string      `json:"scopeKey"     orm:"scope_key"     description:"作用域标识，如 core、plugin_id、project_code"`
	SourceType   string      `json:"sourceType"   orm:"source_type"   description:"来源类型（manual/import/sync）"`
	Status       int         `json:"status"       orm:"status"        description:"状态（0=停用 1=启用）"`
	Remark       string      `json:"remark"       orm:"remark"        description:"备注"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"创建时间"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`
	DeletedAt    *gtime.Time `json:"deletedAt"    orm:"deleted_at"    description:"删除时间"`
}

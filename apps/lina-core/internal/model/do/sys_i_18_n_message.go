// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysI18NMessage is the golang structure of table sys_i18n_message for DAO operations like Where/Data.
type SysI18NMessage struct {
	g.Meta       `orm:"table:sys_i18n_message, do:true"`
	Id           any         // 消息ID
	Locale       any         // 语言编码
	MessageKey   any         // 翻译键
	MessageValue any         // 翻译值
	ScopeType    any         // 作用域类型（host/project/plugin/business）
	ScopeKey     any         // 作用域标识，如 core、plugin_id、project_code
	SourceType   any         // 来源类型（manual/import/sync）
	Status       any         // 状态（0=停用 1=启用）
	Remark       any         // 备注
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
	DeletedAt    *gtime.Time // 删除时间
}

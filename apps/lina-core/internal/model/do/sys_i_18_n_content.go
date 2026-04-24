// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysI18NContent is the golang structure of table sys_i18n_content for DAO operations like Where/Data.
type SysI18NContent struct {
	g.Meta       `orm:"table:sys_i18n_content, do:true"`
	Id           any         // 内容ID
	BusinessType any         // 业务类型
	BusinessId   any         // 业务主键或稳定业务标识
	Field        any         // 业务字段名
	Locale       any         // 语言编码
	ContentType  any         // 内容类型（plain/markdown/html/json）
	Content      any         // 多语言内容值
	Status       any         // 状态（0=停用 1=启用）
	Remark       any         // 备注
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
	DeletedAt    *gtime.Time // 删除时间
}

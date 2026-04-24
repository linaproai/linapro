// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysI18NLocale is the golang structure of table sys_i18n_locale for DAO operations like Where/Data.
type SysI18NLocale struct {
	g.Meta     `orm:"table:sys_i18n_locale, do:true"`
	Id         any         // 语言ID
	Locale     any         // 语言编码，如 zh-CN、en-US
	Name       any         // 语言名称默认展示值
	NativeName any         // 语言原生名称
	Sort       any         // 显示排序
	Status     any         // 状态（0=停用 1=启用）
	IsDefault  any         // 是否默认语言（0=否 1=是）
	Remark     any         // 备注
	CreatedAt  *gtime.Time // 创建时间
	UpdatedAt  *gtime.Time // 更新时间
	DeletedAt  *gtime.Time // 删除时间
}

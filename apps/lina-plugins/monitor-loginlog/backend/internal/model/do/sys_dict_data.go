// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDictData is the golang structure of table sys_dict_data for DAO operations like Where/Data.
type SysDictData struct {
	g.Meta    `orm:"table:sys_dict_data, do:true"`
	Id        any         // 字典数据ID
	DictType  any         // 字典类型
	Label     any         // 字典标签
	Value     any         // 字典键值
	Sort      any         // 显示排序
	TagStyle  any         // 标签样式（primary/success/danger/warning等）
	CssClass  any         // CSS样式类名
	Status    any         // 状态（0停用 1正常）
	Remark    any         // 备注
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotice is the golang structure of table sys_notice for DAO operations like Where/Data.
type SysNotice struct {
	g.Meta    `orm:"table:sys_notice, do:true"`
	Id        any         // 公告ID
	Title     any         // 公告标题
	Type      any         // 公告类型（1通知 2公告）
	Content   any         // 公告内容
	FileIds   any         // 附件文件ID列表，逗号分隔
	Status    any         // 公告状态（0草稿 1已发布）
	Remark    any         // 备注
	CreatedBy any         // 创建者
	UpdatedBy any         // 更新者
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 删除时间
}

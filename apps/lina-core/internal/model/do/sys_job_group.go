// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJobGroup is the golang structure of table sys_job_group for DAO operations like Where/Data.
type SysJobGroup struct {
	g.Meta    `orm:"table:sys_job_group, do:true"`
	Id        any         // 任务分组ID
	Code      any         // 分组编码
	Name      any         // 分组名称
	Remark    any         // 备注
	SortOrder any         // 显示排序
	IsDefault any         // 是否默认分组（1=是 0=否）
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 删除时间
}

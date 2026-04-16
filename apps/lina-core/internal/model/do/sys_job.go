// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJob is the golang structure of table sys_job for DAO operations like Where/Data.
type SysJob struct {
	g.Meta      `orm:"table:sys_job, do:true"`
	Id          any         // 任务ID
	Name        any         // 任务名称
	Command     any         // 执行指令
	CronExpr    any         // Cron表达式
	Description any         // 任务描述
	Status      any         // 状态：1=启用 0=禁用
	Singleton   any         // 执行模式：1=单例 0=并行
	MaxTimes    any         // 最大执行次数，0表示无限制
	ExecTimes   any         // 已执行次数
	IsSystem    any         // 是否系统任务：1=是 0=否
	CreateBy    any         // 创建者
	CreateTime  *gtime.Time // 创建时间
	UpdateBy    any         // 更新者
	UpdateTime  *gtime.Time // 更新时间
	Remark      any         // 备注
}

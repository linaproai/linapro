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
	g.Meta               `orm:"table:sys_job, do:true"`
	Id                   any         // 任务ID
	GroupId              any         // 所属分组ID
	Name                 any         // 任务名称
	Description          any         // 任务描述
	TaskType             any         // 任务类型（handler/shell）
	HandlerRef           any         // Handler 唯一引用
	Params               any         // Handler 参数JSON
	TimeoutSeconds       any         // 执行超时时间（秒）
	ShellCmd             any         // Shell 脚本内容
	WorkDir              any         // 工作目录
	Env                  any         // 环境变量JSON
	CronExpr             any         // Cron 表达式
	Timezone             any         // 时区标识
	Scope                any         // 调度范围（master_only/all_node）
	Concurrency          any         // 并发策略（singleton/parallel）
	MaxConcurrency       any         // 并发上限
	MaxExecutions        any         // 最大执行次数（0=无限）
	ExecutedCount        any         // 已执行次数
	StopReason           any         // 停止原因
	LogRetentionOverride any         // 日志保留策略覆盖JSON
	Status               any         // 任务状态（enabled/disabled/paused_by_plugin）
	IsBuiltin            any         // 是否内置任务（1=是 0=否）
	SeedVersion          any         // 种子版本号
	CreatedBy            any         // 创建者用户ID
	UpdatedBy            any         // 更新者用户ID
	CreatedAt            *gtime.Time // 创建时间
	UpdatedAt            *gtime.Time // 更新时间
	DeletedAt            *gtime.Time // 删除时间
}

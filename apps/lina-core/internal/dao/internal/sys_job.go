// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysJobDao is the data access object for the table sys_job.
type SysJobDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysJobColumns      // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysJobColumns defines and stores column names for the table sys_job.
type SysJobColumns struct {
	Id                   string // 任务ID
	GroupId              string // 所属分组ID
	Name                 string // 任务名称
	Description          string // 任务描述
	TaskType             string // 任务类型（handler/shell）
	HandlerRef           string // Handler 唯一引用
	Params               string // Handler 参数JSON
	TimeoutSeconds       string // 执行超时时间（秒）
	ShellCmd             string // Shell 脚本内容
	WorkDir              string // 工作目录
	Env                  string // 环境变量JSON
	CronExpr             string // Cron 表达式
	Timezone             string // 时区标识
	Scope                string // 调度范围（master_only/all_node）
	Concurrency          string // 并发策略（singleton/parallel）
	MaxConcurrency       string // 并发上限
	MaxExecutions        string // 最大执行次数（0=无限）
	ExecutedCount        string // 已执行次数
	StopReason           string // 停止原因
	LogRetentionOverride string // 日志保留策略覆盖JSON
	Status               string // 任务状态（enabled/disabled/paused_by_plugin）
	IsBuiltin            string // 是否内置任务（1=是 0=否）
	SeedVersion          string // 种子版本号
	CreatedBy            string // 创建者用户ID
	UpdatedBy            string // 更新者用户ID
	CreatedAt            string // 创建时间
	UpdatedAt            string // 更新时间
	DeletedAt            string // 删除时间
}

// sysJobColumns holds the columns for the table sys_job.
var sysJobColumns = SysJobColumns{
	Id:                   "id",
	GroupId:              "group_id",
	Name:                 "name",
	Description:          "description",
	TaskType:             "task_type",
	HandlerRef:           "handler_ref",
	Params:               "params",
	TimeoutSeconds:       "timeout_seconds",
	ShellCmd:             "shell_cmd",
	WorkDir:              "work_dir",
	Env:                  "env",
	CronExpr:             "cron_expr",
	Timezone:             "timezone",
	Scope:                "scope",
	Concurrency:          "concurrency",
	MaxConcurrency:       "max_concurrency",
	MaxExecutions:        "max_executions",
	ExecutedCount:        "executed_count",
	StopReason:           "stop_reason",
	LogRetentionOverride: "log_retention_override",
	Status:               "status",
	IsBuiltin:            "is_builtin",
	SeedVersion:          "seed_version",
	CreatedBy:            "created_by",
	UpdatedBy:            "updated_by",
	CreatedAt:            "created_at",
	UpdatedAt:            "updated_at",
	DeletedAt:            "deleted_at",
}

// NewSysJobDao creates and returns a new DAO object for table data access.
func NewSysJobDao(handlers ...gdb.ModelHandler) *SysJobDao {
	return &SysJobDao{
		group:    "default",
		table:    "sys_job",
		columns:  sysJobColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysJobDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysJobDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysJobDao) Columns() SysJobColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysJobDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysJobDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *SysJobDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

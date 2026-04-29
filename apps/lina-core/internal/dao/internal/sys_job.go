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
	Id                   string // Job ID
	GroupId              string // Owning group ID
	Name                 string // Job name
	Description          string // Job description
	TaskType             string // Job type: handler/shell
	HandlerRef           string // Unique handler reference
	Params               string // Handler parameters JSON
	TimeoutSeconds       string // Execution timeout in seconds
	ShellCmd             string // Shell script content
	WorkDir              string // Working directory
	Env                  string // Environment variables JSON
	CronExpr             string // Cron expression
	Timezone             string // Timezone identifier
	Scope                string // Scheduling scope: master_only/all_node
	Concurrency          string // Concurrency policy: singleton/parallel
	MaxConcurrency       string // Maximum concurrency
	MaxExecutions        string // Maximum executions, 0 means unlimited
	ExecutedCount        string // Executed count
	StopReason           string // Stop reason
	LogRetentionOverride string // Log retention override JSON
	Status               string // Job status: enabled/disabled/paused_by_plugin
	IsBuiltin            string // Built-in job flag: 1=yes, 0=no
	SeedVersion          string // Seed version number
	CreatedBy            string // Creator user ID
	UpdatedBy            string // Updater user ID
	CreatedAt            string // Creation time
	UpdatedAt            string // Update time
	DeletedAt            string // Deletion time
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

// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysJobLogDao is the data access object for the table sys_job_log.
type SysJobLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysJobLogColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysJobLogColumns defines and stores column names for the table sys_job_log.
type SysJobLogColumns struct {
	Id             string // 日志ID
	JobId          string // 所属任务ID
	JobSnapshot    string // 执行时任务快照JSON
	NodeId         string // 执行节点标识
	Trigger        string // 触发方式（cron/manual）
	ParamsSnapshot string // 执行时参数快照JSON
	StartAt        string // 开始时间
	EndAt          string // 结束时间
	DurationMs     string // 执行耗时（毫秒）
	Status         string // 执行状态
	ErrMsg         string // 错误摘要
	ResultJson     string // 执行结果JSON
	CreatedAt      string // 创建时间
}

// sysJobLogColumns holds the columns for the table sys_job_log.
var sysJobLogColumns = SysJobLogColumns{
	Id:             "id",
	JobId:          "job_id",
	JobSnapshot:    "job_snapshot",
	NodeId:         "node_id",
	Trigger:        "trigger",
	ParamsSnapshot: "params_snapshot",
	StartAt:        "start_at",
	EndAt:          "end_at",
	DurationMs:     "duration_ms",
	Status:         "status",
	ErrMsg:         "err_msg",
	ResultJson:     "result_json",
	CreatedAt:      "created_at",
}

// NewSysJobLogDao creates and returns a new DAO object for table data access.
func NewSysJobLogDao(handlers ...gdb.ModelHandler) *SysJobLogDao {
	return &SysJobLogDao{
		group:    "default",
		table:    "sys_job_log",
		columns:  sysJobLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysJobLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysJobLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysJobLogDao) Columns() SysJobLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysJobLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysJobLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysJobLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

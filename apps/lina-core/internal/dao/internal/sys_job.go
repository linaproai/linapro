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
	Id          string // 任务ID
	Name        string // 任务名称
	Command     string // 执行指令
	CronExpr    string // Cron表达式
	Description string // 任务描述
	Status      string // 状态：1=启用 0=禁用
	Singleton   string // 执行模式：1=单例 0=并行
	MaxTimes    string // 最大执行次数，0表示无限制
	ExecTimes   string // 已执行次数
	IsSystem    string // 是否系统任务：1=是 0=否
	CreateBy    string // 创建者
	CreateTime  string // 创建时间
	UpdateBy    string // 更新者
	UpdateTime  string // 更新时间
	Remark      string // 备注
}

// sysJobColumns holds the columns for the table sys_job.
var sysJobColumns = SysJobColumns{
	Id:          "id",
	Name:        "name",
	Command:     "command",
	CronExpr:    "cron_expr",
	Description: "description",
	Status:      "status",
	Singleton:   "singleton",
	MaxTimes:    "max_times",
	ExecTimes:   "exec_times",
	IsSystem:    "is_system",
	CreateBy:    "create_by",
	CreateTime:  "create_time",
	UpdateBy:    "update_by",
	UpdateTime:  "update_time",
	Remark:      "remark",
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

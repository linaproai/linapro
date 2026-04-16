// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysLoginLogDao is the data access object for the table sys_login_log.
type SysLoginLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysLoginLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysLoginLogColumns defines and stores column names for the table sys_login_log.
type SysLoginLogColumns struct {
	Id        string // 日志ID
	UserName  string // 登录账号
	Status    string // 登录状态（0成功 1失败）
	Ip        string // 登录IP地址
	Browser   string // 浏览器类型
	Os        string // 操作系统
	Msg       string // 提示消息
	LoginTime string // 登录时间
}

// sysLoginLogColumns holds the columns for the table sys_login_log.
var sysLoginLogColumns = SysLoginLogColumns{
	Id:        "id",
	UserName:  "user_name",
	Status:    "status",
	Ip:        "ip",
	Browser:   "browser",
	Os:        "os",
	Msg:       "msg",
	LoginTime: "login_time",
}

// NewSysLoginLogDao creates and returns a new DAO object for table data access.
func NewSysLoginLogDao(handlers ...gdb.ModelHandler) *SysLoginLogDao {
	return &SysLoginLogDao{
		group:    "default",
		table:    "sys_login_log",
		columns:  sysLoginLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysLoginLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysLoginLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysLoginLogDao) Columns() SysLoginLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysLoginLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysLoginLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysLoginLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

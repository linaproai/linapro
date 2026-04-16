// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysOnlineSessionDao is the data access object for the table sys_online_session.
type SysOnlineSessionDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  SysOnlineSessionColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// SysOnlineSessionColumns defines and stores column names for the table sys_online_session.
type SysOnlineSessionColumns struct {
	TokenId        string // 会话Token ID（UUID）
	UserId         string // 用户ID
	Username       string // 登录账号
	DeptName       string // 部门名称
	Ip             string // 登录IP
	Browser        string // 浏览器
	Os             string // 操作系统
	LoginTime      string // 登录时间
	LastActiveTime string // 最后活跃时间
}

// sysOnlineSessionColumns holds the columns for the table sys_online_session.
var sysOnlineSessionColumns = SysOnlineSessionColumns{
	TokenId:        "token_id",
	UserId:         "user_id",
	Username:       "username",
	DeptName:       "dept_name",
	Ip:             "ip",
	Browser:        "browser",
	Os:             "os",
	LoginTime:      "login_time",
	LastActiveTime: "last_active_time",
}

// NewSysOnlineSessionDao creates and returns a new DAO object for table data access.
func NewSysOnlineSessionDao(handlers ...gdb.ModelHandler) *SysOnlineSessionDao {
	return &SysOnlineSessionDao{
		group:    "default",
		table:    "sys_online_session",
		columns:  sysOnlineSessionColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysOnlineSessionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysOnlineSessionDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysOnlineSessionDao) Columns() SysOnlineSessionColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysOnlineSessionDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysOnlineSessionDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysOnlineSessionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

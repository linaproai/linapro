// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysUserDao is the data access object for the table sys_user.
type SysUserDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysUserColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysUserColumns defines and stores column names for the table sys_user.
type SysUserColumns struct {
	Id        string // 用户ID
	Username  string // 用户账号
	Password  string // 密码
	Nickname  string // 用户昵称
	Email     string // 邮箱
	Phone     string // 手机号码
	Sex       string // 性别（0未知 1男 2女）
	Avatar    string // 头像地址
	Status    string // 状态（0停用 1正常）
	Remark    string // 备注
	LoginDate string // 最后登录时间
	CreatedAt string // 创建时间
	UpdatedAt string // 更新时间
	DeletedAt string // 删除时间
}

// sysUserColumns holds the columns for the table sys_user.
var sysUserColumns = SysUserColumns{
	Id:        "id",
	Username:  "username",
	Password:  "password",
	Nickname:  "nickname",
	Email:     "email",
	Phone:     "phone",
	Sex:       "sex",
	Avatar:    "avatar",
	Status:    "status",
	Remark:    "remark",
	LoginDate: "login_date",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewSysUserDao creates and returns a new DAO object for table data access.
func NewSysUserDao(handlers ...gdb.ModelHandler) *SysUserDao {
	return &SysUserDao{
		group:    "default",
		table:    "sys_user",
		columns:  sysUserColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysUserDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysUserDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysUserDao) Columns() SysUserColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysUserDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysUserDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysUserDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

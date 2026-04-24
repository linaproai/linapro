// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysI18NContentDao is the data access object for the table sys_i18n_content.
type SysI18NContentDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  SysI18NContentColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// SysI18NContentColumns defines and stores column names for the table sys_i18n_content.
type SysI18NContentColumns struct {
	Id           string // 内容ID
	BusinessType string // 业务类型
	BusinessId   string // 业务主键或稳定业务标识
	Field        string // 业务字段名
	Locale       string // 语言编码
	ContentType  string // 内容类型（plain/markdown/html/json）
	Content      string // 多语言内容值
	Status       string // 状态（0=停用 1=启用）
	Remark       string // 备注
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
	DeletedAt    string // 删除时间
}

// sysI18NContentColumns holds the columns for the table sys_i18n_content.
var sysI18NContentColumns = SysI18NContentColumns{
	Id:           "id",
	BusinessType: "business_type",
	BusinessId:   "business_id",
	Field:        "field",
	Locale:       "locale",
	ContentType:  "content_type",
	Content:      "content",
	Status:       "status",
	Remark:       "remark",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewSysI18NContentDao creates and returns a new DAO object for table data access.
func NewSysI18NContentDao(handlers ...gdb.ModelHandler) *SysI18NContentDao {
	return &SysI18NContentDao{
		group:    "default",
		table:    "sys_i18n_content",
		columns:  sysI18NContentColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysI18NContentDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysI18NContentDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysI18NContentDao) Columns() SysI18NContentColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysI18NContentDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysI18NContentDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysI18NContentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

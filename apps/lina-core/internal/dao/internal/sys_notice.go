// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysNoticeDao is the data access object for the table sys_notice.
type SysNoticeDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysNoticeColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysNoticeColumns defines and stores column names for the table sys_notice.
type SysNoticeColumns struct {
	Id        string // 公告ID
	Title     string // 公告标题
	Type      string // 公告类型（1通知 2公告）
	Content   string // 公告内容
	FileIds   string // 附件文件ID列表，逗号分隔
	Status    string // 公告状态（0草稿 1已发布）
	Remark    string // 备注
	CreatedBy string // 创建者
	UpdatedBy string // 更新者
	CreatedAt string // 创建时间
	UpdatedAt string // 更新时间
	DeletedAt string // 删除时间
}

// sysNoticeColumns holds the columns for the table sys_notice.
var sysNoticeColumns = SysNoticeColumns{
	Id:        "id",
	Title:     "title",
	Type:      "type",
	Content:   "content",
	FileIds:   "file_ids",
	Status:    "status",
	Remark:    "remark",
	CreatedBy: "created_by",
	UpdatedBy: "updated_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewSysNoticeDao creates and returns a new DAO object for table data access.
func NewSysNoticeDao(handlers ...gdb.ModelHandler) *SysNoticeDao {
	return &SysNoticeDao{
		group:    "default",
		table:    "sys_notice",
		columns:  sysNoticeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysNoticeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysNoticeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysNoticeDao) Columns() SysNoticeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysNoticeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysNoticeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysNoticeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

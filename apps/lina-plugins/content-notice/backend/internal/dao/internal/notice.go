// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// NoticeDao is the data access object for the table plugin_content_notice.
type NoticeDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  NoticeColumns      // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// NoticeColumns defines and stores column names for the table plugin_content_notice.
type NoticeColumns struct {
	Id        string // Notice ID
	TenantId  string // Owning tenant ID, 0 means PLATFORM
	Title     string // Notice title
	Type      string // Notice type: 1=notification, 2=announcement
	Content   string // Notice content
	FileIds   string // Attachment file ID list, comma-separated
	Status    string // Notice status: 0=draft, 1=published
	Remark    string // Remark
	CreatedBy string // Creator
	UpdatedBy string // Updater
	CreatedAt string // Creation time
	UpdatedAt string // Update time
	DeletedAt string // Deletion time
}

// noticeColumns holds the columns for the table plugin_content_notice.
var noticeColumns = NoticeColumns{
	Id:        "id",
	TenantId:  "tenant_id",
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

// NewNoticeDao creates and returns a new DAO object for table data access.
func NewNoticeDao(handlers ...gdb.ModelHandler) *NoticeDao {
	return &NoticeDao{
		group:    "default",
		table:    "plugin_content_notice",
		columns:  noticeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *NoticeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *NoticeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *NoticeDao) Columns() NoticeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *NoticeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *NoticeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *NoticeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

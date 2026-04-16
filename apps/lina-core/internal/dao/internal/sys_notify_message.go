// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysNotifyMessageDao is the data access object for the table sys_notify_message.
type SysNotifyMessageDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  SysNotifyMessageColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// SysNotifyMessageColumns defines and stores column names for the table sys_notify_message.
type SysNotifyMessageColumns struct {
	Id           string // 主键ID
	PluginId     string // 来源插件ID，宿主内建流程为空
	SourceType   string // 来源类型：notice=公告 plugin=插件 system=系统
	SourceId     string // 来源业务ID
	CategoryCode string // 消息分类：notice=通知 announcement=公告 other=其他
	Title        string // 消息标题
	Content      string // 消息正文
	PayloadJson  string // 扩展载荷JSON
	SenderUserId string // 发送者用户ID
	CreatedAt    string // 创建时间
}

// sysNotifyMessageColumns holds the columns for the table sys_notify_message.
var sysNotifyMessageColumns = SysNotifyMessageColumns{
	Id:           "id",
	PluginId:     "plugin_id",
	SourceType:   "source_type",
	SourceId:     "source_id",
	CategoryCode: "category_code",
	Title:        "title",
	Content:      "content",
	PayloadJson:  "payload_json",
	SenderUserId: "sender_user_id",
	CreatedAt:    "created_at",
}

// NewSysNotifyMessageDao creates and returns a new DAO object for table data access.
func NewSysNotifyMessageDao(handlers ...gdb.ModelHandler) *SysNotifyMessageDao {
	return &SysNotifyMessageDao{
		group:    "default",
		table:    "sys_notify_message",
		columns:  sysNotifyMessageColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysNotifyMessageDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysNotifyMessageDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysNotifyMessageDao) Columns() SysNotifyMessageColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysNotifyMessageDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysNotifyMessageDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysNotifyMessageDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

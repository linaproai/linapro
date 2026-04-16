// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysNotifyChannelDao is the data access object for the table sys_notify_channel.
type SysNotifyChannelDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  SysNotifyChannelColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// SysNotifyChannelColumns defines and stores column names for the table sys_notify_channel.
type SysNotifyChannelColumns struct {
	Id          string // 主键ID
	ChannelKey  string // 通道标识
	Name        string // 通道名称
	ChannelType string // 通道类型：inbox=站内信 email=邮件 webhook=Webhook
	Status      string // 状态：1=启用 0=停用
	ConfigJson  string // 通道配置JSON
	Remark      string // 备注
	CreatedAt   string // 创建时间
	UpdatedAt   string // 更新时间
	DeletedAt   string // 删除时间
}

// sysNotifyChannelColumns holds the columns for the table sys_notify_channel.
var sysNotifyChannelColumns = SysNotifyChannelColumns{
	Id:          "id",
	ChannelKey:  "channel_key",
	Name:        "name",
	ChannelType: "channel_type",
	Status:      "status",
	ConfigJson:  "config_json",
	Remark:      "remark",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
}

// NewSysNotifyChannelDao creates and returns a new DAO object for table data access.
func NewSysNotifyChannelDao(handlers ...gdb.ModelHandler) *SysNotifyChannelDao {
	return &SysNotifyChannelDao{
		group:    "default",
		table:    "sys_notify_channel",
		columns:  sysNotifyChannelColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysNotifyChannelDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysNotifyChannelDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysNotifyChannelDao) Columns() SysNotifyChannelColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysNotifyChannelDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysNotifyChannelDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysNotifyChannelDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

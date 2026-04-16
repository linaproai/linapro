// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysNotifyDeliveryDao is the data access object for the table sys_notify_delivery.
type SysNotifyDeliveryDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  SysNotifyDeliveryColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// SysNotifyDeliveryColumns defines and stores column names for the table sys_notify_delivery.
type SysNotifyDeliveryColumns struct {
	Id             string // 主键ID
	MessageId      string // 通知消息ID
	ChannelKey     string // 投递通道标识
	ChannelType    string // 投递通道类型
	RecipientType  string // 接收者类型：user=用户 email=邮箱 webhook=Webhook
	RecipientKey   string // 接收者标识，如用户ID邮箱地址或Webhook标识
	UserId         string // 站内信用户ID，非站内信时为0
	DeliveryStatus string // 投递状态：0=待发送 1=成功 2=失败
	IsRead         string // 是否已读：0=未读 1=已读
	ReadAt         string // 已读时间
	ErrorMessage   string // 失败原因
	SentAt         string // 发送完成时间
	CreatedAt      string // 创建时间
	UpdatedAt      string // 更新时间
	DeletedAt      string // 删除时间
}

// sysNotifyDeliveryColumns holds the columns for the table sys_notify_delivery.
var sysNotifyDeliveryColumns = SysNotifyDeliveryColumns{
	Id:             "id",
	MessageId:      "message_id",
	ChannelKey:     "channel_key",
	ChannelType:    "channel_type",
	RecipientType:  "recipient_type",
	RecipientKey:   "recipient_key",
	UserId:         "user_id",
	DeliveryStatus: "delivery_status",
	IsRead:         "is_read",
	ReadAt:         "read_at",
	ErrorMessage:   "error_message",
	SentAt:         "sent_at",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewSysNotifyDeliveryDao creates and returns a new DAO object for table data access.
func NewSysNotifyDeliveryDao(handlers ...gdb.ModelHandler) *SysNotifyDeliveryDao {
	return &SysNotifyDeliveryDao{
		group:    "default",
		table:    "sys_notify_delivery",
		columns:  sysNotifyDeliveryColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysNotifyDeliveryDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysNotifyDeliveryDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysNotifyDeliveryDao) Columns() SysNotifyDeliveryColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysNotifyDeliveryDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysNotifyDeliveryDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysNotifyDeliveryDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MediaNodeDao is the data access object for the table media_node.
type MediaNodeDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MediaNodeColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MediaNodeColumns defines and stores column names for the table media_node.
type MediaNodeColumns struct {
	Id         string // ID（自增，无符号）
	NodeNum    string // 节点编号
	Name       string // 节点名称
	QnUrl      string // 节点网关地址
	BasicUrl   string // 基础平台网关地址
	DnUrl      string // 属地网关地址
	CreatorId  string // 创建人ID
	CreateTime string // 创建时间
	UpdaterId  string // 修改人ID
	UpdateTime string // 修改时间
}

// mediaNodeColumns holds the columns for the table media_node.
var mediaNodeColumns = MediaNodeColumns{
	Id:         "id",
	NodeNum:    "node_num",
	Name:       "name",
	QnUrl:      "qn_url",
	BasicUrl:   "basic_url",
	DnUrl:      "dn_url",
	CreatorId:  "creator_id",
	CreateTime: "create_time",
	UpdaterId:  "updater_id",
	UpdateTime: "update_time",
}

// NewMediaNodeDao creates and returns a new DAO object for table data access.
func NewMediaNodeDao(handlers ...gdb.ModelHandler) *MediaNodeDao {
	return &MediaNodeDao{
		group:    "default",
		table:    "media_node",
		columns:  mediaNodeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MediaNodeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MediaNodeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MediaNodeDao) Columns() MediaNodeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MediaNodeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MediaNodeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MediaNodeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

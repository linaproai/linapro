// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// HgDeviceNodeDao is the data access object for the table hg_device_node.
type HgDeviceNodeDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  HgDeviceNodeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// HgDeviceNodeColumns defines and stores column names for the table hg_device_node.
type HgDeviceNodeColumns struct {
	DeviceId string // 设备国标ID（对应device_code）
	NodeNum  string // 节点编号
}

// hgDeviceNodeColumns holds the columns for the table hg_device_node.
var hgDeviceNodeColumns = HgDeviceNodeColumns{
	DeviceId: "device_id",
	NodeNum:  "node_num",
}

// NewHgDeviceNodeDao creates and returns a new DAO object for table data access.
func NewHgDeviceNodeDao(handlers ...gdb.ModelHandler) *HgDeviceNodeDao {
	return &HgDeviceNodeDao{
		group:    "default",
		table:    "hg_device_node",
		columns:  hgDeviceNodeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *HgDeviceNodeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *HgDeviceNodeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *HgDeviceNodeDao) Columns() HgDeviceNodeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *HgDeviceNodeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *HgDeviceNodeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *HgDeviceNodeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

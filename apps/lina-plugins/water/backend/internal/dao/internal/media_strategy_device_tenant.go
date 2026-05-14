// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MediaStrategyDeviceTenantDao is the data access object for the table media_strategy_device_tenant.
type MediaStrategyDeviceTenantDao struct {
	table    string                           // table is the underlying table name of the DAO.
	group    string                           // group is the database configuration group name of the current DAO.
	columns  MediaStrategyDeviceTenantColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler               // handlers for customized model modification.
}

// MediaStrategyDeviceTenantColumns defines and stores column names for the table media_strategy_device_tenant.
type MediaStrategyDeviceTenantColumns struct {
	TenantId   string // 租户ID
	DeviceId   string // 设备国标ID
	StrategyId string // 策略ID
}

// mediaStrategyDeviceTenantColumns holds the columns for the table media_strategy_device_tenant.
var mediaStrategyDeviceTenantColumns = MediaStrategyDeviceTenantColumns{
	TenantId:   "tenant_id",
	DeviceId:   "device_id",
	StrategyId: "strategy_id",
}

// NewMediaStrategyDeviceTenantDao creates and returns a new DAO object for table data access.
func NewMediaStrategyDeviceTenantDao(handlers ...gdb.ModelHandler) *MediaStrategyDeviceTenantDao {
	return &MediaStrategyDeviceTenantDao{
		group:    "default",
		table:    "media_strategy_device_tenant",
		columns:  mediaStrategyDeviceTenantColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MediaStrategyDeviceTenantDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MediaStrategyDeviceTenantDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MediaStrategyDeviceTenantDao) Columns() MediaStrategyDeviceTenantColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MediaStrategyDeviceTenantDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MediaStrategyDeviceTenantDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MediaStrategyDeviceTenantDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

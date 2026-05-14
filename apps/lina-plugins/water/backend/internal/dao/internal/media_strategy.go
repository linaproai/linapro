// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MediaStrategyDao is the data access object for the table media_strategy.
type MediaStrategyDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  MediaStrategyColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// MediaStrategyColumns defines and stores column names for the table media_strategy.
type MediaStrategyColumns struct {
	Id         string // 策略ID
	Name       string // 策略名称
	Strategy   string // YAML格式策略内容
	Global     string // 是否全局策略：1是，2否
	Enable     string // 启用状态：1开启，2关闭
	CreatorId  string // 创建人ID
	UpdaterId  string // 修改人ID
	CreateTime string // 创建时间
	UpdateTime string // 修改时间
}

// mediaStrategyColumns holds the columns for the table media_strategy.
var mediaStrategyColumns = MediaStrategyColumns{
	Id:         "id",
	Name:       "name",
	Strategy:   "strategy",
	Global:     "global",
	Enable:     "enable",
	CreatorId:  "creator_id",
	UpdaterId:  "updater_id",
	CreateTime: "create_time",
	UpdateTime: "update_time",
}

// NewMediaStrategyDao creates and returns a new DAO object for table data access.
func NewMediaStrategyDao(handlers ...gdb.ModelHandler) *MediaStrategyDao {
	return &MediaStrategyDao{
		group:    "default",
		table:    "media_strategy",
		columns:  mediaStrategyColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MediaStrategyDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MediaStrategyDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MediaStrategyDao) Columns() MediaStrategyColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MediaStrategyDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MediaStrategyDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MediaStrategyDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

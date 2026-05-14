// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// HgTenantStreamConfigDao is the data access object for the table hg_tenant_stream_config.
type HgTenantStreamConfigDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  HgTenantStreamConfigColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// HgTenantStreamConfigColumns defines and stores column names for the table hg_tenant_stream_config.
type HgTenantStreamConfigColumns struct {
	TenantId      string // 租户ID
	MaxConcurrent string // 最大并发数
	NodeNum       string // 节点编号
	Enable        string // 1开启，0关闭
	CreatorId     string // 创建人ID
	CreateTime    string // 创建时间
	UpdaterId     string // 修改人ID
	UpdateTime    string // 修改时间
}

// hgTenantStreamConfigColumns holds the columns for the table hg_tenant_stream_config.
var hgTenantStreamConfigColumns = HgTenantStreamConfigColumns{
	TenantId:      "tenant_id",
	MaxConcurrent: "max_concurrent",
	NodeNum:       "node_num",
	Enable:        "enable",
	CreatorId:     "creator_id",
	CreateTime:    "create_time",
	UpdaterId:     "updater_id",
	UpdateTime:    "update_time",
}

// NewHgTenantStreamConfigDao creates and returns a new DAO object for table data access.
func NewHgTenantStreamConfigDao(handlers ...gdb.ModelHandler) *HgTenantStreamConfigDao {
	return &HgTenantStreamConfigDao{
		group:    "default",
		table:    "hg_tenant_stream_config",
		columns:  hgTenantStreamConfigColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *HgTenantStreamConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *HgTenantStreamConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *HgTenantStreamConfigDao) Columns() HgTenantStreamConfigColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *HgTenantStreamConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *HgTenantStreamConfigDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *HgTenantStreamConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
